package cmd

import (
	"strings"
	"testing"
)

func TestCompleteMonstersOffersIdsAndAliases(t *testing.T) {
	got, _ := completeMonsters(monsterCmd, nil, "drac")
	if len(got) == 0 || !strings.HasPrefix(got[0], "dracula\tDracula") {
		t.Errorf("completeMonsters(drac) = %v", got)
	}
	got, _ = completeMonsters(monsterCmd, nil, "wolfm")
	if len(got) != 1 || !strings.HasPrefix(got[0], "wolfman\tThe Wolf Man") {
		t.Errorf("completeMonsters(wolfm) = %v, want the alias", got)
	}
	if got, _ := completeMonsters(monsterCmd, nil, "zzz"); len(got) != 0 {
		t.Errorf("completeMonsters(zzz) = %v", got)
	}
}

func TestCompletePacksOffersIds(t *testing.T) {
	got, _ := completePacks(rootCmd, nil, "folk")
	if len(got) != 1 || !strings.HasPrefix(got[0], "folklore\t") {
		t.Errorf("completePacks(folk) = %v", got)
	}
	if rootCmd.CompletionOptions.DisableDefaultCmd {
		t.Error("the completion command should be on")
	}
}
