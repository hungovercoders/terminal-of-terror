package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "terminal-of-terror",
	Short: "A terminal tool that terrifies you with universal monsters!",
	Long: `Terminal of Terror is a CLI application that brings classic 
universal monsters to your terminal. Explore famous monsters like 
Dracula, Frankenstein's Monster, The Wolf Man, and more!

Learn fascinating and terrifying facts about these legendary creatures 
from horror history.`,
	Version: "1.0.0",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
