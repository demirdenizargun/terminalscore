package ui

import (
	"fmt"
	"time"

	"github.com/demirdenizargun/terminalscore/internal/espn"
)

var months = [...]string{"Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran",
	"Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık"}

var weekdays = [...]string{"Pazar", "Pazartesi", "Salı", "Çarşamba", "Perşembe", "Cuma", "Cumartesi"}

// ör. "9 Ekim 2026, Cuma"
func turkishDate(t time.Time) string {
	return fmt.Sprintf("%d %s %d, %s", t.Day(), months[t.Month()-1], t.Year(), weekdays[t.Weekday()])
}

// listedeki kısa durum: saat, dakika, İY, MS...
func statusText(m espn.Match, loc *time.Location) string {
	switch m.State {
	case espn.Live:
		if m.StatusName == "STATUS_HALFTIME" {
			return "İY"
		}
		if m.Clock != "" {
			return m.Clock
		}
		return "CANLI"
	case espn.Finished:
		switch m.StatusName {
		case "STATUS_FINAL_PEN":
			return "PEN"
		case "STATUS_FINAL_AET":
			return "UZS"
		case "STATUS_ABANDONED":
			return "YRD"
		}
		return "MS"
	}
	switch m.StatusName {
	case "STATUS_POSTPONED":
		return "ERT"
	case "STATUS_CANCELED":
		return "İPT"
	}
	return m.Kickoff.In(loc).Format("15:04")
}

func statusLong(m espn.Match, loc *time.Location) string {
	switch s := statusText(m, loc); s {
	case "İY":
		return "Devre arası"
	case "MS":
		return "Maç sonu"
	case "UZS":
		return "Uzatmalar sonunda"
	case "PEN":
		return "Penaltılar sonunda"
	case "ERT":
		return "Ertelendi"
	case "İPT":
		return "İptal edildi"
	case "YRD":
		return "Yarıda kaldı"
	default:
		if m.State == espn.Live {
			return "Canlı · " + s
		}
		return "Başlama: " + s
	}
}
