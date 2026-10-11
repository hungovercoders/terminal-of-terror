package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/hungovercoders/terminal-of-terror/internal/crypt"
	"github.com/hungovercoders/terminal-of-terror/internal/mash"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var mashFast bool

// mashJSON is a bout as --json prints it: who fought, round by round,
// and who won.
type mashJSON struct {
	A      fighterJSON  `json:"a"`
	B      fighterJSON  `json:"b"`
	Rounds []mash.Round `json:"rounds"`
	Winner string       `json:"winner"` // the winner's id
	Finale string       `json:"finale"`
}

type fighterJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var mashCmd = &cobra.Command{
	Use:   "mash [monster] [monster]",
	Short: "Pit two monsters against each other in the Monster Mash",
	Long: `Stage a three-round bout between two monsters. Each round tests strength,
speed, cunning or dread, using each monster's stat card plus a roll of
the dice. Leave out one or both names for random contenders.`,
	Example: `  terminal-of-terror mash
  terminal-of-terror mash dracula
  terminal-of-terror mash "wolf man" mummy`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		r := newRand()
		var fighters []monsters.Monster
		for _, a := range args {
			m, err := resolveMonster(a)
			if err != nil {
				return err
			}
			fighters = append(fighters, *m)
		}
		if len(fighters) == 2 && fighters[0].ID == fighters[1].ID {
			return fail(codeUsage, "pick two different monsters", "%s can't fight itself...", fighters[0].Name)
		}
		// Fill empty corners from monsters not already fighting. Picking from
		// this pool (not retrying at random) can't loop forever.
		for len(fighters) < 2 {
			var pool []monsters.Monster
			for _, m := range monsters.GetAllMonsters() {
				if len(fighters) == 0 || m.ID != fighters[0].ID {
					pool = append(pool, m)
				}
			}
			if len(pool) == 0 {
				return fail(codeUsage, "try a different --pack", "the Monster Mash needs at least two different monsters")
			}
			fighters = append(fighters, pool[r.Intn(len(pool))])
		}

		bout := mash.Fight(r, fighters[0], fighters[1])
		if jsonOut {
			win := bout.WinnerMonster()
			return printJSON(mashJSON{
				A: fighterJSON{bout.A.ID, bout.A.Name}, B: fighterJSON{bout.B.ID, bout.B.Name},
				Rounds: bout.Rounds, Winner: win.ID, Finale: bout.Finale,
			})
		}
		w := ui.TerminalWidth()
		pause := func(d time.Duration) {
			if !mashFast && term.IsTerminal(os.Stdout.Fd()) {
				time.Sleep(d)
			}
		}
		fmt.Println(ui.MashHeader(bout.A, bout.B, w))
		pause(1200 * time.Millisecond)
		for _, round := range bout.Rounds {
			fmt.Println(ui.MashRound(bout, round, w))
			pause(1500 * time.Millisecond)
		}
		fmt.Print(ui.MashResult(bout, w))
		record(crypt.Outcome{Kind: "mash"})
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mashCmd)
	mashCmd.Flags().BoolVar(&mashFast, "fast", false, "Skip the dramatic pauses")
}
