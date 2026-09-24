// Package store saves a player's progress (captured monsters, badges and
// scores) as JSON in their config directory.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// EnvHome overrides where progress and community packs are kept.
const EnvHome = "TERMINAL_OF_TERROR_HOME"

// Progress is everything we remember about a player.
type Progress struct {
	Version   int                  `json:"version"`
	Knowledge map[string]int       `json:"knowledge"` // correct answers per monster id
	Captured  map[string]time.Time `json:"captured"`  // monster id -> when captured
	Badges    map[string]time.Time `json:"badges"`    // badge id -> when earned
	Seen      map[string]bool      `json:"seen"`      // monster pages visited

	QuizzesPlayed  int `json:"quizzesPlayed"`
	BestQuizScore  int `json:"bestQuizScore"` // percent
	GuessesPlayed  int `json:"guessesPlayed"`
	BestGuessScore int `json:"bestGuessScore"` // points
	MashesPlayed   int `json:"mashesPlayed"`

	path string
}

// Dir is the app's config directory: $TERMINAL_OF_TERROR_HOME if set,
// otherwise <user config dir>/terminal-of-terror.
func Dir() (string, error) {
	if d := os.Getenv(EnvHome); d != "" {
		return d, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "terminal-of-terror"), nil
}

// Path is where progress is saved.
func Path() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "progress.json"), nil
}

// Load reads saved progress, returning empty progress if there is none yet.
func Load() (*Progress, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	p := &Progress{path: path}
	raw, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return nil, err
	default:
		if err := json.Unmarshal(raw, p); err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
	}
	p.init()
	return p, nil
}

func (p *Progress) init() {
	p.Version = 1
	if p.Knowledge == nil {
		p.Knowledge = map[string]int{}
	}
	if p.Captured == nil {
		p.Captured = map[string]time.Time{}
	}
	if p.Badges == nil {
		p.Badges = map[string]time.Time{}
	}
	if p.Seen == nil {
		p.Seen = map[string]bool{}
	}
}

// Save writes progress atomically, creating the directory if needed.
func (p *Progress) Save() error {
	if p.path == "" {
		path, err := Path()
		if err != nil {
			return err
		}
		p.path = path
	}
	if err := os.MkdirAll(filepath.Dir(p.path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p.path), "progress-*.json")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), p.path)
}

// Reset deletes saved progress.
func Reset() error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// New returns empty progress that saves to the default location.
func New() *Progress {
	p := &Progress{}
	p.init()
	return p
}
