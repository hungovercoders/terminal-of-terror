# Night 25 · Nightly Rituals

> 📺 *"A horror host's real job is habit. Same channel, same time, every night. Tonight the program learns three rituals: a double-feature ticket for the evening, a countdown with a pumpkin the size of your screen, and a Fact of the Night that's the same for everyone until midnight. Put one in your shell and I'll see you every morning."*

**Tonight you'll learn**
- Turning a date into a stable random choice
- Composing screens from the pieces you already have
- A three-row digit font, and why two rows weren't enough
- `JoinVertical` and `JoinHorizontal` together
- Optional data on a card: `Extras`
- Designing output for a shell's startup file

**Where we are:** the calendar package knows about tonight. Three commands draw it.

## The Fact of the Night

`random` has been in the program since Night 4. It now has `--daily`, in [`cmd/random.go`](../../cmd/random.go):

```go
		s := seed()
		title := "RANDOM TERROR FACT"
		if randomDaily || randomDate != "" {
			s = calendar.DailySeed(now)
			title = "FACT OF THE NIGHT · " + now.Format("Monday 2 January")
		}
		r := rand.New(rand.NewSource(s))
		m, fact := monsters.RandomFact(r)
		notes := calendar.Notes(now, monsters.GetAllMonsters())
		...
		card := ui.FactCard{Title: title, Monster: m, Fact: fact, Host: host.Quip(r), Extras: notes}
		fmt.Print(card.Render(ui.TerminalWidth()))
```

The only difference between random and daily is *where the seed comes from*. Night 24's `DailySeed` gives everyone in the world the same fact today and a new one tomorrow, which turns a random command into a ritual. And because `RandomFact` and `host.Quip` draw from the same `r`, the host's quip is the same all day too.

The card grew two optional fields since Night 10:

```go
type FactCard struct {
	Title   string
	Monster monsters.Monster
	Fact    string
	Host    string   // host's comment; empty for none
	Extras  []string // extra lines shown under the fact (calendar notes etc.)
}
```

Empty means absent, so every old caller still works, and the daily card gets "🌕 Full moon tonight" or "🎬 On this day in 1931..." under the fact when there's something to say. Growing a struct with optional fields, rather than adding a `RenderDailyCard`, keeps one renderer and one test.

### For your `.bashrc`

The README suggests `terminal-of-terror random --daily` in a shell startup file, and that shapes the output: short, one card, no interaction, no pauses, and clean text when there's no terminal. Every one of those was a decision made earlier in the course, on Nights 9, 10 and 23, and this is where they pay off. Output meant for a startup file is a good test of a command's manners.

## The ticket

`tonight` picks two monsters for the evening's double bill, in [`cmd/tonight.go`](../../cmd/tonight.go):

```go
		s := calendar.DailySeed(now) * 13
		if tonightShuffle {
			s = newRand().Int63()
		}
		r := rand.New(rand.NewSource(s))
		pick := r.Perm(len(all))
		first, second := all[pick[0]], all[pick[1]]
```

`DailySeed * 13` so the bill isn't just the Fact of the Night's monster again: the same date, a different seed, a different draw. `r.Perm` guarantees two *different* monsters, where two `Intn` calls would sometimes pick the same one.

`RenderTicket` in [`internal/ui/rituals.go`](../../internal/ui/rituals.go) is worth reading as a composition exercise. Each feature line is a closure that puts a fixed-width time column beside a title and credit:

```go
	feature := func(at string, m monsters.Monster) string {
		p := paletteFor(m)
		title := strings.ToUpper(m.Name)
		credit := "A tale from " + m.Origin
		if f := m.Film; f != nil {
			title = fmt.Sprintf("%s (%d)", strings.ToUpper(f.Title), f.Year)
			credit = fmt.Sprintf("Directed by %s · starring %s", f.Director, f.Star)
			if f.Silent {
				credit += " · silent"
			}
		}
		timeCol := lipgloss.NewStyle().Width(10).Foreground(colorOrange).Bold(true).Render(at)
		text := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Foreground(p.primary).Render(m.Emoji+" "+title),
			helpStyle.Render(wrap(credit, inner-12)),
		)
		return lipgloss.JoinHorizontal(lipgloss.Top, timeCol, text)
	}
```

