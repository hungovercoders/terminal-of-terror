package cmd

import (
	"github.com/hungovercoders/terminal-of-terror/internal/crypt"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var guessRounds int

// guessJSON is a round of Guess the Monster as --json prints it: the
// answer included, the way quiz --json gives the questions with theirs.
type guessJSON struct {
	Monster string   `json:"monster"`
	Name    string   `json:"name"`
	Options []string `json:"options"`
	Answer  int      `json:"answer"`
}

var guessCmd = &cobra.Command{
	Use:   "guess",
	Short: "Guess the monster as its portrait emerges from the fog",
	Long: `A monster's portrait is hidden in fog. Guess who it is, or press space to
clear more fog and get clues, at the cost of points.

Correct guesses count towards capturing monsters for your crypt.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if guessRounds < 1 {
			return fail(codeBadFlag, "", "--rounds must be at least 1")
		}
		r := newRand()
		rounds := ui.NewGuessRounds(r, monsters.GetAllMonsters(), guessRounds)
		if len(rounds) == 0 {
			return fail(codeUsage, "try a different --pack", "Guess the Monster needs at least 3 different monsters, including one with a portrait")
		}
		if jsonOut {
			out := make([]guessJSON, 0, len(rounds))
			for _, g := range rounds {
				out = append(out, guessJSON{Monster: g.Monster.ID, Name: g.Monster.Name, Options: g.Options, Answer: g.Answer})
			}
			return printJSON(out)
		}
		if err := requireTerminal("Guess the Monster", "play it from a shell"); err != nil {
			return err
		}
		res, err := ui.RunGuess(rounds, r)
		if err != nil {
			return err
		}
		played := 0
		for _, g := range res.Rounds {
			if g.Chosen >= 0 {
				played++
			}
		}
		if played == 0 {
			return nil
		}
		record(crypt.Outcome{
			Kind:      "guess",
			Correct:   res.CorrectBy(),
			Score:     res.Score(),
			Total:     len(res.Rounds),
			Completed: res.Completed,
			EagleEye:  res.EagleEye(),
		})
		return nil
	},
}

func init() {
	rootCmd.AddCommand(guessCmd)
	guessCmd.Flags().IntVarP(&guessRounds, "rounds", "n", 5, "Number of rounds")
}
