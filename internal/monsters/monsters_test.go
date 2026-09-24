package monsters

import (
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
)

var hexColor = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
var isoDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// TestMonsterData guards the quality of every built-in monster.
func TestMonsterData(t *testing.T) {
	all := GetAllMonsters()
	if len(all) < 8 {
		t.Fatalf("expected at least 8 monsters, got %d", len(all))
	}
	seen := map[string]bool{}
	for _, m := range all {
		t.Run(m.ID, func(t *testing.T) {
			if seen[m.ID] {
				t.Errorf("duplicate id %q", m.ID)
			}
			seen[m.ID] = true

			for field, v := range map[string]string{
				"id": m.ID, "name": m.Name, "emoji": m.Emoji, "description": m.Description,
				"origin": m.Origin, "legend": m.Legend, "pack": m.Pack,
				"debut.medium": m.Debut.Medium, "debut.title": m.Debut.Title, "ascii": m.ASCII,
			} {
				if strings.TrimSpace(v) == "" {
					t.Errorf("%s is empty", field)
				}
			}
			for _, r := range m.ASCII {
				// Wide or right-to-left characters break alignment and the fog in Guess the Monster.
				if r > 0x2FFF || (r >= 0x0590 && r <= 0x08FF) {
					t.Errorf("ASCII art uses %q; stick to single-width, left-to-right characters", r)
					break
				}
			}
			if m.Debut.Year == 0 && m.Debut.Era == "" {
				t.Error("debut needs a year or an era")
			}
			if len(m.Facts) < 5 {
				t.Errorf("want at least 5 facts, got %d", len(m.Facts))
			}
			if len(m.Myths) < 3 {
				t.Errorf("want at least 3 myth checks, got %d", len(m.Myths))
			}
			for _, my := range m.Myths {
				if my.Claim == "" || my.Explanation == "" {
					t.Errorf("incomplete myth check: %+v", my)
				}
			}
			for _, q := range m.Quotes {
				if q.Text == "" || q.Speaker == "" || q.Source == "" {
					t.Errorf("incomplete quote: %+v", q)
				}
			}
			if len(m.Powers) < 2 || len(m.Weaknesses) < 2 {
				t.Error("want at least 2 powers and 2 weaknesses")
			}
			for name, s := range map[string]int{
				"strength": m.Stats.Strength, "speed": m.Stats.Speed,
				"cunning": m.Stats.Cunning, "dread": m.Stats.Dread,
			} {
				if s < 1 || s > 10 {
					t.Errorf("stat %s = %d, want 1-10", name, s)
				}
			}
			if !hexColor.MatchString(m.Theme.Primary) || !hexColor.MatchString(m.Theme.Accent) {
				t.Errorf("theme colours must be #RRGGBB: %+v", m.Theme)
			}
			if f := m.Film; f != nil {
				if f.Title == "" || f.Year == 0 || f.Director == "" || f.Star == "" {
					t.Errorf("incomplete film: %+v", f)
				}
				if f.ReleaseDate != "" && !isoDate.MatchString(f.ReleaseDate) {
					t.Errorf("releaseDate %q must be YYYY-MM-DD", f.ReleaseDate)
				}
			}
		})
	}
}

func TestFind(t *testing.T) {
	cases := map[string]string{
		"dracula":       "dracula",
		"Count Dracula": "dracula",
		"the wolf-man":  "wolf-man",
		"wolfman":       "wolf-man",
		"FRANKENSTEIN":  "frankenstein",
		"gill-man":      "creature",
		"quasi":         "hunchback",
		"invisible":     "invisible-man",
		"Imhotep":       "mummy",
		"Opera":         "phantom",
	}
	for q, want := range cases {
		m, _ := Find(q)
		if m == nil {
			t.Errorf("Find(%q) = nil, want %s", q, want)
			continue
		}
		if m.ID != want {
			t.Errorf("Find(%q) = %s, want %s", q, m.ID, want)
		}
	}
	if m, _ := Find("zombie accountant"); m != nil {
		t.Errorf("Find(nonsense) = %s, want nil", m.ID)
	}
}

func TestFirstAppearance(t *testing.T) {
	m := GetMonsterByName("dracula")
	if got, want := m.FirstAppearance(), "Dracula by Bram Stoker (1897 novel)"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLoadPacksOrderAndArt(t *testing.T) {
	fsys := fstest.MapFS{
		"test/pack.json": {Data: []byte(`{"id":"test","name":"Test","order":["b","a"]}`)},
		"test/a.json":    {Data: []byte(`{"name":"A"}`)},
		"test/b.json":    {Data: []byte(`{"name":"B"}`)},
		"test/c.json":    {Data: []byte(`{"name":"C"}`)},
		"test/b.txt":     {Data: []byte("art\n\n")},
	}
	p, err := LoadPacks(fsys, "test")
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, m := range p[0].Monsters {
		ids = append(ids, m.ID)
	}
	if got := strings.Join(ids, ","); got != "b,a,c" {
		t.Errorf("order = %s, want b,a,c", got)
	}
	if p[0].Monsters[0].ASCII != "art" || p[0].Monsters[0].Pack != "test" {
		t.Errorf("art/pack not loaded: %+v", p[0].Monsters[0])
	}
}
