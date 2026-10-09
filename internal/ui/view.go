package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/demirdenizargun/terminalscore/internal/espn"
)

const (
	statusW = 7
	teamW   = 22
	scoreW  = 7
	sideW   = 30 // detay ekranında ev sahibi / deplasman sütunu
	minuteW = 9

	chromeLines    = 6 // başlık + alt bilgi
	detailFixed    = 4 // kaydırırken yerinde kalan skor bölümü
	standingsFixed = 3
)

func (m Model) View() string {
	var body, help string
	switch {
	case m.setup != nil:
		body = m.setupView()
		help = "yaz: ara · ↑/↓ seç · enter kaydet · esc atla"
	case m.detailID != "":
		body = m.detailView()
		help = "↑/↓ kaydır · esc geri · r yenile · q çıkış"
	case m.tab == tabStandings:
		body = m.standingsView()
		help = "←/→ lig · ↑/↓ kaydır · tab maçlar · f favori takım · q çıkış"
	default:
		body = m.matchesView()
		help = "←/→ gün · t bugün · ↑/↓ seç · enter detay · tab puan durumu · f favori takım · q çıkış"
	}

	var b strings.Builder
	b.WriteString(m.headerView())
	b.WriteString("\n\n")
	b.WriteString(body)
	b.WriteString("\n\n")
	b.WriteString(m.footerView(help))
	return b.String()
}

func (m Model) headerView() string {
	out := logoStyle.Render("⚽ terminalscore") + " "
	if m.setup != nil {
		return out
	}
	for i, name := range []string{"Maçlar", "Puan Durumu"} {
		if tab(i) == m.tab {
			out += activeTabStyle.Render(name)
		} else {
			out += tabStyle.Render(name)
		}
	}
	if fav := m.cfg.FavoriteTeam; fav != nil {
		out += "  " + favStyle.Render("★ "+fav.Name)
	}
	return out
}

func (m Model) footerView(help string) string {
	// hata yoksa da satır boş kalıyor, düzen kaymasın
	var status string
	if m.err != nil {
		status = errorStyle.Render("Hata: " + m.err.Error())
	}
	return status + "\n" + mutedStyle.Render(help)
}

func (m Model) matchesView() string {
	date := "◀  " + titleStyle.Render(turkishDate(m.day)) + "  ▶"
	if now := time.Now().In(m.loc); now.Year() == m.day.Year() && now.YearDay() == m.day.YearDay() {
		date += liveStyle.Render("  Bugün")
	}
	if m.matchesLoaded && len(m.matches) > 0 {
		date += mutedStyle.Render(fmt.Sprintf("  ·  %d maç", len(m.matches)))
	}
	head := date + "\n\n"

	switch {
	case !m.matchesLoaded:
		return head + mutedStyle.Render("  Yükleniyor...")
	case len(m.matches) == 0:
		return head + mutedStyle.Render("  Bu tarihte maç yok. ←/→ ile başka güne geç.")
	}

	var lines []string
	cursorLine := 0
	lastLeague := "\x00"
	for _, mt := range m.matches {
		if key := mt.LeagueID + "|" + mt.LeagueName; key != lastLeague {
			if lastLeague != "\x00" {
				lines = append(lines, "")
			}
			lines = append(lines, leagueStyle.Render(leagueTitle(mt)))
			lastLeague = key
		}
		selected := mt.ID == m.selectedID
		if selected {
			cursorLine = len(lines)
		}
		lines = append(lines, m.matchRow(mt, selected))
	}

	visible := m.height - chromeLines - 2
	start := 0
	if visible > 0 && len(lines) > visible {
		start = min(max(cursorLine-visible/3, 0), len(lines)-visible)
	}
	return head + m.window(lines, 0, start, 2)
}

func leagueTitle(mt espn.Match) string {
	if mt.LeagueName != "" {
		return mt.LeagueName
	}
	return "Diğer"
}

