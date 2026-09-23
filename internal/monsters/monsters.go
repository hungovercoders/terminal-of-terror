package monsters

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"path"
	"sort"
	"strings"
)

// Monster is a single creature and everything we know about it.
type Monster struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Aliases     []string    `json:"aliases,omitempty"`
	Emoji       string      `json:"emoji"`
	Description string      `json:"description"`
	Origin      string      `json:"origin"`
	Debut       Debut       `json:"debut"`
	Film        *Film       `json:"film,omitempty"`
	Legend      string      `json:"legend"`
	Myths       []MythCheck `json:"myths,omitempty"`
	Quotes      []Quote     `json:"quotes,omitempty"`
	Powers      []string    `json:"powers"`
	Weaknesses  []string    `json:"weaknesses"`
	Stats       Stats       `json:"stats"`
	Legacy      []string    `json:"legacy,omitempty"`
	Facts       []string    `json:"facts"`
	HostIntro   string      `json:"hostIntro,omitempty"`
	Theme       Theme       `json:"theme"`
	ASCII       string      `json:"ascii,omitempty"`
	Pack        string      `json:"pack"`
}

// Debut describes where a monster first appeared.
type Debut struct {
	Medium  string `json:"medium"` // novel, film, folklore, ...
	Title   string `json:"title"`
	Creator string `json:"creator,omitempty"`
	Year    int    `json:"year,omitempty"`
	Era     string `json:"era,omitempty"` // for folklore without a single year
}

// Film describes a monster's classic screen outing.
type Film struct {
	Title       string `json:"title"`
	Year        int    `json:"year"`
	ReleaseDate string `json:"releaseDate,omitempty"` // YYYY-MM-DD
	Studio      string `json:"studio,omitempty"`
	Director    string `json:"director"`
	Star        string `json:"star"`
	Makeup      string `json:"makeup,omitempty"`
	Silent      bool   `json:"silent,omitempty"`
	Note        string `json:"note,omitempty"`
}

// MythCheck is a popular belief with a verdict and the real story.
type MythCheck struct {
	Claim       string `json:"claim"`
	True        bool   `json:"true"`
	Explanation string `json:"explanation"`
}

// Quote is a line from a public-domain source text.
type Quote struct {
	Text    string `json:"text"`
	Speaker string `json:"speaker"`
	Source  string `json:"source"`
}

// Stats power the Monster Mash. Each is 1-10 and purely for fun.
type Stats struct {
	Strength int `json:"strength"`
	Speed    int `json:"speed"`
	Cunning  int `json:"cunning"`
	Dread    int `json:"dread"`
}

// Theme holds a monster's colours as hex strings.
type Theme struct {
	Primary string `json:"primary"`
	Accent  string `json:"accent"`
}

// Pack is a themed collection of monsters.
type Pack struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Order       []string  `json:"order"`
	Source      string    `json:"-"` // "built-in" or a file path
	Monsters    []Monster `json:"-"`
}

// FirstAppearance renders the debut as a readable line.
func (m Monster) FirstAppearance() string {
	d := m.Debut
	s := d.Title
	if d.Creator != "" {
		s += " by " + d.Creator
	}
	switch {
	case d.Year != 0 && d.Medium != "":
		s += fmt.Sprintf(" (%d %s)", d.Year, d.Medium)
	case d.Year != 0:
		s += fmt.Sprintf(" (%d)", d.Year)
	case d.Era != "":
		s += " (" + d.Era + ")"
	}
	return s
}

// IsSilent reports whether the monster's classic film is a silent picture.
func (m Monster) IsSilent() bool {
	return m.Film != nil && m.Film.Silent
}

//go:embed packs
var builtinFS embed.FS

var (
	packs    []Pack
	monsters []Monster
)

func init() {
	sub, err := fs.Sub(builtinFS, "packs")
	if err != nil {
		log.Fatalf("Failed to open built-in packs: %v", err)
	}
	loaded, err := LoadPacks(sub, "built-in")
	if err != nil {
		log.Fatalf("Failed to load monsters data: %v", err)
	}
	setPacks(loaded)
}

func setPacks(p []Pack) {
	packs = p
	monsters = nil
	for _, pk := range packs {
		monsters = append(monsters, pk.Monsters...)
	}
}

