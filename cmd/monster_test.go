package cmd

import (
	"strings"
	"testing"
)

func TestResolveMonsterSuggestsOnATypo(t *testing.T) {
	_, err := resolveMonster("draculla")
	if err == nil || !strings.Contains(err.Error(), "Did you mean Dracula?") {
		t.Errorf("resolveMonster(draculla) = %v, want a Dracula suggestion", err)
	}
	_, err = resolveMonster("zombie accountant")
	if err == nil || strings.Contains(err.Error(), "Did you mean") || !strings.Contains(err.Error(), "terminal-of-terror list") {
		t.Errorf("resolveMonster(nonsense) = %v, want the plain not-found error", err)
	}
	if m, err := resolveMonster("dracula"); err != nil || m.ID != "dracula" {
		t.Errorf("resolveMonster(dracula) = %v, %v", m, err)
	}
}

func TestPacksHelpNamesEveryBuiltInPack(t *testing.T) {
	for _, want := range []string{"Universal Classics", "World Folklore", "Cryptids"} {
		if !strings.Contains(packsCmd.Long, want) {
			t.Errorf("packs help does not mention %q:\n%s", want, packsCmd.Long)
		}
	}
}
