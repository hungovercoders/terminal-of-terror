package cmd

import (
	"fmt"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/spf13/cobra"
)

var randomCmd = &cobra.Command{
	Use:   "random",
	Short: "Get a random monster fact",
	Long:  `Display a random fact about a random universal monster.`,
	Run: func(cmd *cobra.Command, args []string) {
		monsterName, fact := monsters.GetRandomFact()
		fmt.Printf("🎃 Random Terror Fact 🎃\n\n")
		fmt.Printf("Monster: %s\n", monsterName)
		fmt.Printf("Fact: %s\n", fact)
	},
}

func init() {
	rootCmd.AddCommand(randomCmd)
}
