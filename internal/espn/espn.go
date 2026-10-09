// Package espn, ESPN'in herkese açık (anahtarsız) futbol API'sinden
// maç, maç olayı ve puan durumu verisini çeker.
package espn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

const baseURL = "https://site.api.espn.com/apis"

// SuperLig, ESPN'deki Trendyol Süper Lig kodu.
const SuperLig = "tur.1"

type Client struct {
	http   *http.Client
	league string
}

func New(league string) *Client {
	return &Client{
		http:   &http.Client{Timeout: 15 * time.Second},
		league: league,
	}
}

// State, maçın genel durumu: başlamadı, oynanıyor, bitti.
type State string

const (
	Scheduled State = "pre"
	Live      State = "in"
	Finished  State = "post"
)

type Team struct {
	ID    string
	Name  string
	Short string
	Abbr  string
}

type Match struct {
	ID         string
	Kickoff    time.Time
	Home, Away Team
	HomeScore  string
	AwayScore  string
	State      State
	StatusName string // ESPN durum kodu, ör. STATUS_HALFTIME
	Clock      string // ör. "67'" veya "90'+4'"
	Venue      string
}

type EventKind int

const (
	OtherEvent EventKind = iota
	Goal
	YellowCard
	RedCard
	Substitution
	PenaltyMissed
)

// Event, maç içindeki bir olay (gol, kart, oyuncu değişikliği...).
type Event struct {
	Minute  string
	Kind    EventKind
	Label   string // ESPN'in olay adı, bilinmeyen türler için
	TeamID  string
	Players []string // değişiklikte [0] giren, [1] çıkan oyuncu
	Penalty bool
	OwnGoal bool
}

type Standing struct {
	Rank         int
	Team         Team
	Played       int
	Won          int
	Drawn        int
	Lost         int
	GoalsFor     int
	GoalsAgainst int
	Points       int
}

func (s Standing) GoalDiff() int { return s.GoalsFor - s.GoalsAgainst }

// Matches, verilen gündeki lig maçlarını başlama saatine göre sıralı döndürür.
func (c *Client) Matches(ctx context.Context, day time.Time) ([]Match, error) {
	url := fmt.Sprintf("%s/site/v2/sports/soccer/%s/scoreboard?dates=%s",
		baseURL, c.league, day.Format("20060102"))

	var r scoreboardJSON
	if err := c.get(ctx, url, &r); err != nil {
		return nil, err
	}

	matches := make([]Match, 0, len(r.Events))
	for _, e := range r.Events {
		if len(e.Competitions) == 0 {
			continue
		}
		comp := e.Competitions[0]
		m := Match{
			ID:         e.ID,
			Kickoff:    parseTime(e.Date),
			State:      State(e.Status.Type.State),
			StatusName: e.Status.Type.Name,
			Clock:      e.Status.DisplayClock,
			Venue:      comp.Venue.FullName,
		}
		for _, cp := range comp.Competitors {
			team := cp.Team.toTeam()
			if cp.HomeAway == "home" {
				m.Home, m.HomeScore = team, string(cp.Score)
			} else {
				m.Away, m.AwayScore = team, string(cp.Score)
			}
		}
		matches = append(matches, m)
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].Kickoff.Before(matches[j].Kickoff)
	})
	return matches, nil
}

// Events, bir maçın önemli olaylarını (gol, kart, değişiklik) döndürür.
func (c *Client) Events(ctx context.Context, matchID string) ([]Event, error) {
	url := fmt.Sprintf("%s/site/v2/sports/soccer/%s/summary?event=%s",
		baseURL, c.league, matchID)

	var r summaryJSON
	if err := c.get(ctx, url, &r); err != nil {
		return nil, err
	}

	events := make([]Event, 0, len(r.KeyEvents))
	for _, ke := range r.KeyEvents {
		ev := Event{
			Minute: ke.Clock.DisplayValue,
			Label:  ke.Type.Text,
			TeamID: ke.Team.ID,
		}
		ev.Kind, ev.Penalty, ev.OwnGoal = classify(ke.Type.Type)
		for _, p := range ke.Participants {
			ev.Players = append(ev.Players, p.Athlete.DisplayName)
		}
		events = append(events, ev)
	}
	return events, nil
}

// Standings, ligin güncel puan durumunu sıralı döndürür.
func (c *Client) Standings(ctx context.Context) ([]Standing, error) {
	url := fmt.Sprintf("%s/v2/sports/soccer/%s/standings", baseURL, c.league)

	var r standingsJSON
	if err := c.get(ctx, url, &r); err != nil {
		return nil, err
	}
	if len(r.Children) == 0 {
		return nil, fmt.Errorf("puan durumu bulunamadı")
	}

	entries := r.Children[0].Standings.Entries
	rows := make([]Standing, 0, len(entries))
	for _, e := range entries {
		s := Standing{Team: e.Team.toTeam()}
		for _, st := range e.Stats {
			v := int(st.Value)
			switch st.Name {
			case "rank":
				s.Rank = v
			case "gamesPlayed":
				s.Played = v
			case "wins":
				s.Won = v
			case "ties":
				s.Drawn = v
			case "losses":
				s.Lost = v
			case "pointsFor":
				s.GoalsFor = v
			case "pointsAgainst":
				s.GoalsAgainst = v
			case "points":
				s.Points = v
			}
		}
		rows = append(rows, s)
	}

	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Rank < rows[j].Rank })
	return rows, nil
}

func (c *Client) get(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "terminalscore/0.1")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("bağlantı hatası: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ESPN %d hatası döndürdü", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func classify(t string) (kind EventKind, penalty, ownGoal bool) {
	switch {
	case t == "penalty---scored":
		return Goal, true, false
	case strings.Contains(t, "own-goal"):
		return Goal, false, true
	case strings.HasPrefix(t, "goal"):
		return Goal, false, false
	case strings.HasPrefix(t, "penalty---"):
		return PenaltyMissed, true, false
	case strings.Contains(t, "red-card"):
		return RedCard, false, false
	case t == "yellow-card":
		return YellowCard, false, false
	case t == "substitution":
		return Substitution, false, false
	}
	return OtherEvent, false, false
}

// ESPN tarihleri bazen saniyesiz geliyor ("2026-10-09T17:00Z").
func parseTime(s string) time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04Z07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
