package monsters

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"os"
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
	allPacks []Pack    // everything loaded
	packs    []Pack    // the packs in use (see UsePacks)
	monsters []Monster // monsters from the packs in use
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
	allPacks = loaded
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
		p, errs := loadPack(fsys, e.Name())
		if len(errs) > 0 {
			return nil, fmt.Errorf("pack %q: %w", e.Name(), errors.Join(errs...))
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

// loadPack reads one pack directory. Files that can't be read are
// skipped and reported, so one bad monster doesn't sink the whole pack.
func loadPack(fsys fs.FS, dir string) (Pack, []error) {
	var p Pack
	var errs []error
	raw, err := fs.ReadFile(fsys, path.Join(dir, "pack.json"))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		p.Name = dir // pack.json is optional for community packs
	case err != nil:
		return p, []error{err}
	default:
		if err := json.Unmarshal(raw, &p); err != nil {
			return p, []error{fmt.Errorf("pack.json: %w", err)}
		}
	}
	if p.ID == "" {
		p.ID = dir
	}

	files, err := fs.Glob(fsys, path.Join(dir, "*.json"))
	if err != nil {
		return p, []error{err}
	}
	byID := map[string]Monster{}
	for _, f := range files {
		if path.Base(f) == "pack.json" {
			continue
		}
		raw, err := fs.ReadFile(fsys, f)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		var m Monster
		if err := json.Unmarshal(raw, &m); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path.Base(f), err))
			continue
		}
		if m.ID == "" {
			m.ID = strings.TrimSuffix(path.Base(f), ".json")
		}
		if art, err := fs.ReadFile(fsys, path.Join(dir, m.ID+".txt")); err == nil {
			// Normalise Windows line endings; a stray \r would garble the layout.
			text := strings.ReplaceAll(string(art), "\r\n", "\n")
			m.ASCII = strings.TrimRight(strings.ReplaceAll(text, "\r", ""), "\n")
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
	return p, errs
}

// GetAllMonsters returns all available monsters
func GetAllMonsters() []Monster {
	return monsters
}

// AllPacks returns every loaded pack, ignoring any --pack filter.
func AllPacks() []Pack {
	return allPacks
}

// EveryMonster returns every loaded monster, ignoring any --pack filter.
func EveryMonster() []Monster {
	var out []Monster
	for _, p := range allPacks {
		out = append(out, p.Monsters...)
	}
	return out
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

// RandomFact picks a random monster and one of its facts using r.
func RandomFact(r *rand.Rand) (Monster, string) {
	m := monsters[r.Intn(len(monsters))]
	return m, m.Facts[r.Intn(len(m.Facts))]
}

// LoadUserPacks reads community packs from dir, where each subdirectory is
// a pack. Broken or incomplete monsters are skipped and reported; a missing
// dir simply means no community packs.
func LoadUserPacks(dir string) ([]Pack, []error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, []error{err}
	}
	fsys := os.DirFS(dir)
	var out []Pack
	var errs []error
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, perrs := loadPack(fsys, e.Name())
		for _, err := range perrs {
			errs = append(errs, fmt.Errorf("pack %q: %w", e.Name(), err))
		}
		p.Source = dir + string(os.PathSeparator) + e.Name()
		var ok []Monster
		for _, m := range p.Monsters {
			if problems := Validate(m); len(problems) > 0 {
				errs = append(errs, fmt.Errorf("pack %q: skipping %q: %s", p.ID, m.ID, strings.Join(problems, "; ")))
				continue
			}
			applyDefaults(&m)
			ok = append(ok, m)
		}
		p.Monsters = ok
		if len(ok) > 0 {
			out = append(out, p)
		}
	}
	return out, errs
}

// Validate lists what a community monster is missing to be playable.
func Validate(m Monster) []string {
	var problems []string
	if strings.TrimSpace(m.Name) == "" {
		problems = append(problems, "needs a name")
	}
	if strings.TrimSpace(m.Description) == "" {
		problems = append(problems, "needs a description")
	}
	if len(m.Facts) == 0 {
		problems = append(problems, "needs at least one fact")
	}
	for _, v := range []int{m.Stats.Strength, m.Stats.Speed, m.Stats.Cunning, m.Stats.Dread} {
		if v < 0 || v > 10 {
			problems = append(problems, "stats must be between 1 and 10")
			break
		}
	}
	return problems
}

// applyDefaults fills in optional fields so community monsters work everywhere.
func applyDefaults(m *Monster) {
	if m.Emoji == "" {
		m.Emoji = "👹"
	}
	if m.Origin == "" {
		m.Origin = "Unknown"
	}
	if m.Debut.Title == "" {
		m.Debut.Title = m.Name
	}
	if m.Debut.Year == 0 && m.Debut.Era == "" {
		m.Debut.Era = "date unknown"
	}
	for _, s := range []*int{&m.Stats.Strength, &m.Stats.Speed, &m.Stats.Cunning, &m.Stats.Dread} {
		if *s == 0 {
			*s = 5
		}
	}
	if len(m.Powers) == 0 {
		m.Powers = []string{"Sheer terror"}
	}
	if len(m.Weaknesses) == 0 {
		m.Weaknesses = []string{"A good night's sleep"}
	}
}

// AddPacks merges extra packs into the collection. Packs or monsters whose
// ids are already taken are skipped and reported.
func AddPacks(extra []Pack) []error {
	var errs []error
	packIDs := map[string]bool{}
	monsterIDs := map[string]bool{}
	for _, p := range allPacks {
		packIDs[p.ID] = true
		for _, m := range p.Monsters {
			monsterIDs[m.ID] = true
		}
	}
	merged := append([]Pack{}, allPacks...)
	for _, p := range extra {
		if packIDs[p.ID] {
			errs = append(errs, fmt.Errorf("pack %q: a pack with that id already exists", p.ID))
			continue
		}
		var ok []Monster
		for _, m := range p.Monsters {
			if monsterIDs[m.ID] {
				errs = append(errs, fmt.Errorf("pack %q: skipping %q: a monster with that id already exists", p.ID, m.ID))
				continue
			}
			monsterIDs[m.ID] = true
			ok = append(ok, m)
		}
		if len(ok) == 0 {
			continue
		}
		p.Monsters = ok
		packIDs[p.ID] = true
		merged = append(merged, p)
	}
	allPacks = merged
	setPacks(merged)
	return errs
}

// UsePacks limits every command to the named packs.
func UsePacks(ids []string) error {
	var keep []Pack
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			continue // --pack a,a means a, not every monster twice
		}
		seen[id] = true
		found := false
		for _, p := range allPacks {
			if p.ID == id {
				keep = append(keep, p)
				found = true
			}
		}
		if !found {
			var names []string
			for _, p := range allPacks {
				names = append(names, p.ID)
			}
			return fmt.Errorf("no pack called %q; choose from: %s", id, strings.Join(names, ", "))
		}
	}
	setPacks(keep)
	return nil
}
