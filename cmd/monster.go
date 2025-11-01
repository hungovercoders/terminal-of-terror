package cmd

import (
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var showAll bool

var monsterCmd = &cobra.Command{
	Use:   "monster",
	Short: "Explore universal monsters interactively",
	Long: `Display universal monsters with their descriptions, facts, and origins.
Use arrow keys or h/l to navigate between monsters.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ui.RunUI(showAll)
	},
}

func init() {
	rootCmd.AddCommand(monsterCmd)
	monsterCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all monsters at once")
}
