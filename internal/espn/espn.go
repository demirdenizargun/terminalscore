// Package espn, ESPN'in anahtarsız futbol API'si için istemci.
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

// "all": bütün ligler tek istekte
const allLeagues = "all"

type Client struct {
	http *http.Client
}

func New() *Client {
	return &Client{http: &http.Client{Timeout: 20 * time.Second}}
}

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
	LeagueID   string // ör. "3946"
	LeagueName string
	Kickoff    time.Time
	Home, Away Team
	HomeScore  string
	AwayScore  string
	State      State
	StatusName string // ör. STATUS_HALFTIME
	Clock      string // ör. "67'", "90'+4'"
	Venue      string
}

func (m Match) HasTeam(teamID string) bool {
	return teamID != "" && (m.Home.ID == teamID || m.Away.ID == teamID)
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

type Event struct {
	Minute  string
	Kind    EventKind
	Label   string // ESPN'deki adı (type.text)
	TeamID  string
	Players []string // değişiklikte [0] giren, [1] çıkan
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
	Zone         string // ör. "Champions League", "Relegation"
	ZoneColor    string // ESPN'in önerdiği renk, "#81D6AC"
}

func (s Standing) GoalDiff() int { return s.GoalsFor - s.GoalsAgainst }

// Ligde tek grup olur, grup usulü turnuvalarda birden fazla.
type StandingGroup struct {
	Name string
	Rows []Standing
}

func (c *Client) Matches(ctx context.Context, day time.Time) ([]Match, error) {
	url := fmt.Sprintf("%s/site/v2/sports/soccer/%s/scoreboard?dates=%s&limit=1000",
		baseURL, allLeagues, day.Format("20060102"))

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
			LeagueID:   leagueIDFromUID(e.UID),
			LeagueName: leagueName(comp.AltGameNote),
			Kickoff:    parseTime(e.Date),
			State:      State(e.Status.Type.State),
			StatusName: e.Status.Type.Name,
			Clock:      e.Status.DisplayClock,
			Venue:      comp.Venue.FullName,
		}
		if l, ok := FindLeague(m.LeagueID); ok {
			m.LeagueName = l.Name
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
	return matches, nil
}

func (c *Client) Events(ctx context.Context, matchID string) ([]Event, error) {
	url := fmt.Sprintf("%s/site/v2/sports/soccer/%s/summary?event=%s",
		baseURL, allLeagues, matchID)

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

func (c *Client) Standings(ctx context.Context, leagueSlug string) ([]StandingGroup, error) {
	url := fmt.Sprintf("%s/v2/sports/soccer/%s/standings", baseURL, leagueSlug)

	var r standingsJSON
	if err := c.get(ctx, url, &r); err != nil {
		return nil, err
	}

	groups := make([]StandingGroup, 0, len(r.Children))
	for _, child := range r.Children {
		g := StandingGroup{Name: child.Name}
		for _, e := range child.Standings.Entries {
			s := Standing{Team: e.Team.toTeam(), Zone: e.Note.Description}
			if c := strings.TrimLeft(e.Note.Color, "#"); c != "" {
				s.ZoneColor = "#" + c // bazen "##B5E7CE" geliyor
			}
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
			g.Rows = append(g.Rows, s)
		}
		sort.SliceStable(g.Rows, func(i, j int) bool { return g.Rows[i].Rank < g.Rows[j].Rank })
		groups = append(groups, g)
	}
	if len(groups) == 0 {
		return nil, fmt.Errorf("puan durumu bulunamadı")
	}
	return groups, nil
}

func (c *Client) Teams(ctx context.Context, leagueSlug string) ([]Team, error) {
	url := fmt.Sprintf("%s/site/v2/sports/soccer/%s/teams", baseURL, leagueSlug)

	var r teamsJSON
	if err := c.get(ctx, url, &r); err != nil {
		return nil, err
	}

	var teams []Team
	for _, s := range r.Sports {
		for _, l := range s.Leagues {
			for _, t := range l.Teams {
				teams = append(teams, t.Team.toTeam())
			}
		}
	}
	return teams, nil
}

func (c *Client) get(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "terminalscore")

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

// "s:600~l:3946~e:401879276" -> "3946"
func leagueIDFromUID(uid string) string {
	for _, part := range strings.Split(uid, "~") {
		if id, ok := strings.CutPrefix(part, "l:"); ok {
			return id
		}
	}
	return ""
}

// "UEFA Europa League, League Phase" -> "UEFA Europa League"
func leagueName(note string) string {
	name, _, _ := strings.Cut(note, ", ")
	return name
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

// tarih bazen saniyesiz geliyor: "2026-10-09T17:00Z"
func parseTime(s string) time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04Z07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