func (m Model) matchRow(mt espn.Match, selected bool) string {
	marker := " "
	if selected {
		marker = selectedStyle.Render("▶")
	}
	favTeam := m.cfg.TeamID()
	star := " "
	if mt.HasTeam(favTeam) {
		star = favStyle.Render("★")
	}

	status := padRight(statusText(mt, m.loc), statusW)
	switch mt.State {
	case espn.Live:
		status = liveStyle.Render(status)
	case espn.Scheduled:
		status = mutedStyle.Render(status)
	}

	homeName := ansi.Truncate(mt.Home.Name, teamW, "…")
	awayName := ansi.Truncate(mt.Away.Name, teamW, "…")
	home, away := homeName, awayName

	score := "-"
	if mt.State != espn.Scheduled {
		score = mt.HomeScore + " - " + mt.AwayScore
		if mt.State == espn.Finished {
			switch compareScores(mt.HomeScore, mt.AwayScore) {
			case 1:
				home = winnerStyle.Render(home)
			case -1:
				away = winnerStyle.Render(away)
			}
		}
	}
	score = center(score, scoreW)
	goal := ""
	switch {
	case m.recentGoal(mt.ID):
		score = blink(score)
		goal = " " + blink(" GOL! ")
	case mt.State == espn.Live:
		score = liveStyle.Render(score)
	}

	switch {
	case selected:
		home, away = selectedStyle.Render(homeName), selectedStyle.Render(awayName)
	case favTeam != "":
		if mt.Home.ID == favTeam {
			home = favStyle.Render(homeName)
		}
		if mt.Away.ID == favTeam {
			away = favStyle.Render(awayName)
		}
	}

	return marker + star + status + padLeft(home, teamW) + " " + score + " " + padRight(away, teamW) + goal
}

// ekran saniyede bir çizildiği için tek/çift saniyeye göre renk değiştirmek yetiyor
func blink(s string) string {
	if time.Now().Second()%2 == 0 {
		return goalStyle.Render(s)
	}
	return goalDimStyle.Render(s)
}

func (m Model) detailView() string {
	return m.window(m.detailLines(), detailFixed, m.scroll, 0)
}

func (m Model) detailLines() []string {
	mt, ok := m.selectedMatch()
	if !ok {
		return []string{mutedStyle.Render("Maç bulunamadı.")}
	}

	width := 2*sideW + minuteW
	var lines []string
	score := "-"
	if mt.State != espn.Scheduled {
		score = mt.HomeScore + " - " + mt.AwayScore
	}
	scoreCol := titleStyle.Render(center(score, minuteW))
	switch {
	case m.recentGoal(mt.ID):
		scoreCol = blink(center(score, minuteW))
	case mt.State == espn.Live:
		scoreCol = liveStyle.Render(center(score, minuteW))
	}
	info := statusLong(mt, m.loc)
	if mt.Venue != "" {
		info += " · " + mt.Venue
	}
	lines = append(lines,
		center(leagueStyle.Render(leagueTitle(mt)), width),
		padLeft(titleStyle.Render(mt.Home.Name), sideW)+scoreCol+padRight(titleStyle.Render(mt.Away.Name), sideW),
		center(mutedStyle.Render(info), width),
		mutedStyle.Render(strings.Repeat("─", width)),
	)

	switch {
	case !m.eventsLoaded:
		lines = append(lines, mutedStyle.Render("Yükleniyor..."))
	case len(m.events) == 0:
		lines = append(lines, mutedStyle.Render("Henüz maç olayı yok."))
	}

	for _, ev := range m.events {
		if label, ok := periodLabels[ev.Label]; ok {
			if label != "" {
				lines = append(lines, center(mutedStyle.Render("── "+label+" ──"), width))
			}
			continue
		}
		icon, texts := describeEvent(ev)
		for i, text := range texts {
			minute := ""
			if i == 0 {
				minute = ev.Minute
			} else {
				icon = " "
			}
			minuteCol := center(mutedStyle.Render(minute), minuteW)
			text = ansi.Truncate(text, sideW-4, "…")
			switch ev.TeamID {
			case mt.Home.ID:
				lines = append(lines, padLeft(text+" "+icon, sideW)+minuteCol)
			case mt.Away.ID:
				lines = append(lines, strings.Repeat(" ", sideW)+minuteCol+padRight(icon+" "+text, sideW))
			default:
				lines = append(lines, center(mutedStyle.Render(minute+" "+text), width))
			}
		}
	}
	return lines
}

// window ekrana sığan kısmı keser: ilk `fixed` satır sabit, gerisi `offset`
// kadar kayar. `extra` gövdenin üstündeki ek satır sayısı.
func (m Model) window(lines []string, fixed, offset, extra int) string {
	visible := m.height - chromeLines - extra
	if m.height == 0 || visible <= fixed || len(lines) <= visible {
		return strings.Join(lines, "\n")
	}
	rest := lines[fixed:]
	room := visible - fixed
	start := min(max(offset, 0), max(len(rest)-room, 0))
	end := min(start+room, len(rest))
	out := append(append([]string{}, lines[:fixed]...), rest[start:end]...)
	return strings.Join(out, "\n")
}

func (m Model) maxScroll(lines []string, fixed int) int {
	room := m.height - chromeLines - fixed
	if m.height == 0 || room <= 0 {
		return 0
	}
	return max(len(lines)-fixed-room, 0)
}

