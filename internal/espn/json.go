package espn

import (
	"encoding/json"
	"strings"
)

// ESPN cevaplarının sadece kullandığımız alanları. Metin alanları `text`
// tipinde: okunurken kontrol karakterleri atılıyor, terminale kaçış kodu
// gidemesin.

type teamJSON struct {
	ID               text `json:"id"`
	DisplayName      text `json:"displayName"`
	ShortDisplayName text `json:"shortDisplayName"`
	Abbreviation     text `json:"abbreviation"`
}

func (t teamJSON) toTeam() Team {
	return Team{
		ID:    string(t.ID),
		Name:  TurkishName(string(t.DisplayName)),
		Short: TurkishName(string(t.ShortDisplayName)),
		Abbr:  string(t.Abbreviation),
	}
}

type scoreboardJSON struct {
	Events []struct {
		ID     text `json:"id"`
		UID    text `json:"uid"`
		Date   text `json:"date"`
		Status struct {
			DisplayClock text `json:"displayClock"`
			Type         struct {
				Name  text `json:"name"`
				State text `json:"state"`
			} `json:"type"`
		} `json:"status"`
		Competitions []struct {
			AltGameNote text `json:"altGameNote"` // lig adı, ör. "English Premier League"
			Venue       struct {
				FullName text `json:"fullName"`
			} `json:"venue"`
			Competitors []struct {
				HomeAway text     `json:"homeAway"`
				Score    text     `json:"score"`
				Team     teamJSON `json:"team"`
			} `json:"competitors"`
		} `json:"competitions"`
	} `json:"events"`
}

type summaryJSON struct {
	KeyEvents []struct {
		Type struct {
			Text text `json:"text"`
			Type text `json:"type"`
		} `json:"type"`
		Clock struct {
			DisplayValue text `json:"displayValue"`
		} `json:"clock"`
		Team struct {
			ID text `json:"id"`
		} `json:"team"`
		Participants []struct {
			Athlete struct {
				DisplayName text `json:"displayName"`
			} `json:"athlete"`
		} `json:"participants"`
	} `json:"keyEvents"`
}

type teamsJSON struct {
	Sports []struct {
		Leagues []struct {
			Teams []struct {
				Team teamJSON `json:"team"`
			} `json:"teams"`
		} `json:"leagues"`
	} `json:"sports"`
}

type standingsJSON struct {
	Children []struct {
		Name      text `json:"name"`
		Standings struct {
			Entries []struct {
				Team teamJSON `json:"team"`
				Note struct {
					Color       text `json:"color"`
					Description text `json:"description"`
				} `json:"note"`
				Stats []struct {
					Name  text    `json:"name"`
					Value float64 `json:"value"`
				} `json:"stats"`
			} `json:"entries"`
		} `json:"standings"`
	} `json:"children"`
}

// Kontrol karakterleri atılmış metin. Sayı olarak gelen alanları da
// (skor bazen "2" bazen 2) metne çeviriyor.
type text string

func (t *text) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	if len(b) == 0 || b[0] != '"' {
		*t = text(clean(string(b)))
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*t = text(clean(s))
	return nil
}

func clean(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || (r >= 0x7f && r < 0xa0) {
			return -1
		}
		return r
	}, s)
}
