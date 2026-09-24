package cmd

import (
	"fmt"

	"github.com/hungovercoders/terminal-of-terror/internal/calendar"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var countdownDate string

var countdownCmd = &cobra.Command{
	Use:   "countdown",
	Short: "Count down the nights until Halloween",
	Long: `Show how many nights are left until Halloween, tonight's moon and the next
classic film anniversary. During October, each night of the 31 Nights of
Fright features a different monster.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		now, err := calendar.Date(countdownDate)
		if err != nil {
			return err
		}
		fmt.Print(ui.RenderCountdown(now, monsters.GetAllMonsters(), ui.TerminalWidth()))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(countdownCmd)
	countdownCmd.Flags().StringVar(&countdownDate, "date", "", "Pretend it's this date (YYYY-MM-DD)")
}
