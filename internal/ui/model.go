// Package ui, Bubble Tea ile terminal arayüzünü çizer.
//
// Bubble Tea "Elm mimarisi" kullanır:
//   - Model: ekranın o anki durumu (hangi gün, hangi maçlar, imleç nerede...)
//   - Update: bir mesaj (tuş, gelen veri, zamanlayıcı) gelince Model'i günceller
//   - View: Model'den ekrana basılacak metni üretir
package ui

import (
	"context"
	"fmt"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"terminalscore/internal/config"
	"terminalscore/internal/espn"
)

// Yenileme aralıkları ESPN'in önbellek sürelerine göre seçildi: maç listesi
// sunucuda 10 sn, maç detayı 1 sn önbellekte tutuluyor. Daha sık sormak
// aynı veriyi tekrar indirmek olur.
const (
	matchesRefresh   = 10 * time.Second
	eventsRefresh    = 5 * time.Second
	standingsRefresh = time.Minute
	goalFlashFor     = time.Minute // gol olan maç bu süre boyunca "GOL!" ile işaretlenir
)

// standingLeagues, puan durumu sekmesinde ←/→ ile gezilen ligler.
var standingLeagues = func() []espn.League {
	var out []espn.League
	for _, l := range espn.Leagues {
		if l.HasStandings() {
			out = append(out, l)
		}
	}
	return out
}()

type tab int

const (
	tabMatches tab = iota
	tabStandings
)

type Model struct {
	client *espn.Client
	loc    *time.Location
	cfg    config.Config

	tab tab
	day time.Time

	matches       []espn.Match // ekrandaki sırayla: favori önce, sonra lig büyüklüğü
	matchesLoaded bool
	selectedID    string // imleç maç kimliğiyle tutulur; yenilemede sıra değişse de kaybolmaz

	standingsIdx int // standingLeagues içindeki seçili lig
	standings    map[string][]espn.StandingGroup

	detailID     string // boş değilse maç detayı ekranı açık
	events       []espn.Event
	eventsLoaded bool

	scroll int // maç detayı ve puan durumunda kaydırma

	setup *setupModel // nil değilse favori takım seçme ekranı açık

	goals map[string]time.Time // maç kimliği → son golün görüldüğü an

	// son istek zamanları; saniyelik tick bunlara bakıp yenileme kararı verir
	matchesAsked, eventsAsked, standingsAsked time.Time

	err           error
	updated       time.Time
	width, height int
}

func New(client *espn.Client, loc *time.Location, cfg config.Config) Model {
	now := time.Now().In(loc)
	m := Model{
		client:    client,
		loc:       loc,
		cfg:       cfg,
		day:       time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc),
		standings: map[string][]espn.StandingGroup{},
		goals:     map[string]time.Time{},
	}
	m.standingsIdx = m.defaultStandingsIdx()
	if !cfg.SetupDone {
		m.setup = &setupModel{}
	}
	return m
}

