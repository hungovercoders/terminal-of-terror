package monsters

import "strings"

// SearchHit is a monster matching a search, with where and why it matched.
type SearchHit struct {
	Monster Monster
	Field   string
	Snippet string
}

// Search finds monsters whose name, aliases or lore mention the query. Each
// monster appears at most once, at its best-ranked matching field.
func Search(query string) []SearchHit {
	return searchIn(monsters, query)
}

func searchIn(list []Monster, query string) []SearchHit {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	var hits []SearchHit
	for _, m := range list {
		for _, f := range searchFields(m) {
			lower := strings.ToLower(f.text)
			if i := strings.Index(lower, q); i >= 0 {
				text := f.text
				if len(lower) != len(text) {
					// Lowercasing changed byte offsets; show the lowered text rather than mis-slice.
					text = lower
				}
				hits = append(hits, SearchHit{Monster: m, Field: f.name, Snippet: snippet(text, i, len(q))})
				break
			}
		}
	}
	return hits
}

type field struct{ name, text string }

// searchFields lists what we search, most important first.
func searchFields(m Monster) []field {
	fs := []field{{"name", m.Name}}
	for _, a := range m.Aliases {
		fs = append(fs, field{"alias", a})
	}
	fs = append(fs, field{"description", m.Description}, field{"origin", m.Origin})
	if m.Film != nil {
		fs = append(fs, field{"film", m.Film.Title + " · " + m.Film.Director + " · " + m.Film.Star})
	}
	for _, f := range m.Facts {
		fs = append(fs, field{"fact", f})
	}
	fs = append(fs, field{"legend", m.Legend})
	for _, my := range m.Myths {
		fs = append(fs, field{"myth", my.Claim + " " + my.Explanation})
	}
	for _, l := range m.Legacy {
		fs = append(fs, field{"legacy", l})
	}
	return fs
}

// snippet trims long text to a window around the match.
func snippet(text string, at, n int) string {
	const pad = 30
	r := []rune(text)
	// convert byte offsets to rune offsets
	start := len([]rune(text[:at]))
	end := start + len([]rune(text[at:at+n]))
	from, to := start-pad, end+pad
	prefix, suffix := "…", "…"
	if from <= 0 {
		from, prefix = 0, ""
	}
	if to >= len(r) {
		to, suffix = len(r), ""
	}
	return prefix + string(r[from:to]) + suffix
}
