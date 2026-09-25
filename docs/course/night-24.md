# Night 24 · By the Light of the Moon

> 📺 *"Werewolves care about one thing: the moon. Tonight we compute it. No internet, no almanac, just one reference date, one number, and a cosine. Then Halloween, Friday the 13th, and the anniversary of every film in the vault. It's the most maths we'll do all month, and it fits on one screen."*

**Tonight you'll learn**
- Pure functions of time, and why they take a `time.Time`
- The mean lunar month: one constant, one reference, `math.Mod`
- `time.Time` arithmetic: `Sub`, `AddDate`, `Weekday`
- Dates without times, and time zones
- Tests with known answers from the real sky
- `time.Parse` layouts, and Go's odd reference date

**Where we are:** Night 22's badges call `calendar.IsFullMoon` and friends. Tonight, the package behind them.

## Functions of time

Everything in [`internal/calendar/calendar.go`](../../internal/calendar/calendar.go) takes a `time.Time` and returns an answer. Nothing calls `time.Now()` inside the package. That one rule makes every function testable with a fixed date, lets `countdown --date 2026-10-31` pretend it's Halloween, and lets the demo recordings on Night 30 show the same screen every time.

## The moon in ten lines

A full astronomical model of the moon is thousands of lines. But the *average* time between new moons is a very stable number, and for "is it a full moon tonight?" the average is within a day of the truth, which is plenty for werewolves:

```go
// synodicMonth is the average time between new moons, in days.
const synodicMonth = 29.530588853

// knownNewMoon is a reference new moon: 6 January 2000, 18:14 UTC.
var knownNewMoon = time.Date(2000, 1, 6, 18, 14, 0, 0, time.UTC)

// Moon returns the moon's phase at t, using the mean lunar month. It is
// accurate to within about a day, which is plenty for werewolves.
func Moon(t time.Time) MoonPhase {
	days := t.Sub(knownNewMoon).Hours() / 24
	age := math.Mod(days, synodicMonth)
	if age < 0 {
		age += synodicMonth
	}
	p := MoonPhase{
		Age:          age,
		Illumination: (1 - math.Cos(2*math.Pi*age/synodicMonth)) / 2,
	}
	eighths := age / synodicMonth * 8
	for _, ph := range phases {
		if eighths < ph.until {
			p.Name, p.Emoji = ph.name, ph.emoji
			break
		}
	}
	return p
}
```

Days since a known new moon, modulo the month length, is the moon's *age* in days. `math.Mod` is floating-point remainder, and can be negative for dates before 2000, hence the fix-up. Illumination is a cosine: 0 at age 0 (new), 1 at half a month (full), back to 0. The phase name comes from a table of eighths, `{0.5, "New Moon", "🌑"}, {1.5, "Waxing Crescent", "🌒"}, ...`, so the full moon is the eighth from 3.5 to 4.5.

```go
// IsFullMoon reports whether t falls within a day of the full moon.
func IsFullMoon(t time.Time) bool {
	return math.Abs(Moon(t).Age-synodicMonth/2) < 1.0
}
```

That's the whole of the astronomy. The comment states the accuracy, which is the honest thing to do with an approximation: someone will compare it to a real almanac one day, and the comment tells them what to expect.

## Dates, not moments

```go
// NightsUntilHalloween counts the nights until the next 31 October (0 on
// Halloween itself).
func NightsUntilHalloween(t time.Time) int {
	today := dateOnly(t)
	h := time.Date(today.Year(), time.October, 31, 0, 0, 0, 0, time.UTC)
	if today.After(h) {
		h = h.AddDate(1, 0, 0)
	}
	return int(h.Sub(today).Hours()/24 + 0.5)
}

// dateOnly strips the time of day, keeping the local calendar date.
func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
```

"How many nights until Halloween" is a question about *dates*, and a `time.Time` is a *moment*. At 11pm on 30 October the honest answer is 1, not 0.04. So `dateOnly` takes the calendar date the caller's clock shows and rebuilds it at midnight UTC, and the subtraction is between whole days. The `+ 0.5` rounds, because a leap second or a daylight-saving shift can make the difference 23.9999 hours.

`h.AddDate(1, 0, 0)` adds a year and handles leap years correctly; `h.Add(365 * 24 * time.Hour)` would not, which is why the test includes a year that crosses 29 February:

```go
func TestHalloweenCountdown(t *testing.T) {
	cases := map[time.Time]int{
		utc(2026, time.September, 23, 20): 38,
		utc(2026, time.October, 30, 23):   1,
		utc(2026, time.October, 31, 1):    0,
		utc(2026, time.November, 1, 0):    364,
		utc(2027, time.November, 1, 0):    365, // 2028 is a leap year
	}
	...
```

