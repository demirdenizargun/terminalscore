// Package config: ~/.config/terminalscore/config.json (Windows'ta %AppData% altında).
package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type Config struct {
	SetupDone    bool          `json:"setup_done"` // favori takım soruldu (seçildi ya da atlandı)
	FavoriteTeam *FavoriteTeam `json:"favorite_team,omitempty"`
	Language     string        `json:"language,omitempty"` // "tr", "en"; boşsa sistem dili
}

type FavoriteTeam struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	LeagueSlug string `json:"league"` // ör. "tur.1"
}

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

// Dosya yoksa (ilk açılış) boş Config döner.
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
