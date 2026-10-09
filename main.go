// terminalscore: terminalde canlı futbol takibi.
package main

import (
	"fmt"
	"os"
	"time"
	_ "time/tzdata" // saat dilimi verisini binary'nin içine gömer

	tea "github.com/charmbracelet/bubbletea"

	"github.com/demirdenizargun/terminalscore/internal/config"
	"github.com/demirdenizargun/terminalscore/internal/espn"
	"github.com/demirdenizargun/terminalscore/internal/ui"
)

// version, sürüm yayınlanırken derleme komutuyla doldurulur
// (-ldflags "-X main.version=0.1.0"); elle derlenen sürümde "dev" kalır.
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Println("terminalscore", version)
			return
		case "-h", "--help", "help":
			fmt.Println("terminalscore — terminalde canlı futbol takibi")
			fmt.Println()
			fmt.Println("Kullanım: terminalscore [--version]")
			fmt.Println("Program açıkken tuşlar ekranın altında yazar.")
			return
		}
	}

	loc, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		loc = time.FixedZone("TRT", 3*60*60)
	}

	cfg, err := config.Load()
	if err != nil {
		// Bozuk ayar dosyası programı durdurmasın; favori takım yeniden sorulur.
		fmt.Fprintln(os.Stderr, "uyarı: ayarlar okunamadı:", err)
	}

	model := ui.New(espn.New(), loc, cfg)
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "hata:", err)
		os.Exit(1)
	}
}
