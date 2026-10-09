// Package config, kullanıcı ayarlarını (favori takım) diskte saklar.
//
// Dosya yeri işletim sistemine göre değişir:
//   - Linux:   ~/.config/terminalscore/config.json
//   - Windows: %AppData%\terminalscore\config.json
package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type Config struct {
	// SetupDone, ilk açılıştaki favori takım sorusunun cevaplandığını
	// (ya da atlandığını) gösterir; bir daha sorulmaz.
	SetupDone    bool          `json:"setup_done"`
	FavoriteTeam *FavoriteTeam `json:"favorite_team,omitempty"`
}

type FavoriteTeam struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	LeagueSlug string `json:"league"` // takımın ulusal ligi, ör. "tur.1"
}

// TeamID, favori takım yoksa boş döner; böylece çağıranlar nil kontrolü yapmaz.
func (c Config) TeamID() string {
	if c.FavoriteTeam == nil {
		return ""
	}
	return c.FavoriteTeam.ID
}

func (c Config) LeagueSlug() string {
	if c.FavoriteTeam == nil {
		return ""
	}
	return c.FavoriteTeam.LeagueSlug
}

func path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "terminalscore", "config.json"), nil
}

// Load, ayar dosyasını okur. Dosya yoksa (ilk açılış) boş ayar döner.
func Load() (Config, error) {
	var c Config
	p, err := path()
	if err != nil {
		return c, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(data, &c)
	return c, err
}

func (c Config) Save() error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
