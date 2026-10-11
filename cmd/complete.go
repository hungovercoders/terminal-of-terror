package cmd

import (
	"strings"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/spf13/cobra"
)

// completeMonsters offers monster ids and aliases to shell completion, each
// with the monster's name beside it. Ids and aliases are single words, so
// they complete without quoting; the names themselves, with their spaces
// and apostrophes, would not.
func completeMonsters(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ensureCommunityPacks()
	prefix := strings.ToLower(toComplete)
	var out []string
	seen := map[string]bool{}
	for _, m := range monsters.GetAllMonsters() {
		for _, key := range append([]string{m.ID}, m.Aliases...) {
			key = strings.ToLower(key) // "Wolfman" in the data; lookups are case-insensitive
			if strings.HasPrefix(key, prefix) && !seen[key] {
				seen[key] = true
				out = append(out, key+"\t"+m.Name)
			}
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

// completePacks offers pack ids to --pack, built-in and community alike.
func completePacks(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ensureCommunityPacks()
	var out []string
	for _, p := range monsters.AllPacks() {
		if strings.HasPrefix(p.ID, strings.ToLower(toComplete)) {
			out = append(out, p.ID+"\t"+p.Name)
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}