// Mesajlar: arka planda yapılan isteklerin sonuçları.
type (
	matchesMsg struct {
		day     time.Time
		matches []espn.Match
		err     error
	}
	standingsMsg struct {
		slug   string
		groups []espn.StandingGroup
		err    error
	}
	eventsMsg struct {
		id     string
		events []espn.Event
		err    error
	}
	tickMsg time.Time
)

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.fetchMatches(), m.fetchStandings(), tick()}
	if m.setup != nil {
		cmds = append(cmds, fetchAllTeams(m.client))
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		if m.setup != nil {
			return m.updateSetup(msg)
		}
		return m.handleKey(msg)

	case teamsMsg:
		if m.setup != nil {
			m.setup.options = msg.options
			m.setup.loaded = true
		}

	case matchesMsg:
		if !msg.day.Equal(m.day) {
			return m, nil // kullanıcı bu arada başka güne geçmiş
		}
		m.err = msg.err
		if msg.err == nil {
			m.detectGoals(msg.matches)
			m.matches = msg.matches
			m.matchesLoaded = true
			m.updated = time.Now()
			m.sortMatches()
		}

	case standingsMsg:
		m.err = msg.err
		if msg.err == nil {
			m.standings[msg.slug] = msg.groups
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
		// Her saniye çalışır: ekran yeniden çizilir ("x sn önce" sayacı, GOL!
		// yanıp sönmesi) ve süresi dolan veriler yeniden istenir.
		now := time.Time(msg)
		cmds := []tea.Cmd{tick()}
		if m.hasActiveMatch() {
			if now.Sub(m.matchesAsked) >= matchesRefresh {
				m.matchesAsked = now
				cmds = append(cmds, m.fetchMatches())
			}
			if m.tab == tabStandings && now.Sub(m.standingsAsked) >= standingsRefresh {
				m.standingsAsked = now
				cmds = append(cmds, m.fetchStandings())
			}
		}
		if mt, ok := m.selectedMatch(); ok && m.detailID != "" && mt.State == espn.Live &&
			now.Sub(m.eventsAsked) >= eventsRefresh {
			m.eventsAsked = now
			cmds = append(cmds, m.fetchEvents())
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// detectGoals, yeni gelen skorları eskileriyle karşılaştırır; skoru artan
// maçları "gol oldu" diye işaretler.
func (m *Model) detectGoals(fresh []espn.Match) {
	old := make(map[string]espn.Match, len(m.matches))
	for _, mt := range m.matches {
		old[mt.ID] = mt
	}
	now := time.Now()
	for _, mt := range fresh {
		prev, ok := old[mt.ID]
		if !ok {
			continue
		}
		if scoreValue(mt.HomeScore) > scoreValue(prev.HomeScore) ||
			scoreValue(mt.AwayScore) > scoreValue(prev.AwayScore) {
			m.goals[mt.ID] = now
		}
	}
}

// recentGoal, maçta son bir dakika içinde gol olup olmadığını söyler.
func (m Model) recentGoal(matchID string) bool {
	t, ok := m.goals[matchID]
	return ok && time.Since(t) < goalFlashFor
}

func scoreValue(s string) int {
	var v int
	if _, err := fmt.Sscan(s, &v); err != nil {
		return -1
	}
	return v
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
	case "f":
		m.setup = &setupModel{}
		return m, fetchAllTeams(m.client)
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
			return m.switchTab(tabStandings)
		}
		return m.switchTab(tabMatches)
	case "1":
		return m.switchTab(tabMatches)
	case "2":
		return m.switchTab(tabStandings)
	}

	if m.tab == tabStandings {
		return m.handleStandingsKey(msg)
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
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)
	case "pgup":
		m.moveCursor(-10)
	case "pgdown":
		m.moveCursor(10)
	case "enter":
		if m.selectedID != "" {
			m.detailID = m.selectedID
			m.events = nil
			m.eventsLoaded = false
			m.scroll = 0
			m.eventsAsked = time.Now()
			return m, m.fetchEvents()
		}
	}
	return m, nil
}

func (m Model) handleStandingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		m.standingsIdx = (m.standingsIdx - 1 + len(standingLeagues)) % len(standingLeagues)
	case "right", "l":
		m.standingsIdx = (m.standingsIdx + 1) % len(standingLeagues)
	case "up", "k":
		m.scroll = max(m.scroll-1, 0)
		return m, nil
	case "down", "j":
		m.scroll++
		return m, nil
	default:
		return m, nil
	}
	// lig değişti
	m.scroll = 0
	m.err = nil
	if _, ok := m.standings[m.standingsSlug()]; ok {
		return m, nil
	}
	return m, m.fetchStandings()
}

func (m Model) switchTab(t tab) (tea.Model, tea.Cmd) {
	m.tab = t
	m.scroll = 0
	if t == tabStandings {
		if _, ok := m.standings[m.standingsSlug()]; !ok {
			return m, m.fetchStandings()
		}
	}
	return m, nil
}

