package espn

import "encoding/json"

// Bu dosyadaki tipler sadece ESPN'in JSON cevaplarını okumak için var;
// uygulamanın geri kalanı espn.go'daki sade tipleri kullanır.

type teamJSON struct {
	ID               string `json:"id"`
	DisplayName      string `json:"displayName"`
	ShortDisplayName string `json:"shortDisplayName"`
	Abbreviation     string `json:"abbreviation"`
}

func (t teamJSON) toTeam() Team {
	return Team{
		ID:    t.ID,
		Name:  TurkishName(t.DisplayName),
		Short: TurkishName(t.ShortDisplayName),
		Abbr:  t.Abbreviation,
	}
}

type scoreboardJSON struct {
	Events []struct {
		ID     string `json:"id"`
		Date   string `json:"date"`
		Status struct {
			DisplayClock string `json:"displayClock"`
			Type         struct {
				Name  string `json:"name"`
				State string `json:"state"`
			} `json:"type"`
		} `json:"status"`
		Competitions []struct {
			Venue struct {
				FullName string `json:"fullName"`
			} `json:"venue"`
			Competitors []struct {
				HomeAway string     `json:"homeAway"`
				Score    flexString `json:"score"`
				Team     teamJSON   `json:"team"`
			} `json:"competitors"`
		} `json:"competitions"`
	} `json:"events"`
}

type summaryJSON struct {
	KeyEvents []struct {
		Type struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"type"`
		Clock struct {
			DisplayValue string `json:"displayValue"`
		} `json:"clock"`
		Team struct {
			ID string `json:"id"`
		} `json:"team"`
		Participants []struct {
			Athlete struct {
				DisplayName string `json:"displayName"`
			} `json:"athlete"`
		} `json:"participants"`
	} `json:"keyEvents"`
}

type standingsJSON struct {
	Children []struct {
		Standings struct {
			Entries []struct {
				Team  teamJSON `json:"team"`
				Stats []struct {
					Name  string  `json:"name"`
					Value float64 `json:"value"`
				} `json:"stats"`
			} `json:"entries"`
		} `json:"standings"`
	} `json:"children"`
}

// flexString, ESPN'in bazen metin ("2") bazen sayı (2) olarak
// gönderdiği alanları tek tipte okur.
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*f = flexString(s)
		return nil
	}
	*f = flexString(b)
	return nil
}
