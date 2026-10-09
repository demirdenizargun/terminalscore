package ui

import (
	"fmt"
	"time"

	"github.com/demirdenizargun/terminalscore/internal/espn"
	"github.com/demirdenizargun/terminalscore/internal/i18n"
)

// Arayüzdeki bütün yazılar. Yeni dil = yeni bir texts değeri.
type texts struct {
	tabMatches, tabStandings string

	today, matchCount, loading, noMatches, other string
	matchNotFound, noEvents, standingsFailed     string
	errPrefix                                    string

	helpMatches, helpStandings, helpDetail, helpSetup string

	setupTitle, setupSubtitle, setupSearch, setupLoading    string
	setupNoTeams, setupNoTeamsHint, setupHint, setupNoMatch string

	cols [10]string // #, takım, O, G, B, M, AG, YG, AV, P

	date func(time.Time) string

	// kısa durumlar (listede) ve uzun karşılıkları (detayda)
	halftime, fullTime, aet, pens, abandoned, postponed, canceled, live string
	long                                                                map[string]string
	liveLong, kickoff                                                   string

	goal, penalty, ownGoal, penMissed string

	periods map[string]string // ESPN'in takımsız olayları; boş = gösterme
	zone    func(string) string
	league  func(espn.League) string
}

var trTexts = texts{
	tabMatches:      "Maçlar",
	tabStandings:    "Puan Durumu",
	today:           "Bugün",
	matchCount:      "%d maç",
	loading:         "Yükleniyor...",
	noMatches:       "Bu tarihte maç yok. ←/→ ile başka güne geç.",
	other:           "Diğer",
	matchNotFound:   "Maç bulunamadı.",
	noEvents:        "Henüz maç olayı yok.",
	standingsFailed: "Puan durumu alınamadı.",
	errPrefix:       "Hata: ",

	helpMatches:   "←/→ gün · t bugün · ↑/↓ seç · enter detay · tab puan durumu · f favori takım · q çıkış",
	helpStandings: "←/→ lig · ↑/↓ kaydır · tab maçlar · f favori takım · q çıkış",
	helpDetail:    "↑/↓ kaydır · esc geri · r yenile · q çıkış",
	helpSetup:     "yaz: ara · ↑/↓ seç · enter kaydet · esc atla",

	setupTitle:       "Favori takımın hangisi?",
	setupSubtitle:    "Favori takımının maçları ve ligi listede en üstte görünür.",
	setupSearch:      "Ara: ",
	setupLoading:     "Takımlar yükleniyor...",
	setupNoTeams:     "Takım listesi alınamadı, internet bağlantını kontrol et.",
	setupNoTeamsHint: "esc ile şimdilik atlayıp sonra f tuşuyla tekrar deneyebilirsin.",
	setupHint:        "Takım adını yazmaya başla (ör. galatasaray, besiktas, real madrid)",
	setupNoMatch:     "Eşleşen takım yok.",

	cols: [10]string{"#", "Takım", "O", "G", "B", "M", "AG", "YG", "AV", "P"},

	date: func(t time.Time) string {
		months := [...]string{"Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran",
			"Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık"}
		days := [...]string{"Pazar", "Pazartesi", "Salı", "Çarşamba", "Perşembe", "Cuma", "Cumartesi"}
		return fmt.Sprintf("%d %s %d, %s", t.Day(), months[t.Month()-1], t.Year(), days[t.Weekday()])
	},

	halftime: "İY", fullTime: "MS", aet: "UZS", pens: "PEN",
	abandoned: "YRD", postponed: "ERT", canceled: "İPT", live: "CANLI",
	long: map[string]string{
		"İY": "Devre arası", "MS": "Maç sonu", "UZS": "Uzatmalar sonunda", "PEN": "Penaltılar sonunda",
		"ERT": "Ertelendi", "İPT": "İptal edildi", "YRD": "Yarıda kaldı",
	},
	liveLong: "Canlı · ",
	kickoff:  "Başlama: ",

	goal: "GOL!", penalty: " (P)", ownGoal: " (K.K.)", penMissed: " (P kaçtı)",

	periods: map[string]string{
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
	},
	zone:   zoneReplacerTR.Replace,
	league: func(l espn.League) string { return l.NameTR },
}

