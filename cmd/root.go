package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/store"
	"github.com/spf13/cobra"
)

var packFilter []string

var rootCmd = &cobra.Command{
	Use:   "terminal-of-terror",
	Short: "A terminal tool that terrifies you with universal monsters!",
	Long: `Terminal of Terror brings classic Universal monsters and creatures of world
folklore to your terminal, presented by your late-night horror host,
Count Cathode, live on Channel 13's Creature Feature.

Explore the real history behind Dracula, Frankenstein's Monster, the Wolf
Man and more, then test yourself in the Midnight Quiz, capture monsters for
your crypt, and check tonight's double feature.

Start here:  terminal-of-terror monster`,
	Version: appVersion(),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		loadCommunityPacks()
		if len(packFilter) > 0 {
			return monsters.UsePacks(packFilter)
		}
		return nil
	},
}

// loadCommunityPacks adds any packs from the user's config directory,
// warning about (but skipping) anything broken.
func loadCommunityPacks() {
	dir, err := store.Dir()
	if err != nil {
		return
	}
	loaded, errs := monsters.LoadUserPacks(filepath.Join(dir, "packs"))
	errs = append(errs, monsters.AddPacks(loaded)...)
	for _, e := range errs {
		fmt.Fprintln(os.Stderr, "⚠️  community pack:", e)
	}
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// Runtime errors (like an unknown monster) are explained in the error
	// itself, so don't bury them under the usage text.
	rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().StringSliceVar(&packFilter, "pack", nil, "Only use monsters from these packs, e.g. --pack folklore")
}
