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
	tonightDate    string
	tonightShuffle bool
)

var tonightCmd = &cobra.Command{
	Use:   "tonight",
	Short: "See tonight's Creature Feature double bill",
	Long: `Print a ticket for tonight's Channel 13 double feature. The line-up changes
every night, so check back tomorrow, or use --shuffle for a different bill.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		now, err := calendar.Date(tonightDate)
		if err != nil {
			return err
		}
		all := monsters.GetAllMonsters()
		if len(all) < 2 {
			return fmt.Errorf("a double feature needs at least two monsters")
		}
		s := calendar.DailySeed(now) * 13
		if tonightShuffle {
			s = newRand().Int63()
		}
		r := rand.New(rand.NewSource(s))
		pick := r.Perm(len(all))
		first, second := all[pick[0]], all[pick[1]]

		w := ui.TerminalWidth()
		fmt.Print(ui.RenderTicket(now, first, second, w))
		card := ui.FactCard{
			Title:   "FROM THE HOST",
			Monster: first,
			Fact:    host.Intro(first),
			Extras:  calendar.Notes(now, all),
		}
		fmt.Print(card.Render(w))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tonightCmd)
	tonightCmd.Flags().StringVar(&tonightDate, "date", "", "See the bill for another night (YYYY-MM-DD)")
	tonightCmd.Flags().BoolVar(&tonightShuffle, "shuffle", false, "Pick a random double bill instead of tonight's")
}
