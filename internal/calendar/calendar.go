// Package calendar knows the spooky side of dates: moon phases, the
// Halloween countdown, Friday the 13th and classic film anniversaries.
// Everything is computed offline.
package calendar

import (
	"fmt"
	"math"
	"time"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

// synodicMonth is the average time between new moons, in days.
const synodicMonth = 29.530588853

// knownNewMoon is a reference new moon: 6 January 2000, 18:14 UTC.
var knownNewMoon = time.Date(2000, 1, 6, 18, 14, 0, 0, time.UTC)

// MoonPhase describes the moon at a moment.
type MoonPhase struct {
	Age          float64 // days since the last new moon
	Illumination float64 // 0 (new) to 1 (full)
	Name         string
	Emoji        string
}

var phases = []struct {
	until float64 // upper bound on age, in eighths of a month
	name  string
	emoji string
}{
	{0.5, "New Moon", "🌑"},
	{1.5, "Waxing Crescent", "🌒"},
	{2.5, "First Quarter", "🌓"},
	{3.5, "Waxing Gibbous", "🌔"},
	{4.5, "Full Moon", "🌕"},
	{5.5, "Waning Gibbous", "🌖"},
	{6.5, "Last Quarter", "🌗"},
	{7.5, "Waning Crescent", "🌘"},
	{8.0, "New Moon", "🌑"},
}

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

// IsFullMoon reports whether t falls within a day of the full moon.
func IsFullMoon(t time.Time) bool {
	return math.Abs(Moon(t).Age-synodicMonth/2) < 1.0
}

// IsFriday13 reports whether t is a Friday the 13th.
func IsFriday13(t time.Time) bool {
	return t.Weekday() == time.Friday && t.Day() == 13
}

// IsHalloween reports whether t is 31 October.
func IsHalloween(t time.Time) bool {
	return t.Month() == time.October && t.Day() == 31
}

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

// NightOfFright returns which of the 31 Nights of Fright t is, during October.
func NightOfFright(t time.Time) (int, bool) {
	if t.Month() != time.October {
		return 0, false
	}
	return t.Day(), true
}

// DailySeed turns a date into a stable random seed, so everyone gets the
// same fact of the night on the same day.
func DailySeed(t time.Time) int64 {
	return int64(t.Year()*10000 + int(t.Month())*100 + t.Day())
}

// Anniversary is a classic film released on this day of the year.
type Anniversary struct {
	Monster monsters.Monster
	Years   int
}

// Anniversaries lists films released on t's day and month.
func Anniversaries(t time.Time, all []monsters.Monster) []Anniversary {
	var out []Anniversary
	for _, m := range all {
		if rd, ok := releaseDate(m); ok && rd.Month() == t.Month() && rd.Day() == t.Day() && t.Year() > rd.Year() {
			out = append(out, Anniversary{Monster: m, Years: t.Year() - rd.Year()})
		}
	}
	return out
}

// NextAnniversary finds the next film anniversary after t.
func NextAnniversary(t time.Time, all []monsters.Monster) (Anniversary, time.Time, bool) {
	today := dateOnly(t)
	var best Anniversary
	var bestDate time.Time
	found := false
	for _, m := range all {
		rd, ok := releaseDate(m)
		if !ok {
			continue
		}
		next := time.Date(today.Year(), rd.Month(), rd.Day(), 0, 0, 0, 0, time.UTC)
		if !next.After(today) {
			next = next.AddDate(1, 0, 0)
		}
		if !found || next.Before(bestDate) {
			best, bestDate, found = Anniversary{Monster: m, Years: next.Year() - rd.Year()}, next, true
		}
	}
	return best, bestDate, found
}

func releaseDate(m monsters.Monster) (time.Time, bool) {
	if m.Film == nil || m.Film.ReleaseDate == "" {
		return time.Time{}, false
	}
	d, err := time.Parse("2006-01-02", m.Film.ReleaseDate)
	return d, err == nil
}

// dateOnly strips the time of day, keeping the local calendar date.
func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// Notes lists anything special about t, most exciting first.
func Notes(t time.Time, all []monsters.Monster) []string {
	var out []string
	if IsHalloween(t) {
		out = append(out, "🎃 It's Halloween! The busiest night of the year for monsters.")
	}
	for _, a := range Anniversaries(t, all) {
		out = append(out, fmt.Sprintf("🎬 On this day in %d, %s was released: %d years ago tonight.", t.Year()-a.Years, a.Monster.Film.Title, a.Years))
	}
	if IsFullMoon(t) {
		out = append(out, "🌕 Full moon tonight. Keep the silver handy and your doors locked.")
	}
	if IsFriday13(t) {
		out = append(out, "🐈 Friday the 13th! Unlucky for some, delightful for monsters.")
	}
	if n, ok := NightOfFright(t); ok && !IsHalloween(t) {
		out = append(out, fmt.Sprintf("🎃 Night %d of the 31 Nights of Fright · %s until Halloween.", n, plural(NightsUntilHalloween(t), "night")))
	} else if d := NightsUntilHalloween(t); d > 0 && d <= 60 {
		out = append(out, fmt.Sprintf("🎃 %s until Halloween.", plural(d, "night")))
	}
	return out
}

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

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}
