package cmd

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hungovercoders/terminal-of-terror/internal/crypt"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var guessRounds int

var guessCmd = &cobra.Command{
	Use:   "guess",
	Short: "Guess the monster as its portrait emerges from the fog",
	Long: `A monster's portrait is hidden in fog. Guess who it is, or press space to
clear more fog and get clues, at the cost of points.

Correct guesses count towards capturing monsters for your crypt.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if guessRounds < 1 {
			return fmt.Errorf("--rounds must be at least 1")
		}
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		rounds := ui.NewGuessRounds(r, monsters.GetAllMonsters(), guessRounds)
		if len(rounds) == 0 {
			return fmt.Errorf("Guess the Monster needs at least 3 different monsters, including one with a portrait; try a different --pack")
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
