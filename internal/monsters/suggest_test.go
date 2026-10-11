package monsters

import (
	"reflect"
	"testing"
)

func TestSuggestCatchesTypos(t *testing.T) {
	cases := map[string][]string{
		"draculla":     {"Dracula"},
		"frankenstien": {"Frankenstein's Monster"},
		"Wolfmen":      {"The Wolf Man"}, // one edit from the "wolfman" alias
		"mumy":         {"The Mummy"},
		"zombie":       nil, // nothing within reach
		"x":            nil, // one letter is inside half the names
		"":             nil,
	}
	for q, want := range cases {
		if got := Suggest(q); !reflect.DeepEqual(got, want) {
			t.Errorf("Suggest(%q) = %v, want %v", q, got, want)
		}
	}
}

func TestClosestRanksAndCaps(t *testing.T) {
	got := closest("folklor", []string{"folklore", "cryptids", "universal", "folklore"})
	if want := []string{"folklore"}; !reflect.DeepEqual(got, want) {
		t.Errorf("closest(folklor) = %v, want %v", got, want)
	}
	if got := Suggest("ma"); len(got) > maxSuggestions {
		t.Errorf("Suggest(ma) = %v, more than %d", got, maxSuggestions)
	}
}

func TestDidYouMean(t *testing.T) {
	cases := map[string][]string{
		"":                        nil,
		"Did you mean A?":         {"A"},
		"Did you mean A or B?":    {"A", "B"},
		"Did you mean A, B or C?": {"A", "B", "C"},
	}
	for want, names := range cases {
		if got := DidYouMean(names); got != want {
			t.Errorf("DidYouMean(%v) = %q, want %q", names, got, want)
		}
	}
}

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{{"", "", 0}, {"", "abc", 3}, {"kitten", "sitting", 3}, {"flaw", "lawn", 2}, {"héllo", "hello", 1}}
	for _, c := range cases {
		if got := levenshtein(c.a, c.b); got != c.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