A film gets its title, year, director and star; a folklore monster gets "A tale from Slavic folklore". Then the body is a list of parts, with a "perforation" of dashes, in a `DoubleBorder` with a fixed `Width`, so the ticket is the same shape whatever's printed on it. `JoinVertical` stacks; `JoinHorizontal` places side by side; nesting them is how every non-trivial terminal layout is built.

## The countdown

The big number needs a font, and Night 15's two-row letters won't do for digits:

```go
// digits is a three-row block font; two rows made 3 and 8 look alike.
var digits = map[rune][3]string{
	'0': {"█▀█", "█ █", "█▄█"},
	'1': {"▀█ ", " █ ", "▄█▄"},
	...
	'8': {"█▀█", "█▀█", "█▄█"},
```

The comment is the design note: two rows were tried, 3 and 8 were indistinguishable, so three rows it is. `bigNumber` builds three rows across the digits of `fmt.Sprint(n)`, the same shape as `bigText`. `TestBigNumber` pins the rendering of 38, the pair that two rows couldn't tell apart, character for character.

`RenderCountdown` puts the pumpkin beside the number:

```go
			lipgloss.JoinHorizontal(lipgloss.Center, orange.Render(pumpkin), "    ",
				lipgloss.JoinVertical(lipgloss.Left, orange.Render(bigNumber(n)), "", headingStyle.Render(label))),
```

`lipgloss.Center` as the alignment for `JoinHorizontal` centres the shorter block *vertically* against the taller: the number floats in the middle of the pumpkin's height. And the pumpkin itself is a raw string literal:

```go
const pumpkin = `         )
    .-"""(""-.
  .'  ^  |  ^  '.
 /   /_\ | /_\   \
...
```

Backticks make a **raw string**: no escapes, so backslashes and quotes are literal and the art can be pasted as-is. Any multi-line text in Go code wants backticks.

In October the screen adds the 31 Nights of Fright: night N features monster `(N-1) % len(all)`, with fact `(N-1) / len(all) % len(facts)`, so after a lap of the vault the second lap shows each monster's second fact. Arithmetic on a date turns nineteen monsters into thirty-one different nights.

## Run it

```bash
go run . random --daily
go run . tonight
go run . tonight --date 2026-10-31
go run . countdown --date 2026-10-13
```

Run `random --daily` twice: same fact, same quip. `tonight --shuffle` for a different bill. Every ritual command takes `--date`, so you can see any night you like, which is also how the README screenshots were made.

## Try it

- Add `--date` to `random`'s sibling, `list`? No: `list` has nothing to do with dates. Decide which commands *should* take `--date` and check the project agrees.
- Make the ticket's showtimes depend on the season: 9pm in summer, 8pm in winter, via `now.Month()`.
- The countdown's next-anniversary line uses `when.Format("2 January")`. Change it to include the weekday.

## 💀 Terrifying fact

`fmt.Sprint(n)` for the digit font, not `strconv.Itoa(n)`. Both work; `Itoa` is faster and says what it does. `Sprint` is used here because the loop ranges over the *runes* of the result and reads as "for each character of the number". Speed doesn't matter for a two-digit countdown. It's worth knowing that `fmt` is the slow, flexible option and `strconv` the fast, specific one, for the day it does matter.

## 🕯️ Before dawn

Add `terminal-of-terror random --daily` to your shell's startup file and leave it there for the rest of the course. Then, tomorrow morning, notice what you'd change about the card now that you see it daily. Small outputs get polished by use, not by design.

> 📺 *"Ticket, countdown, fact. Three rituals for the discerning insomniac. Tomorrow night we stop talking to people entirely, and start talking to other programs."*

[← Night 24](night-24.md) · [Index](README.md) · [Night 26 →](night-26.md)
