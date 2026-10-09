// terminalscore: terminalde Süper Lig maç takibi.
package main

import (
	"fmt"
	"os"
	"time"
	_ "time/tzdata" // saat dilimi verisini binary'nin içine gömer

	tea "github.com/charmbracelet/bubbletea"

	"terminalscore/internal/espn"
	"terminalscore/internal/ui"
)

func main() {
	loc, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		loc = time.FixedZone("TRT", 3*60*60)
	}

	model := ui.New(espn.New(espn.SuperLig), loc)
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "hata:", err)
		os.Exit(1)
	}
}
