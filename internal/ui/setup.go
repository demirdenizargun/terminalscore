package ui

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/demirdenizargun/terminalscore/internal/config"
	"github.com/demirdenizargun/terminalscore/internal/espn"
)

const maxResults = 12

// setupModel, favori takım seçme ekranının durumu.
type setupModel struct {
	query   string
	options []teamOption
	loaded  bool
	cursor  int
}

type teamOption struct {
	team   espn.Team
	league espn.League
	folded string // aramada kullanılan sade ad (küçük harf, Türkçe karaktersiz)
}

type teamsMsg struct{ options []teamOption }

// fetchAllTeams, bütün ulusal liglerin takımlarını aynı anda çeker.
//
// Go'da "go func() {...}()" bir goroutine başlatır: fonksiyon arka planda,
// diğerleriyle eşzamanlı çalışır. 40'tan fazla ligi tek tek beklemek yerine
// hepsini birden isteriz. sync.WaitGroup hepsinin bitmesini beklemeye,
// sync.Mutex de aynı listeye aynı anda yazılmasını engellemeye yarar.
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
					return // bir lig gelmezse diğerleriyle devam
				}
				mu.Lock()
				defer mu.Unlock()
				for _, t := range teams {
					options = append(options, teamOption{team: t, league: league, folded: fold(t.Name)})
				}
			}()
		}
		wg.Wait()

		// Büyük ligin takımları önce; aynı takım iki ligde çıkarsa büyük olanda kalır.
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

// results, arama kutusundaki metne uyan takımlar: adı aramayla
// başlayanlar önce, sonra adında geçenler.
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
	case tea.KeyEsc:
		// İlk açılışta "atla" demek; sonradan açıldıysa favori aynen kalır.
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
	m.selectedID = "" // sıra değişti; imleç en üstten başlasın
	m.sortMatches()
	if _, ok := m.standings[m.standingsSlug()]; !ok {
		return m, m.fetchStandings()
	}
	return m, nil
}

func (m Model) setupView() string {
	s := m.setup
	var b strings.Builder

	b.WriteString(titleStyle.Render("Favori takımın hangisi?") + "\n")
	b.WriteString(mutedStyle.Render("Favori takımının maçları ve ligi listede en üstte görünür.") + "\n\n")
	b.WriteString("  Ara: " + s.query + selectedStyle.Render("█") + "\n\n")

	switch {
	case !s.loaded:
		b.WriteString(mutedStyle.Render("  Takımlar yükleniyor..."))
	case strings.TrimSpace(s.query) == "":
		b.WriteString(mutedStyle.Render("  Takım adını yazmaya başla (ör. galatasaray, besiktas, real madrid)"))
	default:
		res := s.results()
		if len(res) == 0 {
			b.WriteString(mutedStyle.Render("  Eşleşen takım yok."))
		}
		for i, o := range res {
			line := padRight(o.team.Name, 26) + mutedStyle.Render(o.league.Name)
			if i == s.cursor {
				b.WriteString(selectedStyle.Render("▶ ") + selectedStyle.Render(padRight(o.team.Name, 26)) + mutedStyle.Render(o.league.Name))
			} else {
				b.WriteString("  " + line)
			}
			b.WriteString("\n")
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

// fold, aramayı Türkçe karakter ve aksanlardan bağımsız yapar:
// "Beşiktaş", "besiktas" ve "BEŞİKTAŞ" aynı sonucu verir.
func fold(s string) string {
	s = strings.ReplaceAll(s, "İ", "i")
	return foldReplacer.Replace(strings.ToLower(s))
}