var enTexts = texts{
	tabMatches:      "Matches",
	tabStandings:    "Standings",
	today:           "Today",
	matchCount:      "%d matches",
	loading:         "Loading...",
	noMatches:       "No matches on this day. Use ←/→ to change the day.",
	other:           "Other",
	matchNotFound:   "Match not found.",
	noEvents:        "No match events yet.",
	standingsFailed: "Couldn't load standings.",
	errPrefix:       "Error: ",

	helpMatches:   "←/→ day · t today · ↑/↓ select · enter details · tab standings · f favorite team · q quit",
	helpStandings: "←/→ league · ↑/↓ scroll · tab matches · f favorite team · q quit",
	helpDetail:    "↑/↓ scroll · esc back · r refresh · q quit",
	helpSetup:     "type to search · ↑/↓ select · enter save · esc skip",

	setupTitle:       "What's your favorite team?",
	setupSubtitle:    "Your team's matches and league are shown at the top.",
	setupSearch:      "Search: ",
	setupLoading:     "Loading teams...",
	setupNoTeams:     "Couldn't load the team list, check your internet connection.",
	setupNoTeamsHint: "Press esc to skip for now, then f to try again later.",
	setupHint:        "Start typing a team name (e.g. arsenal, real madrid, galatasaray)",
	setupNoMatch:     "No matching team.",

	cols: [10]string{"#", "Team", "P", "W", "D", "L", "GF", "GA", "GD", "Pts"},

	date: func(t time.Time) string { return t.Format("Monday, 2 January 2006") },

	halftime: "HT", fullTime: "FT", aet: "AET", pens: "PEN",
	abandoned: "ABD", postponed: "PP", canceled: "CAN", live: "LIVE",
	long: map[string]string{
		"HT": "Half-time", "FT": "Full time", "AET": "After extra time", "PEN": "After penalties",
		"PP": "Postponed", "CAN": "Cancelled", "ABD": "Abandoned",
	},
	liveLong: "Live · ",
	kickoff:  "Kick-off: ",

	goal: "GOAL!", penalty: " (pen)", ownGoal: " (o.g.)", penMissed: " (pen missed)",

	periods: map[string]string{
		"Kickoff":          "Kick-off",
		"Halftime":         "Half-time",
		"Start 2nd Half":   "Second half",
		"End Regular Time": "End of normal time",
		"Start Extra Time": "Extra time",
		"End Extra Time":   "End of extra time",
		"Penalty Shootout": "Penalty shootout",
		"Full Time":        "Full time",
		"Start Delay":      "",
		"End Delay":        "",
	},
	zone:   zoneReplacerEN.Replace,
	league: func(l espn.League) string { return l.NameEN },
}

func textsFor(l i18n.Lang) *texts {
	if l == i18n.TR {
		return &trTexts
	}
	return &enTexts
}

// listedeki kısa durum: saat, dakika, İY/HT, MS/FT...
func (m Model) statusText(mt espn.Match) string {
	tx := m.tx
	switch mt.State {
	case espn.Live:
		if mt.StatusName == "STATUS_HALFTIME" {
			return tx.halftime
		}
		if mt.Clock != "" {
			return mt.Clock
		}
		return tx.live
	case espn.Finished:
		switch mt.StatusName {
		case "STATUS_FINAL_PEN":
			return tx.pens
		case "STATUS_FINAL_AET":
			return tx.aet
		case "STATUS_ABANDONED":
			return tx.abandoned
		}
		return tx.fullTime
	}
	switch mt.StatusName {
	case "STATUS_POSTPONED":
		return tx.postponed
	case "STATUS_CANCELED":
		return tx.canceled
	}
	return mt.Kickoff.In(m.loc).Format("15:04")
}

func (m Model) statusLong(mt espn.Match) string {
	s := m.statusText(mt)
	if long, ok := m.tx.long[s]; ok {
		return long
	}
	if mt.State == espn.Live {
		return m.tx.liveLong + s
	}
	return m.tx.kickoff + s
}

// bilinen liglerde kendi adımız, diğerlerinde ESPN'inki
func (m Model) leagueTitle(mt espn.Match) string {
	if l, ok := espn.FindLeague(mt.LeagueID); ok {
		return m.tx.league(l)
	}
	if mt.LeagueName != "" {
		return mt.LeagueName
	}
	return m.tx.other
}