// LoadPacks reads every pack directory in fsys. Each directory holds a
// pack.json, one <id>.json per monster and an optional <id>.txt of ASCII art.
func LoadPacks(fsys fs.FS, source string) ([]Pack, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, err
	}
	var out []Pack
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := loadPack(fsys, e.Name())
		if err != nil {
			return nil, fmt.Errorf("pack %q: %w", e.Name(), err)
		}
		p.Source = source
		out = append(out, p)
	}
	sort.SliceStable(out, func(i, j int) bool { return packRank(out[i].ID) < packRank(out[j].ID) })
	return out, nil
}

// packRank keeps the classics first; everything else follows alphabetically.
func packRank(id string) string {
	if id == "universal" {
		return "0"
	}
	return "1" + id
}

func loadPack(fsys fs.FS, dir string) (Pack, error) {
	var p Pack
	raw, err := fs.ReadFile(fsys, path.Join(dir, "pack.json"))
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, fmt.Errorf("pack.json: %w", err)
	}
	if p.ID == "" {
		p.ID = dir
	}

	files, err := fs.Glob(fsys, path.Join(dir, "*.json"))
	if err != nil {
		return p, err
	}
	byID := map[string]Monster{}
	for _, f := range files {
		if path.Base(f) == "pack.json" {
			continue
		}
		raw, err := fs.ReadFile(fsys, f)
		if err != nil {
			return p, err
		}
		var m Monster
		if err := json.Unmarshal(raw, &m); err != nil {
			return p, fmt.Errorf("%s: %w", path.Base(f), err)
		}
		if m.ID == "" {
			m.ID = strings.TrimSuffix(path.Base(f), ".json")
		}
		if art, err := fs.ReadFile(fsys, path.Join(dir, m.ID+".txt")); err == nil {
			m.ASCII = strings.TrimRight(string(art), "\n")
		}
		m.Pack = p.ID
		byID[m.ID] = m
	}

	// Listed monsters first, in the pack's chosen order, then any extras.
	for _, id := range p.Order {
		if m, ok := byID[id]; ok {
			p.Monsters = append(p.Monsters, m)
			delete(byID, id)
		}
	}
	var rest []string
	for id := range byID {
		rest = append(rest, id)
	}
	sort.Strings(rest)
	for _, id := range rest {
		p.Monsters = append(p.Monsters, byID[id])
	}
	return p, nil
}

// GetAllMonsters returns all available monsters
func GetAllMonsters() []Monster {
	return monsters
}

// GetPacks returns all loaded packs
func GetPacks() []Pack {
	return packs
}

// GetRandomMonster returns a random monster
func GetRandomMonster() Monster {
	return monsters[rand.Intn(len(monsters))]
}

// GetMonsterByName returns a monster by id, name or alias (case-insensitive)
func GetMonsterByName(name string) *Monster {
	key := normalize(name)
	for i := range monsters {
		m := &monsters[i]
		if normalize(m.ID) == key || normalize(m.Name) == key {
			return m
		}
		for _, a := range m.Aliases {
			if normalize(a) == key {
				return m
			}
		}
	}
	return nil
}

// Find looks a monster up by exact name first, then by partial match. When
// the query is ambiguous or unknown it returns nil plus any candidates.
func Find(query string) (*Monster, []Monster) {
	if m := GetMonsterByName(query); m != nil {
		return m, nil
	}
	key := normalize(query)
	if key == "" {
		return nil, nil
	}
	var matches []Monster
	for _, m := range monsters {
		hay := []string{m.ID, m.Name}
		hay = append(hay, m.Aliases...)
		for _, h := range hay {
			if strings.Contains(normalize(h), key) {
				matches = append(matches, m)
				break
			}
		}
	}
	if len(matches) == 1 {
		return &matches[0], nil
	}
	return nil, matches
}

// normalize lowercases and drops a leading "the" and punctuation so that
// "the wolf-man" and "Wolf Man" match.
func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "the ")
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// GetRandomFact returns a random fact from a random monster
func GetRandomFact() (string, string) {
	monster := GetRandomMonster()
	fact := monster.Facts[rand.Intn(len(monster.Facts))]
	return monster.Name, fact
}
