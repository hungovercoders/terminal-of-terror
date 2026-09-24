package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/store"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var (
	cryptJSON bool
	resetYes  bool
)

var cryptCmd = &cobra.Command{
	Use:   "crypt",
	Short: "See the monsters you've captured and the badges you've earned",
	Long: `Your crypt holds every monster you've captured by answering questions about
it in the quiz or Guess the Monster, plus your badges and records.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := store.Load()
		if err != nil {
			return err
		}
		if cryptJSON {
			return printJSON(p)
		}
		fmt.Print(ui.RenderCrypt(p, monsters.GetAllMonsters(), ui.TerminalWidth()))
		if path, err := store.Path(); err == nil {
			fmt.Println()
			fmt.Println(ui.Dim("Progress is saved in " + ui.Tilde(path)))
		}
		return nil
	},
}

var cryptResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Empty your crypt and start again",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !resetYes {
			fmt.Print("This releases every captured monster and forgets all your badges. Type 'yes' to confirm: ")
			line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			if strings.TrimSpace(strings.ToLower(line)) != "yes" {
				fmt.Println("Phew. Your crypt is untouched.")
				return nil
			}
		}
		if err := store.Reset(); err != nil {
			return err
		}
		fmt.Println("The crypt doors swing open... every monster has escaped. Your progress has been reset.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cryptCmd)
	cryptCmd.AddCommand(cryptResetCmd)
	cryptCmd.Flags().BoolVar(&cryptJSON, "json", false, "Output your progress as JSON")
	cryptResetCmd.Flags().BoolVarP(&resetYes, "yes", "y", false, "Don't ask for confirmation")
}
