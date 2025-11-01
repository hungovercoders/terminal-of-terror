package cmd

import (
	"fmt"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available monsters",
	Long:  `Display a list of all universal monsters available in the Terminal of Terror.`,
	Run: func(cmd *cobra.Command, args []string) {
		allMonsters := monsters.GetAllMonsters()
		fmt.Println("Universal Monsters:")
		fmt.Println("==================")
		for i, monster := range allMonsters {
			fmt.Printf("%d. %s - %s\n", i+1, monster.Name, monster.Description)
		}
		fmt.Printf("\nTotal: %d monsters\n", len(allMonsters))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
