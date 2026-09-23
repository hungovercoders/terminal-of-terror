package cmd

import (
	"fmt"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var listJSON bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available monsters",
	Long:  `Display a list of all the monsters lurking in the Terminal of Terror, grouped by pack.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all := monsters.GetAllMonsters()
		if listJSON {
			return printJSON(all)
		}
		fmt.Print(ui.RenderList(all, ui.TerminalWidth()))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output the full monster data as JSON")
}
