package cmd

import (
	"fmt"
	"strings"

	"github.com/hungovercoders/terminal-of-terror/internal/crypt"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var (
	showAll     bool
	noIntro     bool
	monsterJSON bool
)

var monsterCmd = &cobra.Command{
	Use:   "monster [name]",
	Short: "Explore universal monsters interactively",
	Long: `Display universal monsters with their descriptions, facts, and origins.
Use arrow keys or h/l to navigate between monsters.

Pass a name to jump straight to a monster. Partial names and nicknames
work too, e.g. "dracula", "wolfman", "gill-man" or "quasimodo".`,
	Example: `  terminal-of-terror monster
  terminal-of-terror monster dracula
  terminal-of-terror monster "the mummy"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		startID := ""
		if len(args) == 1 {
			m, err := resolveMonster(args[0])
			if err != nil {
				return err
			}
			if monsterJSON {
				return printJSON(m)
			}
			startID = m.ID
		} else if monsterJSON {
			return printJSON(monsters.GetAllMonsters())
		}
		seen, err := ui.RunUI(ui.Options{ShowAll: showAll, StartID: startID, NoIntro: noIntro, Seed: seed()})
		if err != nil {
			return err
		}
		record(crypt.Outcome{Kind: "explore", Seen: seen})
		return nil
	},
}

// resolveMonster turns a user-supplied name into a monster, with a helpful
// error listing candidates when the name is ambiguous or unknown.
func resolveMonster(query string) (*monsters.Monster, error) {
	m, candidates := monsters.Find(query)
	if m != nil {
		return m, nil
	}
	if len(candidates) > 1 {
		names := make([]string, len(candidates))
		for i, c := range candidates {
			names[i] = c.Name
		}
		return nil, fmt.Errorf("%q could be any of: %s", query, strings.Join(names, ", "))
	}
	return nil, fmt.Errorf("no monster called %q lurks here; try 'terminal-of-terror list'", query)
}

func init() {
	rootCmd.AddCommand(monsterCmd)
	monsterCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Open on the gallery of every monster")
	monsterCmd.Flags().BoolVar(&noIntro, "no-intro", false, "Skip the Channel 13 opening")
	monsterCmd.Flags().BoolVar(&monsterJSON, "json", false, "Print the monster's data as JSON instead of opening the explorer")
}