`IsFriday13` is a two-line `Weekday()` and `Day()` check; `IsHalloween` compares month and day. Trivial, but each is tested with a real date, because "13 October 2026 is a Tuesday" is the kind of fact you want in a test rather than in your head.

## Known answers

The moon test uses three real full moons:

```go
func TestMoonMatchesKnownPhases(t *testing.T) {
	fulls := []time.Time{
		utc(2024, time.April, 23, 23),
		utc(2025, time.October, 7, 3),
		utc(2023, time.August, 31, 1),
	}
	for _, f := range fulls {
		if !IsFullMoon(f) { ... }
	}
```

When you implement a formula, find some answers you *know* from an independent source and pin them. If someone later "improves" the constant, the sky will disagree with them.

## Anniversaries

Every film has a `releaseDate` in its JSON. `Anniversaries(t, all)` finds films released on this day of the year; `NextAnniversary` finds the soonest one after today:

```go
func releaseDate(m monsters.Monster) (time.Time, bool) {
	if m.Film == nil || m.Film.ReleaseDate == "" {
		return time.Time{}, false
	}
	d, err := time.Parse("2006-01-02", m.Film.ReleaseDate)
	return d, err == nil
}
```

**Go's date layouts** are the strangest thing you'll meet this month. Instead of `YYYY-MM-DD`, Go uses a *reference date*, Monday 2 January 2006 at 15:04:05, and you write your format by writing that date in your layout: `"2006-01-02"` means year-month-day, `"2 January"` means day and full month name, `"Monday 2 January 2006"` for the ticket. The digits are 1 2 3 4 5 6 7 in order (month 1, day 2, hour 3, minute 4, second 5, year 6, zone 7), which is the mnemonic. Everyone finds it odd; everyone gets used to it.

`Notes(t, all)` gathers everything special about tonight into a list of strings, most exciting first: Halloween, anniversaries, a full moon, Friday the 13th, the countdown. The intro shows up to three of them; the fact card and the ticket show them all. One function, several screens.

## Daily seeds

```go
// DailySeed turns a date into a stable random seed, so everyone gets the
// same fact of the night on the same day.
func DailySeed(t time.Time) int64 {
	return int64(t.Year()*10000 + int(t.Month())*100 + t.Day())
}
```

Tomorrow's `random --daily` and `tonight` need a random choice that's the same for everyone all day and different tomorrow. A seed made from the date does it: 20261031 today, 20261101 tomorrow. Randomness with a chosen seed is *deterministic*, and that's a feature you'll reach for more often than you'd think.

## Parsing the user's date

```go
// Date parses YYYY-MM-DD as that evening at 9pm local time, or returns now
// for an empty string.
func Date(s string) (time.Time, error) {
	if s == "" {
		return time.Now(), nil
	}
	d, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("dates look like 2026-10-31: %w", err)
	}
	return d.Add(21 * time.Hour), nil
}
```

The one place the package calls `time.Now()`, and it's the *edge*: the function that turns a `--date` flag into a `time.Time`, used by every command with the flag. "9pm local" is a choice: `--date 2026-10-31` should mean Halloween *night*, when the moon's up and the Night Owl badge is out of reach. The error message shows the format instead of Go's own `parsing time "31/10/2026": ...`, which is accurate and unhelpful.

## Run it

```bash
go run . countdown
go run . countdown --date 2026-10-31
go run . countdown --date 2026-10-13
```

Tonight's moon and the next anniversary; the pumpkin on Halloween; Night 13 of the 31 Nights of Fright in October. Tomorrow's lesson draws these.

## Try it

- Print the moon's age and illumination for every day this month and compare with a calendar app. Where does the mean-month model drift?
- Add `IsWalpurgisNight` (30 April, the other great witches' night) and a note for it.
- What does `NightsUntilHalloween` return on 29 February? Write the test before you run it.

## 💀 Terrifying fact

`time.Time` carries a location, and comparing two times in different zones with `==` compares the location too, so equal instants can be "not equal". Use `t.Equal(u)` for instants and `t.Before`/`t.After` for order; reserve `==` for the zero value check (`t.IsZero()` is clearer). The map in `TestHalloweenCountdown` keyed by `time.Time` works only because every key is built the same way, in UTC.

## 🕯️ Before dawn

`NightOfFright` returns the day of the month in October. Make it return a *monster* too: night N features `all[(N-1) % len(all)]`, which is what tomorrow's countdown screen does. Then work out which night features the Golem, and check with `--date`.

> 📺 *"A constant, a cosine, and the moon is yours. Tomorrow night, tickets. Actual tickets, with a perforated edge. Admit one."*

[← Night 23](night-23.md) · [Index](README.md) · [Night 25 →](night-25.md)
