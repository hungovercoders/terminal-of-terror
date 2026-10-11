package monsters

import (
	"sort"
	"strings"
	"unicode/utf8"
)

// maxSuggestions bounds a "did you mean" list: enough to hold the likely
// intended name, few enough to read in one line of error.
const maxSuggestions = 3

// Suggest returns the names of the monsters whose id, name or alias is
// within a few edits of query, nearest first, for a typo the substring
// match in Find cannot catch ("draculla", "frankenstien"). Nothing close
// gives nil: a wrong name should not get a guess for the sake of it.
func Suggest(query string) []string {
	type candidate struct {
		key, name string
	}
	var cands []candidate
	for _, m := range monsters {
		keys := append([]string{m.ID, m.Name}, m.Aliases...)
		for _, k := range keys {
			cands = append(cands, candidate{k, m.Name})
		}
	}
	keys := make([]string, len(cands))
	for i, c := range cands {
		keys[i] = c.key
	}
	var names []string
	seen := map[string]bool{}
	for _, k := range closest(query, keys) {
		for _, c := range cands {
			if c.key == k && !seen[c.name] {
				seen[c.name] = true
				names = append(names, c.name)
			}
		}
	}
	if len(names) > maxSuggestions {
		names = names[:maxSuggestions]
	}
	return names
}

// closest ranks candidates within a few edits of query (one edit for a
// short query, up to three for a long one), nearest first. Comparison uses
// normalize, so case, punctuation and a leading "the" do not count. A query
// of one letter gets no guess: it is inside half the names.
func closest(query string, candidates []string) []string {
	q := normalize(query)
	n := utf8.RuneCountInString(q)
	if n < 2 {
		return nil
	}
	limit := min(1+n/5, 3)
	type scored struct {
		name string
		dist int
	}
	var found []scored
	seen := map[string]bool{}
	for _, c := range candidates {
		if seen[c] {
			continue
		}
		seen[c] = true
		if d := levenshtein(q, normalize(c)); d <= limit && d < n {
			found = append(found, scored{c, d})
		}
	}
	sort.Slice(found, func(i, j int) bool {
		if found[i].dist != found[j].dist {
			return found[i].dist < found[j].dist
		}
		return found[i].name < found[j].name
	})
	out := make([]string, 0, len(found))
	for _, s := range found {
		out = append(out, s.name)
	}
	return out
}

// DidYouMean phrases suggestions for an error: "Did you mean Dracula?",
// "Did you mean Dracula or The Mummy?", or "" for none.
func DidYouMean(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return "Did you mean " + names[0] + "?"
	}
	return "Did you mean " + strings.Join(names[:len(names)-1], ", ") + " or " + names[len(names)-1] + "?"
}

// levenshtein is the edit distance between a and b, in runes.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}
