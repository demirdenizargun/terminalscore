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
	teamW   = 20
	scoreW  = 7
	sideW   = 30 // maç detayında ev sahibi / deplasman olay sütunları
	minuteW = 9
)

func (m Model) View() string {
	var body, help string
	switch {
	case m.detailID != "":
		body = m.detailView()
		help = "↑/↓ kaydır · esc geri · r yenile · q çıkış"
	case m.tab == tabStandings:
		body = m.standingsView()
		help = "tab maçlar · r yenile · q çıkış"
	default:
		body = m.matchesView()
		help = "←/→ gün · t bugün · ↑/↓ seç · enter detay · tab puan durumu · q çıkış"
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
	tabs := []string{"Maçlar", "Puan Durumu"}
	out := logoStyle.Render("⚽ terminalscore") + " "
	for i, name := range tabs {
		if tab(i) == m.tab {
			out += activeTabStyle.Render(name)
		} else {
			out += tabStyle.Render(name)
		}
	}
	return out
}

func (m Model) footerView(help string) string {
	var status string
	switch {
	case m.err != nil:
		status = errorStyle.Render("Hata: " + m.err.Error())
	case !m.updated.IsZero():
		status = mutedStyle.Render("Son güncelleme " + m.updated.In(m.loc).Format("15:04:05"))
		if m.hasActiveMatch() {
			status += liveStyle.Render(" · canlı takip açık")
		}
	}
	return status + "\n" + mutedStyle.Render(help)
}

func (m Model) matchesView() string {
	var b strings.Builder

	date := "◀  " + titleStyle.Render(turkishDate(m.day)) + "  ▶"
	if now := time.Now().In(m.loc); now.Year() == m.day.Year() && now.YearDay() == m.day.YearDay() {
		date += liveStyle.Render("  Bugün")
	}
	b.WriteString(date + "\n\n")
	b.WriteString(titleStyle.Render("Trendyol Süper Lig") + "\n")

	switch {
	case !m.matchesLoaded:
		b.WriteString(mutedStyle.Render("  Yükleniyor..."))
	case len(m.matches) == 0:
		b.WriteString(mutedStyle.Render("  Bu tarihte Süper Lig maçı yok. ←/→ ile başka güne geç."))
	}

	for i, mt := range m.matches {
		b.WriteString(m.matchRow(mt, i == m.cursor))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m Model) matchRow(mt espn.Match, selected bool) string {
	marker := "  "
	if selected {
		marker = selectedStyle.Render("▶ ")
	}

	status := padRight(statusText(mt, m.loc), statusW)
	switch mt.State {
	case espn.Live:
		status = liveStyle.Render(status)
	case espn.Scheduled:
		status = mutedStyle.Render(status)
	}

	home, away := mt.Home.Name, mt.Away.Name
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
	if mt.State == espn.Live {
		score = liveStyle.Render(score)
	}
	if selected {
		home, away = selectedStyle.Render(mt.Home.Name), selectedStyle.Render(mt.Away.Name)
	}

	return marker + status + padLeft(home, teamW) + " " + score + " " + padRight(away, teamW)
}

func (m Model) detailView() string {
	mt, ok := m.selectedMatch()
	if !ok {
		return mutedStyle.Render("Maç bulunamadı.")
	}

	var lines []string
	score := "-"
	if mt.State != espn.Scheduled {
		score = mt.HomeScore + " - " + mt.AwayScore
	}
	scoreStyle := titleStyle
	if mt.State == espn.Live {
		scoreStyle = liveStyle
	}
	lines = append(lines,
		padLeft(titleStyle.Render(mt.Home.Name), sideW)+scoreStyle.Render(center(score, minuteW))+padRight(titleStyle.Render(mt.Away.Name), sideW),
		center(mutedStyle.Render(statusLong(mt, m.loc)), 2*sideW+minuteW),
	)
	if mt.Venue != "" {
		lines = append(lines, center(mutedStyle.Render(mt.Venue), 2*sideW+minuteW))
	}
	lines = append(lines, mutedStyle.Render(strings.Repeat("─", 2*sideW+minuteW)))

	switch {
	case !m.eventsLoaded:
		lines = append(lines, mutedStyle.Render("Yükleniyor..."))
	case len(m.events) == 0:
		lines = append(lines, mutedStyle.Render("Henüz maç olayı yok."))
	}

	for _, ev := range m.events {
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
				lines = append(lines, center(mutedStyle.Render(minute+" "+text), 2*sideW+minuteW))
			}
		}
	}

	return m.scrolled(lines)
}

// scrolled, ekrana sığmayan satırları kaydırma konumuna göre keser.
func (m Model) scrolled(lines []string) string {
	const header = 4 // sabit tutulan skor bölümü
	visible := m.height - 8
	if visible <= header || len(lines) <= visible {
		return strings.Join(lines, "\n")
	}
	rest := lines[header:]
	room := visible - header
	start := min(m.scroll, max(len(rest)-room, 0))
	end := min(start+room, len(rest))
	return strings.Join(append(lines[:header:header], rest[start:end]...), "\n")
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
	if len(m.standings) == 0 {
		if m.err != nil {
			return mutedStyle.Render("Puan durumu alınamadı.")
		}
		return mutedStyle.Render("Yükleniyor...")
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Trendyol Süper Lig · Puan Durumu") + "\n\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("%3s  %-*s %3s %3s %3s %3s %4s %4s %4s %4s",
		"#", teamW, "Takım", "O", "G", "B", "M", "AG", "YG", "AV", "P")) + "\n")

	for _, s := range m.standings {
		row := fmt.Sprintf("%3d  %s %3d %3d %3d %3d %4d %4d %+4d ",
			s.Rank, padRight(s.Team.Name, teamW), s.Played, s.Won, s.Drawn, s.Lost,
			s.GoalsFor, s.GoalsAgainst, s.GoalDiff())
		b.WriteString(row + titleStyle.Render(fmt.Sprintf("%4d", s.Points)) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// compareScores: ev sahibi önde 1, deplasman önde -1, eşit/bilinmiyor 0.
func compareScores(home, away string) int {
	var h, a int
	if _, err := fmt.Sscan(home, &h); err != nil {
		return 0
	}
	if _, err := fmt.Sscan(away, &a); err != nil {
		return 0
	}
	switch {
	case h > a:
		return 1
	case h < a:
		return -1
	}
	return 0
}
