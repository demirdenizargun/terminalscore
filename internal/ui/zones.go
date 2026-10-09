package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Sıra önemli: uzun ifadeler kısalarından önce gelmeli.
var zoneReplacer = strings.NewReplacer(
	"Knockout phase playoffs - unseeded", "Play-off",
	"Knockout phase playoffs - seeded", "Play-off (seri başı)",
	"Qualifies for round of 16", "Son 16",
	"Qualifies for MLS Cup Playoffs - Round One Best-of-3 series", "MLS play-off",
	"Qualifies for MLS Cup Playoffs - Wild Card Matches", "MLS play-off (wild card)",
	"Champions League", "Şampiyonlar Ligi",
	"Europa League", "Avrupa Ligi",
	"Conference League", "Konferans Ligi",
	"Championship playoffs", "Şampiyonluk grubu",
	"European playoffs", "Avrupa play-off",
	"Relegation playoffs", "Küme düşme play-off",
	"Relegation playoff", "Küme düşme play-off",
	"Relegated", "Küme düşme",
	"Relegation", "Küme düşme",
	"Promotion playoffs", "Üst lig play-off",
	"Promotion playoff", "Üst lig play-off",
	"Promotion", "Üst lige çıkış",
	"Eliminated", "Elendi",
	" qualifying", " ön eleme",
)

func zoneLabel(desc string) string { return zoneReplacer.Replace(desc) }

var (
	zoneGreen  = lipgloss.Color("#3FB950")
	zoneLight  = lipgloss.Color("#9BE9A8")
	zoneBlue   = lipgloss.Color("#58A6FF")
	zoneCyan   = lipgloss.Color("#39C5CF")
	zoneOrange = lipgloss.Color("#F0883E")
	zoneRed    = lipgloss.Color("#F85149")
)

// ESPN'in renkleri tutarsız (aynı bölge farklı tonlarda), o yüzden addan seçiyoruz.
func zoneColor(desc, espnColor string) lipgloss.Color {
	d := strings.ToLower(desc)
	switch {
	case strings.Contains(d, "relegation playoff"):
		return zoneOrange
	case strings.Contains(d, "relegat"), strings.Contains(d, "eliminated"):
		return zoneRed
	case strings.Contains(d, "conference"):
		return zoneCyan
	case strings.Contains(d, "europa"), strings.Contains(d, "unseeded"):
		return zoneBlue
	case strings.Contains(d, "round one"):
		return zoneGreen
	case strings.Contains(d, "qualifying"), strings.Contains(d, "playoff"):
		return zoneLight
	case strings.Contains(d, "champions"), strings.Contains(d, "round of 16"),
		strings.Contains(d, "promotion"):
		return zoneGreen
	case espnColor != "":
		return lipgloss.Color(espnColor)
	}
	return zoneLight
}

func zoneBar(desc, espnColor string) string {
	if desc == "" {
		return " "
	}
	return lipgloss.NewStyle().Foreground(zoneColor(desc, espnColor)).Render("▌")
}
