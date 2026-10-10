package espn

import (
	"encoding/json"
	"testing"
)

func TestTextStripsControlChars(t *testing.T) {
	var v struct {
		Name  text `json:"name"`
		Score text `json:"score"`
	}
	// \u001b]0;x\u0007 terminal başlığını değiştiren bir kaçış dizisi
	in := `{"name": "Gala\u001b]0;pwned\u0007tasaray\u009b31m", "score": 3}`
	if err := json.Unmarshal([]byte(in), &v); err != nil {
		t.Fatal(err)
	}
	if v.Name != "Gala]0;pwnedtasaray31m" {
		t.Errorf("name = %q", v.Name)
	}
	if v.Score != "3" {
		t.Errorf("score = %q", v.Score)
	}
}

func TestLeagueName(t *testing.T) {
	if got := leagueName("UEFA Europa League, League Phase"); got != "UEFA Europa League" {
		t.Errorf("got %q", got)
	}
}

func TestLeagueIDFromUID(t *testing.T) {
	if got := leagueIDFromUID("s:600~l:3946~e:401879276"); got != "3946" {
		t.Errorf("got %q", got)
	}
}

func TestIsHexColor(t *testing.T) {
	for in, want := range map[string]bool{"81D6AC": true, "b5e7ce": true, "#81D6AC": false, "81D6A": false, "zzzzzz": false} {
		if isHexColor(in) != want {
			t.Errorf("isHexColor(%q) != %v", in, want)
		}
	}
}
