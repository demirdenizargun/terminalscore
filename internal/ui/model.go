// Package ui, Bubble Tea ile terminal arayüzünü çizer.
//
// Bubble Tea "Elm mimarisi" kullanır:
//   - Model: ekranın o anki durumu (hangi gün, hangi maçlar, imleç nerede...)
//   - Update: bir mesaj (tuş, gelen veri, zamanlayıcı) gelince Model'i günceller
//   - View: Model'den ekrana basılacak metni üretir
package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"terminalscore/internal/espn"
)

const refreshEvery = 30 * time.Second

type tab int

const (
	tabMatches tab = iota
	tabStandings
)

type Model struct {
	client *espn.Client
	loc    *time.Location

	tab tab
	day time.Time

	matches       []espn.Match
	matchesLoaded bool
	cursor        int

	standings []espn.Standing

	detailID     string // boş değilse maç detayı ekranı açık
	events       []espn.Event
	eventsLoaded bool
	scroll       int

	err           error
	updated       time.Time
	width, height int
}

func New(client *espn.Client, loc *time.Location) Model {
	now := time.Now().In(loc)
	return Model{
		client: client,
		loc:    loc,
		day:    time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc),
	}
}

// Mesajlar: arka planda yapılan isteklerin sonuçları.
type (
	matchesMsg struct {
		day     time.Time
		matches []espn.Match
		err     error
	}
	standingsMsg struct {
		rows []espn.Standing
		err  error
	}
	eventsMsg struct {
		id     string
		events []espn.Event
		err    error
	}
	tickMsg time.Time
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetchMatches(), m.fetchStandings(), tick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		return m.handleKey(msg)

	case matchesMsg:
		if !msg.day.Equal(m.day) {
			return m, nil // kullanıcı bu arada başka güne geçmiş
		}
		m.err = msg.err
		if msg.err == nil {
			m.matches = msg.matches
			m.matchesLoaded = true
			m.updated = time.Now()
			m.cursor = min(m.cursor, max(len(m.matches)-1, 0))
		}

	case standingsMsg:
		m.err = msg.err
		if msg.err == nil {
			m.standings = msg.rows
		}

	case eventsMsg:
		if msg.id != m.detailID {
			return m, nil
		}
		m.err = msg.err
		if msg.err == nil {
			m.events = msg.events
			m.eventsLoaded = true
		}

	case tickMsg:
		cmds := []tea.Cmd{tick()}
		if m.hasActiveMatch() {
			cmds = append(cmds, m.fetchMatches(), m.fetchStandings())
			if m.detailID != "" {
				cmds = append(cmds, m.fetchEvents())
			}
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "r":
		cmds := []tea.Cmd{m.fetchMatches(), m.fetchStandings()}
		if m.detailID != "" {
			cmds = append(cmds, m.fetchEvents())
		}
		return m, tea.Batch(cmds...)
	}

	if m.detailID != "" {
		switch msg.String() {
		case "esc", "backspace", "left", "h":
			m.detailID = ""
		case "up", "k":
			m.scroll = max(m.scroll-1, 0)
		case "down", "j":
			m.scroll++
		}
		return m, nil
	}

	switch msg.String() {
	case "tab", "shift+tab":
		if m.tab == tabMatches {
			m.tab = tabStandings
		} else {
			m.tab = tabMatches
		}
		return m, nil
	case "1":
		m.tab = tabMatches
		return m, nil
	case "2":
		m.tab = tabStandings
		return m, nil
	}

	if m.tab != tabMatches {
		return m, nil
	}

	switch msg.String() {
	case "left", "h":
		return m.changeDay(m.day.AddDate(0, 0, -1))
	case "right", "l":
		return m.changeDay(m.day.AddDate(0, 0, 1))
	case "t":
		now := time.Now().In(m.loc)
		return m.changeDay(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, m.loc))
	case "up", "k":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j":
		m.cursor = min(m.cursor+1, max(len(m.matches)-1, 0))
	case "enter":
		if len(m.matches) > 0 {
			m.detailID = m.matches[m.cursor].ID
			m.events = nil
			m.eventsLoaded = false
			m.scroll = 0
			return m, m.fetchEvents()
		}
	}
	return m, nil
}

func (m Model) changeDay(day time.Time) (tea.Model, tea.Cmd) {
	m.day = day
	m.matches = nil
	m.matchesLoaded = false
	m.cursor = 0
	m.err = nil
	return m, m.fetchMatches()
}

// hasActiveMatch: canlı maç varsa ya da başlama saati gelmiş ama
// henüz "başladı" görünmeyen maç varsa otomatik yenileme yapılır.
func (m Model) hasActiveMatch() bool {
	now := time.Now()
	for _, mt := range m.matches {
		if mt.State == espn.Live {
			return true
		}
		if mt.State == espn.Scheduled && !mt.Kickoff.IsZero() && now.After(mt.Kickoff) &&
			now.Sub(mt.Kickoff) < 3*time.Hour {
			return true
		}
	}
	return false
}

func (m Model) selectedMatch() (espn.Match, bool) {
	for _, mt := range m.matches {
		if mt.ID == m.detailID {
			return mt, true
		}
	}
	return espn.Match{}, false
}

func tick() tea.Cmd {
	return tea.Tick(refreshEvery, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) fetchMatches() tea.Cmd {
	client, day := m.client, m.day
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		matches, err := client.Matches(ctx, day)
		return matchesMsg{day: day, matches: matches, err: err}
	}
}

func (m Model) fetchStandings() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		rows, err := client.Standings(ctx)
		return standingsMsg{rows: rows, err: err}
	}
}

func (m Model) fetchEvents() tea.Cmd {
	client, id := m.client, m.detailID
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		events, err := client.Events(ctx, id)
		return eventsMsg{id: id, events: events, err: err}
	}
}
