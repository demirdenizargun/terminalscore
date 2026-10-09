package ui

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/demirdenizargun/terminalscore/internal/config"
	"github.com/demirdenizargun/terminalscore/internal/espn"
)

const maxResults = 12

// favori takım seçme ekranı
type setupModel struct {
	query   string
	options []teamOption
	loaded  bool
	cursor  int
}

type teamOption struct {
	team   espn.Team
	league espn.League
	folded string // fold(team.Name)
}

type teamsMsg struct{ options []teamOption }

// Ulusal liglerin takımlarını paralel çeker (~45 istek).
func fetchAllTeams(client *espn.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var (
			wg      sync.WaitGroup
			mu      sync.Mutex
			options []teamOption
		)
		for _, league := range espn.Leagues {
			if league.Kind != espn.Domestic {
				continue
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				teams, err := client.Teams(ctx, league.Slug)
				if err != nil {
					return
				}
				mu.Lock()
				defer mu.Unlock()
				for _, t := range teams {
					options = append(options, teamOption{team: t, league: league, folded: fold(t.Name)})
				}
			}()
		}
		wg.Wait()

		// aynı takım iki ligde varsa büyük olanda kalsın
		sort.SliceStable(options, func(i, j int) bool {
			ri, rj := espn.LeagueRank(options[i].league.Slug), espn.LeagueRank(options[j].league.Slug)
			if ri != rj {
				return ri < rj
			}
			return options[i].team.Name < options[j].team.Name
		})
		seen := map[string]bool{}
		unique := options[:0]
		for _, o := range options {
			if !seen[o.team.ID] {
				seen[o.team.ID] = true
				unique = append(unique, o)
			}
		}
		return teamsMsg{options: unique}
	}
}

// önce adı aramayla başlayanlar, sonra içinde geçenler
func (s setupModel) results() []teamOption {
	q := fold(strings.TrimSpace(s.query))
	if q == "" {
		return nil
	}
	var prefix, contains []teamOption
	for _, o := range s.options {
		switch {
		case strings.HasPrefix(o.folded, q):
			prefix = append(prefix, o)
		case strings.Contains(o.folded, q):
			contains = append(contains, o)
		}
	}
	out := append(prefix, contains...)
	if len(out) > maxResults {
		out = out[:maxResults]
	}
	return out
}

func (m Model) updateSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := m.setup
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc: // atla; favori varsa olduğu gibi kalır
		m.cfg.SetupDone = true
		return m.closeSetup()
	case tea.KeyEnter:
		res := s.results()
		if len(res) == 0 {
			return m, nil
		}
		o := res[min(s.cursor, len(res)-1)]
		m.cfg.SetupDone = true
		m.cfg.FavoriteTeam = &config.FavoriteTeam{ID: o.team.ID, Name: o.team.Name, LeagueSlug: o.league.Slug}
		m.standingsIdx = m.defaultStandingsIdx()
		return m.closeSetup()
	case tea.KeyUp:
		s.cursor = max(s.cursor-1, 0)
	case tea.KeyDown:
		s.cursor = min(s.cursor+1, max(len(s.results())-1, 0))
	case tea.KeyBackspace:
		if r := []rune(s.query); len(r) > 0 {
			s.query = string(r[:len(r)-1])
			s.cursor = 0
		}
	case tea.KeyRunes, tea.KeySpace:
		s.query += string(msg.Runes)
		s.cursor = 0
	}
	return m, nil
}

func (m Model) closeSetup() (tea.Model, tea.Cmd) {
	m.setup = nil
	if err := m.cfg.Save(); err != nil {
		m.err = err
	}
	m.selectedID = "" // sıralama değişti, imleç başa
	m.sortMatches()
	if _, ok := m.standings[m.standingsSlug()]; !ok {
		return m, m.fetchStandings()
	}
	return m, nil
}

func (m Model) setupView() string {
	s := m.setup
	var b strings.Builder

	tx := m.tx
	b.WriteString(titleStyle.Render(tx.setupTitle) + "\n")
	b.WriteString(mutedStyle.Render(tx.setupSubtitle) + "\n\n")
	b.WriteString("  " + tx.setupSearch + s.query + selectedStyle.Render("█") + "\n\n")

	switch {
	case !s.loaded:
		b.WriteString(mutedStyle.Render("  " + tx.setupLoading))
	case len(s.options) == 0:
		b.WriteString(errorStyle.Render("  "+tx.setupNoTeams) + "\n")
		b.WriteString(mutedStyle.Render("  " + tx.setupNoTeamsHint))
	case strings.TrimSpace(s.query) == "":
		b.WriteString(mutedStyle.Render("  " + tx.setupHint))
	default:
		res := s.results()
		if len(res) == 0 {
			b.WriteString(mutedStyle.Render("  " + tx.setupNoMatch))
		}
		for i, o := range res {
			name := padRight(ansi.Truncate(o.team.Name, 25, "…"), 26)
			if i == s.cursor {
				b.WriteString(selectedStyle.Render("▶ " + name))
			} else {
				b.WriteString("  " + name)
			}
			b.WriteString(mutedStyle.Render(tx.league(o.league)) + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

var foldReplacer = strings.NewReplacer(
	"ı", "i", "i̇", "i", "ş", "s", "ğ", "g", "ü", "u", "ö", "o", "ç", "c",
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a", "å", "a", "æ", "ae",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ø", "o",
	"ú", "u", "ù", "u", "û", "u", "ñ", "n", "ß", "ss",
)

// "BEŞİKTAŞ", "Beşiktaş" -> "besiktas"
func fold(s string) string {
	s = strings.ReplaceAll(s, "İ", "i")
	return foldReplacer.Replace(strings.ToLower(s))
}
