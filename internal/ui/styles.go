package ui

import "github.com/charmbracelet/lipgloss"

var (
	red    = lipgloss.Color("#E30A17")
	green  = lipgloss.Color("#3FB950")
	yellow = lipgloss.Color("#F2CC0C")
	muted  = lipgloss.Color("#8B949E")

	logoStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(red).Padding(0, 1)
	tabStyle       = lipgloss.NewStyle().Foreground(muted).Padding(0, 1)
	activeTabStyle = lipgloss.NewStyle().Bold(true).Underline(true).Padding(0, 1)

	titleStyle    = lipgloss.NewStyle().Bold(true)
	mutedStyle    = lipgloss.NewStyle().Foreground(muted)
	liveStyle     = lipgloss.NewStyle().Bold(true).Foreground(green)
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(red)
	winnerStyle   = lipgloss.NewStyle().Bold(true)
	errorStyle    = lipgloss.NewStyle().Foreground(red)
	favStyle      = lipgloss.NewStyle().Bold(true).Foreground(yellow)
	goalStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#000000")).Background(green)
	goalDimStyle  = lipgloss.NewStyle().Bold(true).Foreground(green)
	leagueStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#58A6FF"))

	yellowCard = lipgloss.NewStyle().Foreground(yellow).Render("▮")
	redCard    = lipgloss.NewStyle().Foreground(red).Render("▮")
	subIn      = lipgloss.NewStyle().Foreground(green).Render("↑")
	subOut     = lipgloss.NewStyle().Foreground(red).Render("↓")
)

func padLeft(s string, w int) string {
	return lipgloss.NewStyle().Width(w).Align(lipgloss.Right).Render(s)
}

func padRight(s string, w int) string {
	return lipgloss.NewStyle().Width(w).Align(lipgloss.Left).Render(s)
}

func center(s string, w int) string {
	return lipgloss.NewStyle().Width(w).Align(lipgloss.Center).Render(s)
}
