package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/store"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var (
	packFilter []string
	noColor    bool
)

var rootCmd = &cobra.Command{
	Use:   "terminal-of-terror",
	Short: "Channel 13's Creature Feature: classic monsters, folklore, cryptids and myth in your terminal",
	Long: `Terminal of Terror brings classic Universal monsters, creatures of world
folklore, cryptids, literary monsters, the beasts of Greek myth and the first
monsters ever filmed to your terminal, presented by your late-night horror
host, Count Cathode, live on Channel 13's Creature Feature.

Explore the real history behind Dracula, Frankenstein's Monster, the Wolf
Man and more, then test yourself in the Midnight Quiz, capture monsters for
your crypt, and check tonight's double feature.

Start here:  terminal-of-terror monster`,
	Version: appVersion(),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Plain output for a pipe, a log, or anyone who asked: --no-color,
		// NO_COLOR (https://no-color.org) or stdout not being a terminal.
		ui.SetPlain(noColor || os.Getenv("NO_COLOR") != "" || !isTerminal(os.Stdout))
		loadCommunityPacks()
		if len(packFilter) > 0 {
			if err := monsters.UsePacks(packFilter); err != nil {
				return fail(codeBadPack, "", "%v", err)
			}
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
		fmt.Fprintln(os.Stderr, "warning: community pack:", e)
	}
}

// Execute runs the command line. An error exits 1; under --json it is
// one JSON object on stderr rather than Cobra's "Error: ..." line.
func Execute() {
	if wantsJSON(os.Args[1:]) {
		rootCmd.SilenceErrors = true
	}
	err := rootCmd.Execute()
	if err != nil {
		if wantsJSON(os.Args[1:]) {
			writeErrorJSON(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// Runtime errors (like an unknown monster) are explained in the error
	// itself, so don't bury them under the usage text.
	rootCmd.SilenceUsage = true
	// A bad flag is Cobra's error; give it the code a script expects.
	rootCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fail(codeBadFlag, "", "%v", err)
	})
	rootCmd.PersistentFlags().StringSliceVar(&packFilter, "pack", nil, "Only use monsters from these packs, e.g. --pack folklore")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Print the command's data as JSON instead of drawing it (errors become a JSON object on stderr)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Plain output: no colour, ASCII borders, no emoji (also set by NO_COLOR, or by piping)")
}