// takımsız olaylar; boş olanlar gösterilmiyor
var periodLabels = map[string]string{
	"Kickoff":          "Maç başladı",
	"Halftime":         "Devre arası",
	"Start 2nd Half":   "İkinci yarı",
	"End Regular Time": "Normal süre bitti",
	"Start Extra Time": "Uzatmalar",
	"End Extra Time":   "Uzatmalar bitti",
	"Penalty Shootout": "Penaltı atışları",
	"Full Time":        "Maç sonu",
	"Start Delay":      "",
	"End Delay":        "",
}

// oyuncu değişikliği iki satır: giren, çıkan
func describeEvent(ev espn.Event) (icon string, lines []string) {
	player := func(i int) string {
		if i < len(ev.Players) {
			return ev.Players[i]
		}
		return ""
	}
	switch ev.Kind {
	case espn.Goal:
		text := player(0)
		if ev.Penalty {
			text += " (P)"
		}
		if ev.OwnGoal {
			text += " (K.K.)"
		}
		return "⚽", []string{text}
	case espn.PenaltyMissed:
		return "✗", []string{player(0) + " (P kaçtı)"}
	case espn.YellowCard:
		return yellowCard, []string{player(0)}
	case espn.RedCard:
		return redCard, []string{player(0)}
	case espn.Substitution:
		return "⇄", []string{subIn + " " + player(0), subOut + " " + player(1)}
	}
	return "•", []string{strings.TrimSpace(ev.Label + " " + player(0))}
}

func (m Model) standingsView() string {
	return m.window(m.standingsLines(), standingsFixed, m.scroll, 0)
}

func (m Model) standingsLines() []string {
	league := standingLeagues[m.standingsIdx]
	title := "◀  " + leagueStyle.Render(league.Name) + "  ▶" +
		mutedStyle.Render(fmt.Sprintf("  (%d/%d)", m.standingsIdx+1, len(standingLeagues)))

	groups, ok := m.standings[league.Slug]
	if !ok {
		if m.err != nil {
			return []string{title, "", mutedStyle.Render("Puan durumu alınamadı.")}
		}
		return []string{title, "", mutedStyle.Render("Yükleniyor...")}
	}

	lines := []string{
		title,
		"",
		mutedStyle.Render(fmt.Sprintf(" %3s  %-*s %3s %3s %3s %3s %4s %4s %4s %4s",
			"#", teamW, "Takım", "O", "G", "B", "M", "AG", "YG", "AV", "P")),
	}

	type zone struct{ desc, color string }
	var zones []zone
	seen := map[string]bool{}

	favTeam := m.cfg.TeamID()
	for gi, g := range groups {
		if len(groups) > 1 {
			if gi > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, titleStyle.Render(g.Name))
		}
		for _, s := range g.Rows {
			name := padRight(ansi.Truncate(s.Team.Name, teamW, "…"), teamW)
			row := fmt.Sprintf("%3d  %s %3d %3d %3d %3d %4d %4d %+4d ",
				s.Rank, name, s.Played, s.Won, s.Drawn, s.Lost,
				s.GoalsFor, s.GoalsAgainst, s.GoalDiff())
			points := titleStyle.Render(fmt.Sprintf("%4d", s.Points))
			if s.Team.ID == favTeam {
				row = favStyle.Render(row + "★")
				points = favStyle.Render(fmt.Sprintf("%3d", s.Points))
			}
			lines = append(lines, zoneBar(s.Zone, s.ZoneColor)+row+points)

			if s.Zone != "" && !seen[zoneLabel(s.Zone)] {
				seen[zoneLabel(s.Zone)] = true
				zones = append(zones, zone{s.Zone, s.ZoneColor})
			}
		}
	}

	// lejant, satır başına sığdığı kadar
	if len(zones) > 0 {
		lines = append(lines, "")
		const maxW = 66
		line, w := "", 0
		for _, z := range zones {
			label := zoneLabel(z.desc)
			itemW := lipgloss.Width(label) + 4
			if w > 0 && w+itemW > maxW {
				lines = append(lines, line)
				line, w = "", 0
			}
			line += zoneBar(z.desc, z.color) + " " + mutedStyle.Render(label) + "  "
			w += itemW
		}
		lines = append(lines, line)
	}
	return lines
}

// 1: ev sahibi önde, -1: deplasman önde, 0: eşit ya da skor yok
func compareScores(home, away string) int {
	h, a := scoreValue(home), scoreValue(away)
	switch {
	case h < 0 || a < 0:
		return 0
	case h > a:
		return 1
	case h < a:
		return -1
	}
	return 0
}
