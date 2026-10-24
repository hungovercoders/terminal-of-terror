package cmd

import (
	"fmt"
	"math/rand"

	"github.com/hungovercoders/terminal-of-terror/internal/calendar"
	"github.com/hungovercoders/terminal-of-terror/internal/host"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var (
	tonightDate    string
	tonightShuffle bool
)

// tonightJSON is the double bill as --json prints it.
type tonightJSON struct {
	Date     string        `json:"date"`
	Features []featureJSON `json:"features"`
	Notes    []string      `json:"notes,omitempty"`
}

type featureJSON struct {
	Time    string `json:"time"`
	Monster string `json:"monster"`
	Name    string `json:"name"`
	Film    string `json:"film,omitempty"`
	Year    int    `json:"year,omitempty"`
}

var tonightCmd = &cobra.Command{
	Use:   "tonight",
	Short: "See tonight's Creature Feature double bill",
	Long: `Print a ticket for tonight's Channel 13 double feature. The line-up changes
every night, so check back tomorrow, or use --shuffle for a different bill.
During October the 31 Nights of Fright monster headlines the bill.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		now, err := parseDate(tonightDate)
		if err != nil {
			return err
		}
		all := monsters.GetAllMonsters()
		if len(all) < 2 {
			return fail(codeUsage, "try a different --pack", "a double feature needs at least two monsters")
		}
		s := calendar.DailySeed(now) * 13
		if tonightShuffle {
			s = newRand().Int63()
		}
		r := rand.New(rand.NewSource(s))
		pick := r.Perm(len(all))
		first, second := all[pick[0]], all[pick[1]]
		if f, ok := calendar.TonightsFright(now, all); ok && !tonightShuffle {
			// The 31 Nights of Fright: tonight's monster tops the bill, and
			// the support act is still the night's draw.
			first, second = f.Monster, all[pick[0]]
			if second.ID == first.ID {
				second = all[pick[1]]
			}
		}
		if jsonOut {
			feature := func(at string, m monsters.Monster) featureJSON {
				f := featureJSON{Time: at, Monster: m.ID, Name: m.Name}
				if m.Film != nil {
					f.Film, f.Year = m.Film.Title, m.Film.Year
				}
				return f
			}
			return printJSON(tonightJSON{
				Date:     now.Format("2006-01-02"),
				Features: []featureJSON{feature("9:00 PM", first), feature("10:45 PM", second)},
				Notes:    calendar.Notes(now, all),
			})
		}

		w := ui.TerminalWidth()
		fmt.Print(ui.RenderTicket(now, first, second, w))
		card := ui.FactCard{
			Title:   "FROM THE HOST",
			Monster: first,
			Fact:    host.Intro(first),
			Extras:  calendar.Notes(now, all),
		}
		fmt.Print(card.Render(w))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tonightCmd)
	tonightCmd.Flags().StringVar(&tonightDate, "date", "", "See the bill for another night (YYYY-MM-DD)")
	tonightCmd.Flags().BoolVar(&tonightShuffle, "shuffle", false, "Pick a random double bill instead of tonight's")
}
