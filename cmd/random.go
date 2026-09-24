package cmd

import (
	"fmt"
	"math/rand"

	"github.com/hungovercoders/terminal-of-terror/internal/calendar"
	"github.com/hungovercoders/terminal-of-terror/internal/host"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var (
	randomJSON  bool
	randomDaily bool
	randomDate  string
)

// factJSON is the machine-readable shape of a single fact.
type factJSON struct {
	ID      string   `json:"id"`
	Monster string   `json:"monster"`
	Fact    string   `json:"fact"`
	Notes   []string `json:"notes,omitempty"`
}

var randomCmd = &cobra.Command{
	Use:   "random",
	Short: "Get a random monster fact",
	Long: `Display a random fact about a random monster, introduced by your host,
Count Cathode.

With --daily you get the Fact of the Night: the same for everyone all day,
and a new one tomorrow. It's perfect for your shell's startup file.`,
	Example: `  terminal-of-terror random
  terminal-of-terror random --daily
  echo 'terminal-of-terror random --daily' >> ~/.bashrc`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		now, err := calendar.Date(randomDate)
		if err != nil {
			return err
		}
		s := seed()
		title := "RANDOM TERROR FACT"
		if randomDaily || randomDate != "" {
			s = calendar.DailySeed(now)
			title = "FACT OF THE NIGHT · " + now.Format("Monday 2 January")
		}
		r := rand.New(rand.NewSource(s))
		m, fact := monsters.RandomFact(r)
		notes := calendar.Notes(now, monsters.GetAllMonsters())
		if randomJSON {
			return printJSON(factJSON{ID: m.ID, Monster: m.Name, Fact: fact, Notes: notes})
		}
		card := ui.FactCard{Title: title, Monster: m, Fact: fact, Host: host.Quip(r), Extras: notes}
		fmt.Print(card.Render(ui.TerminalWidth()))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(randomCmd)
	randomCmd.Flags().BoolVar(&randomJSON, "json", false, "Output the fact as JSON")
	randomCmd.Flags().BoolVarP(&randomDaily, "daily", "d", false, "Show the Fact of the Night (the same all day)")
	randomCmd.Flags().StringVar(&randomDate, "date", "", "Show the Fact of the Night for another date (YYYY-MM-DD)")
}
