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
		body, help = m.setupView(), m.tx.helpSetup
	case m.detailID != "":
		body, help = m.detailView(), m.tx.helpDetail
	case m.tab == tabStandings:
		body, help = m.standingsView(), m.tx.helpStandings
	default:
		body, help = m.matchesView(), m.tx.helpMatches
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
	for i, name := range []string{m.tx.tabMatches, m.tx.tabStandings} {
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
		status = errorStyle.Render(m.tx.errPrefix + m.err.Error())
	}
	return status + "\n" + mutedStyle.Render(help)
}

func (m Model) matchesView() string {
	date := "◀  " + titleStyle.Render(m.tx.date(m.day)) + "  ▶"
	if now := time.Now().In(m.loc); now.Year() == m.day.Year() && now.YearDay() == m.day.YearDay() {
		date += liveStyle.Render("  " + m.tx.today)
	}
	if m.matchesLoaded && len(m.matches) > 0 {
		date += mutedStyle.Render("  ·  " + fmt.Sprintf(m.tx.matchCount, len(m.matches)))
	}
	head := date + "\n\n"

	switch {
	case !m.matchesLoaded:
		return head + mutedStyle.Render("  "+m.tx.loading)
	case len(m.matches) == 0:
		return head + mutedStyle.Render("  "+m.tx.noMatches)
	}

	var lines []string
	cursorLine := 0
	lastLeague := "\x00"
	for _, mt := range m.matches {
		if key := mt.LeagueID + "|" + mt.LeagueName; key != lastLeague {
			if lastLeague != "\x00" {
				lines = append(lines, "")
			}
			lines = append(lines, leagueStyle.Render(m.leagueTitle(mt)))
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

	status := padRight(m.statusText(mt), statusW)
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
		goal = " " + blink(" "+m.tx.goal+" ")
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
		return []string{mutedStyle.Render(m.tx.matchNotFound)}
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
	info := m.statusLong(mt)
	if mt.Venue != "" {
		info += " · " + mt.Venue
	}
	lines = append(lines,
		center(leagueStyle.Render(m.leagueTitle(mt)), width),
		padLeft(titleStyle.Render(mt.Home.Name), sideW)+scoreCol+padRight(titleStyle.Render(mt.Away.Name), sideW),
		center(mutedStyle.Render(info), width),
		mutedStyle.Render(strings.Repeat("─", width)),
	)

	switch {
	case !m.eventsLoaded:
		lines = append(lines, mutedStyle.Render(m.tx.loading))
	case len(m.events) == 0:
		lines = append(lines, mutedStyle.Render(m.tx.noEvents))
	}

	for _, ev := range m.events {
		if label, ok := m.tx.periods[ev.Label]; ok {
			if label != "" {
				lines = append(lines, center(mutedStyle.Render("── "+label+" ──"), width))
			}
			continue
		}
		icon, texts := m.describeEvent(ev)
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

// oyuncu değişikliği iki satır: giren, çıkan
func (m Model) describeEvent(ev espn.Event) (icon string, lines []string) {
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
			text += m.tx.penalty
		}
		if ev.OwnGoal {
			text += m.tx.ownGoal
		}
		return "⚽", []string{text}
	case espn.PenaltyMissed:
		return "✗", []string{player(0) + m.tx.penMissed}
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
	title := "◀  " + leagueStyle.Render(m.tx.league(league)) + "  ▶" +
		mutedStyle.Render(fmt.Sprintf("  (%d/%d)", m.standingsIdx+1, len(standingLeagues)))

	groups, ok := m.standings[league.Slug]
	if !ok {
		if m.err != nil {
			return []string{title, "", mutedStyle.Render(m.tx.standingsFailed)}
		}
		return []string{title, "", mutedStyle.Render(m.tx.loading)}
	}

	c := m.tx.cols
	lines := []string{
		title,
		"",
		mutedStyle.Render(fmt.Sprintf(" %3s  %-*s %3s %3s %3s %3s %4s %4s %4s %4s",
			c[0], teamW, c[1], c[2], c[3], c[4], c[5], c[6], c[7], c[8], c[9])),
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

			if s.Zone != "" && !seen[m.tx.zone(s.Zone)] {
				seen[m.tx.zone(s.Zone)] = true
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
			label := m.tx.zone(z.desc)
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
