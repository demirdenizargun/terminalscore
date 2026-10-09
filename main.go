package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/demirdenizargun/terminalscore/internal/config"
	"github.com/demirdenizargun/terminalscore/internal/espn"
	"github.com/demirdenizargun/terminalscore/internal/i18n"
	"github.com/demirdenizargun/terminalscore/internal/ui"
)

var version = "dev" // goreleaser -X main.version ile set ediyor

func main() {
	cfg, cfgErr := config.Load()

	// öncelik: --lang, ayar dosyası, sistem dili
	lang := i18n.Detect()
	if l, ok := i18n.Parse(cfg.Language); ok {
		lang = l
	}

	langFlag := flag.String("lang", "", "tr | en")
	showVersion := flag.Bool("version", false, "")
	flag.BoolVar(showVersion, "v", false, "")
	flag.Usage = func() {
		if lang == i18n.TR {
			fmt.Println("terminalscore: terminalde canlı futbol takibi")
			fmt.Println("\nkullanım: terminalscore [--lang tr|en] [--version]")
		} else {
			fmt.Println("terminalscore: live football scores in your terminal")
			fmt.Println("\nusage: terminalscore [--lang tr|en] [--version]")
		}
	}
	flag.Parse()

	if *showVersion {
		fmt.Println("terminalscore", version)
		return
	}
	if *langFlag != "" {
		l, ok := i18n.Parse(*langFlag)
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown language %q (tr, en)\n", *langFlag)
			os.Exit(2)
		}
		lang = l
	}
	if cfgErr != nil {
		fmt.Fprintln(os.Stderr, "config:", cfgErr)
	}

	model := ui.New(espn.New(), time.Local, cfg, lang)
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
