package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/store"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var packsJSON bool

// packJSON is the machine-readable shape of a pack listing.
type packJSON struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Source      string   `json:"source"`
	Monsters    []string `json:"monsters"`
}

var packsCmd = &cobra.Command{
	Use:   "packs",
	Short: "List monster packs, or create your own",
	Long: `Monsters come in packs. Terminal of Terror ships with Universal Classics
and World Folklore, and you can add community packs by dropping them into
your config directory. Use --pack with any command to pick packs,
e.g. 'terminal-of-terror quiz --pack folklore'.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		all := monsters.AllPacks()
		if packsJSON {
			var out []packJSON
			for _, p := range all {
				pj := packJSON{ID: p.ID, Name: p.Name, Description: p.Description, Source: p.Source}
				for _, m := range p.Monsters {
					pj.Monsters = append(pj.Monsters, m.ID)
				}
				out = append(out, pj)
			}
			return printJSON(out)
		}
		dir, _ := communityDir()
		fmt.Print(ui.RenderPacks(all, dir, ui.TerminalWidth()))
		return nil
	},
}

var packIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

var packsNewCmd = &cobra.Command{
	Use:   "new <pack-id>",
	Short: "Create a community pack with an example monster to edit",
	Example: `  terminal-of-terror packs new cryptids
  terminal-of-terror monster "my cryptids monster"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		if !packIDPattern.MatchString(id) {
			return fmt.Errorf("pack ids use lowercase letters, numbers and dashes, e.g. 'cryptids'")
		}
		dir, err := communityDir()
		if err != nil {
			return err
		}
		packDir := filepath.Join(dir, id)
		if _, err := os.Stat(packDir); err == nil {
			return fmt.Errorf("%s already exists", packDir)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		if err := os.MkdirAll(packDir, 0o755); err != nil {
			return err
		}
		// Base the example monster on the pack id, so monsters from different
		// packs never share an id or name (duplicates are skipped at load).
		monsterID := id + "-monster"
		monsterName := "My " + id + " monster"
		files := map[string]any{
			"pack.json": map[string]any{
				"id":          id,
				"name":        "My " + id + " pack",
				"description": "Describe your pack here.",
				"order":       []string{monsterID},
			},
			monsterID + ".json": exampleMonster(monsterID, monsterName),
		}
		for name, v := range files {
			raw, _ := json.MarshalIndent(v, "", "  ")
			if err := os.WriteFile(filepath.Join(packDir, name), append(raw, '\n'), 0o644); err != nil {
				return err
			}
		}
		art := "   .-\"\"\"-.\n  /  o o  \\\n |    ^    |\n  \\ '---' /\n   '-----'\n"
		if err := os.WriteFile(filepath.Join(packDir, monsterID+".txt"), []byte(art), 0o644); err != nil {
			return err
		}
		fmt.Printf("Created %s\n\n", ui.Tilde(packDir))
		fmt.Printf("Edit %s.json (and %s.txt for its portrait), then try:\n", monsterID, monsterID)
		fmt.Printf("  terminal-of-terror monster %q\n  terminal-of-terror quiz --pack %s\n\n", monsterName, id)
		fmt.Println("Only name, description and facts are required. See CONTRIBUTING.md for every field.")
		return nil
	},
}

func communityDir() (string, error) {
	d, err := store.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "packs"), nil
}

func exampleMonster(id, name string) map[string]any {
	return map[string]any{
		"id":          id,
		"name":        name,
		"emoji":       "👹",
		"description": "A one-line description of your monster",
		"origin":      "Where the legend comes from",
		"debut":       map[string]any{"medium": "folklore", "title": "Where it first appeared", "era": "when"},
		"legend":      "A paragraph about the real history or folklore behind the monster.",
		"myths": []map[string]any{
			{"claim": "A popular belief about the monster.", "true": false, "explanation": "What's really going on."},
		},
		"powers":     []string{"A terrifying power"},
		"weaknesses": []string{"A surprising weakness"},
		"stats":      map[string]int{"strength": 5, "speed": 5, "cunning": 5, "dread": 5},
		"facts": []string{
			"A fascinating fact",
			"Another fascinating fact",
			"Add at least three facts so the quiz has plenty to ask",
		},
		"hostIntro": "What Count Cathode says to introduce your monster.",
		"theme":     map[string]string{"primary": "#FF6B6B", "accent": "#FFE66D"},
	}
}

func init() {
	rootCmd.AddCommand(packsCmd)
	packsCmd.AddCommand(packsNewCmd)
	packsCmd.Flags().BoolVar(&packsJSON, "json", false, "Output the packs as JSON")
}