func (m Model) changeDay(day time.Time) (tea.Model, tea.Cmd) {
	m.day = day
	m.matches = nil
	m.matchesLoaded = false
	m.selectedID = ""
	m.err = nil
	return m, m.fetchMatches()
}

func (m *Model) moveCursor(delta int) {
	if len(m.matches) == 0 {
		return
	}
	i := m.selectedIndex() + delta
	i = min(max(i, 0), len(m.matches)-1)
	m.selectedID = m.matches[i].ID
}

func (m Model) selectedIndex() int {
	for i, mt := range m.matches {
		if mt.ID == m.selectedID {
			return i
		}
	}
	return 0
}

// sortMatches, maçları ekrandaki sıraya dizer:
//  1. favori takımın oynadığı lig/turnuva
//  2. favori takımın kendi ligi
//  3. lig büyüklüğü (espn.Leagues sırası), listede olmayanlar en sonda
//
// Aynı ligde favori takımın maçı başa, diğerleri başlama saatine göre.
func (m *Model) sortMatches() {
	favTeam := m.cfg.TeamID()
	favLeagueID := ""
	if l, ok := espn.FindLeague(m.cfg.LeagueSlug()); ok {
		favLeagueID = l.ID
	}

	favPlays := map[string]bool{} // favori takımın bugün oynadığı ligler
	for _, mt := range m.matches {
		if mt.HasTeam(favTeam) {
			favPlays[mt.LeagueID] = true
		}
	}

	priority := func(mt espn.Match) int {
		switch {
		case favPlays[mt.LeagueID]:
			return 0
		case mt.LeagueID == favLeagueID:
			return 1
		}
		return 2
	}

	sort.SliceStable(m.matches, func(i, j int) bool {
		a, b := m.matches[i], m.matches[j]
		if pa, pb := priority(a), priority(b); pa != pb {
			return pa < pb
		}
		if ra, rb := espn.LeagueRank(a.LeagueID), espn.LeagueRank(b.LeagueID); ra != rb {
			return ra < rb
		}
		if a.LeagueName != b.LeagueName {
			return a.LeagueName < b.LeagueName
		}
		if a.LeagueID != b.LeagueID {
			return a.LeagueID < b.LeagueID
		}
		if fa, fb := a.HasTeam(favTeam), b.HasTeam(favTeam); fa != fb {
			return fa
		}
		if !a.Kickoff.Equal(b.Kickoff) {
			return a.Kickoff.Before(b.Kickoff)
		}
		return a.ID < b.ID
	})

	// seçili maç artık yoksa (gün değişti vb.) en üsttekini seç
	found := false
	for _, mt := range m.matches {
		if mt.ID == m.selectedID {
			found = true
			break
		}
	}
	if !found {
		m.selectedID = ""
		if len(m.matches) > 0 {
			m.selectedID = m.matches[0].ID
		}
	}
}

// defaultStandingsIdx: favori takımın ligi, yoksa ilk ulusal lig.
func (m Model) defaultStandingsIdx() int {
	first := -1
	for i, l := range standingLeagues {
		if l.Slug == m.cfg.LeagueSlug() {
			return i
		}
		if first < 0 && l.Kind == espn.Domestic {
			first = i
		}
	}
	return max(first, 0)
}

func (m Model) standingsSlug() string { return standingLeagues[m.standingsIdx].Slug }

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
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) fetchMatches() tea.Cmd {
	client, day := m.client, m.day
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		matches, err := client.Matches(ctx, day)
		return matchesMsg{day: day, matches: matches, err: err}
	}
}

func (m Model) fetchStandings() tea.Cmd {
	client, slug := m.client, m.standingsSlug()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		groups, err := client.Standings(ctx, slug)
		return standingsMsg{slug: slug, groups: groups, err: err}
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
