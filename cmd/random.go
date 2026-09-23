package cmd

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hungovercoders/terminal-of-terror/internal/host"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var randomJSON bool

// factJSON is the machine-readable shape of a single fact.
type factJSON struct {
	ID      string `json:"id"`
	Monster string `json:"monster"`
	Fact    string `json:"fact"`
}

var randomCmd = &cobra.Command{
	Use:   "random",
	Short: "Get a random monster fact",
	Long:  `Display a random fact about a random monster, introduced by your host, Count Cathode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		m, fact := monsters.RandomFact(r)
		if randomJSON {
			return printJSON(factJSON{ID: m.ID, Monster: m.Name, Fact: fact})
		}
		card := ui.FactCard{Title: "RANDOM TERROR FACT", Monster: m, Fact: fact, Host: host.Quip(r)}
		fmt.Print(card.Render(ui.TerminalWidth()))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(randomCmd)
	randomCmd.Flags().BoolVar(&randomJSON, "json", false, "Output the fact as JSON")
}
