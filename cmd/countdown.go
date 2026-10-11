package cmd

import (
	"fmt"

	"github.com/hungovercoders/terminal-of-terror/internal/calendar"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var countdownDate string

// countdownJSON is the countdown as --json prints it.
type countdownJSON struct {
	Date                 string           `json:"date"`
	Halloween            bool             `json:"halloween"`
	NightsUntilHalloween int              `json:"nightsUntilHalloween"`
	NightOfFright        *nightJSON       `json:"nightOfFright,omitempty"`
	Moon                 moonJSON         `json:"moon"`
	NextAnniversary      *anniversaryJSON `json:"nextAnniversary,omitempty"`
	Notes                []string         `json:"notes,omitempty"`
}

type nightJSON struct {
	Night   int    `json:"night"`
	Monster string `json:"monster"`
	Name    string `json:"name"`
}

type moonJSON struct {
	Name         string  `json:"name"`
	Emoji        string  `json:"emoji"`
	Illumination float64 `json:"illumination"`
}

type anniversaryJSON struct {
	Monster string `json:"monster"`
	Film    string `json:"film"`
	Years   int    `json:"years"`
	Date    string `json:"date"`
}

var countdownCmd = &cobra.Command{
	Use:   "countdown",
	Short: "Count down the nights until Halloween",
	Long: `Show how many nights are left until Halloween, tonight's moon and the next
classic film anniversary. During October, each night of the 31 Nights of
Fright features a different monster.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		now, err := parseDate(countdownDate)
		if err != nil {
			return err
		}
		all := monsters.GetAllMonsters()
		if jsonOut {
			moon := calendar.Moon(now)
			out := countdownJSON{
				Date:                 now.Format("2006-01-02"),
				Halloween:            calendar.IsHalloween(now),
				NightsUntilHalloween: calendar.NightsUntilHalloween(now),
				Moon:                 moonJSON{Name: moon.Name, Emoji: moon.Emoji, Illumination: moon.Illumination},
				Notes:                calendar.Notes(now, all),
			}
			if night, ok := calendar.NightOfFright(now); ok && len(all) > 0 {
				m := all[(night-1)%len(all)]
				out.NightOfFright = &nightJSON{Night: night, Monster: m.ID, Name: m.Name}
			}
			if a, when, ok := calendar.NextAnniversary(now, all); ok {
				out.NextAnniversary = &anniversaryJSON{Monster: a.Monster.ID, Film: a.Monster.Film.Title, Years: a.Years, Date: when.Format("2006-01-02")}
			}
			return printJSON(out)
		}
		fmt.Print(ui.RenderCountdown(now, all, ui.TerminalWidth()))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(countdownCmd)
	countdownCmd.Flags().StringVar(&countdownDate, "date", "", "Pretend it's this date (YYYY-MM-DD)")
}
