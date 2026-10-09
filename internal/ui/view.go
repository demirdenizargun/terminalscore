package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"

	"terminalscore/internal/espn"
)

const (
	statusW = 7
	teamW   = 22
	scoreW  = 7
	sideW   = 30 // maç detayında ev sahibi / deplasman olay sütunları
	minuteW = 9

	chromeLines = 6 // üst başlık + alt bilgi satırları
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
	var status string
	switch {
	case m.err != nil:
		status = errorStyle.Render("Hata: " + m.err.Error())
	case !m.updated.IsZero():
		ago := int(time.Since(m.updated).Seconds())
		status = mutedStyle.Render(fmt.Sprintf("Son güncelleme %d sn önce", ago))
		if m.hasActiveMatch() {
			every := matchesRefresh
			if m.detailID != "" {
				every = eventsRefresh
			}
			status += liveStyle.Render(fmt.Sprintf(" · canlı: %d sn'de bir yenileniyor", int(every.Seconds())))
		}
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

	// İmleç ekranda kalsın diye listeyi imlecin etrafında kesiyoruz.
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

// blink, saniyede bir zemin rengi değiştirerek yanıp sönme etkisi verir
// (ekran her saniye yeniden çizildiği için).
func blink(s string) string {
	if time.Now().Second()%2 == 0 {
		return goalStyle.Render(s)
	}
	return goalDimStyle.Render(s)
}

func (m Model) detailView() string {
	mt, ok := m.selectedMatch()
	if !ok {
		return mutedStyle.Render("Maç bulunamadı.")
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
	const fixed = 4 // kaydırılırken sabit kalan üst kısım

	switch {
	case !m.eventsLoaded:
		lines = append(lines, mutedStyle.Render("Yükleniyor..."))
	case len(m.events) == 0:
		lines = append(lines, mutedStyle.Render("Henüz maç olayı yok."))
	}

	for _, ev := range m.events {
		// devre arası, maç sonu gibi ara başlıklar; boş karşılıklılar (gecikmeler) gizlenir
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
				icon = " " // değişikliğin ikinci satırında ikon ve dakika tekrarlanmaz
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

	return m.window(lines, fixed, m.scroll, 0)
}

// window, ekrana sığmayan satırlardan görünür kısmı seçer. İlk `fixed`
// satır hep görünür, kalanı `offset` kadar kaydırılır. `extra`, gövdenin
// üstüne eklenen başka satırların sayısıdır (ör. tarih satırı).
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

// periodLabels, ESPN'in takımsız olaylarının Türkçe karşılıkları.
// Boş karşılık, olayın gösterilmeyeceği anlamına gelir.
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

// describeEvent, olayın ikonunu ve satır(lar)ını döndürür.
// Oyuncu değişikliği iki satırdır: giren ve çıkan oyuncu.
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
	league := standingLeagues[m.standingsIdx]
	title := "◀  " + leagueStyle.Render(league.Name) + "  ▶" +
		mutedStyle.Render(fmt.Sprintf("  (%d/%d)", m.standingsIdx+1, len(standingLeagues)))

	groups, ok := m.standings[league.Slug]
	if !ok {
		if m.err != nil {
			return title + "\n\n" + mutedStyle.Render("Puan durumu alınamadı.")
		}
		return title + "\n\n" + mutedStyle.Render("Yükleniyor...")
	}

	lines := []string{
		title,
		"",
		mutedStyle.Render(fmt.Sprintf("%3s  %-*s %3s %3s %3s %3s %4s %4s %4s %4s",
			"#", teamW, "Takım", "O", "G", "B", "M", "AG", "YG", "AV", "P")),
	}
	const fixed = 3

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
			lines = append(lines, row+points)
		}
	}
	return m.window(lines, fixed, m.scroll, 0)
}

// compareScores: ev sahibi önde 1, deplasman önde -1, eşit/bilinmiyor 0.
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
