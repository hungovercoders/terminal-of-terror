package calendar

import (
	"strings"
	"testing"
	"time"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

func utc(y int, m time.Month, d, h int) time.Time { return time.Date(y, m, d, h, 0, 0, 0, time.UTC) }

func TestMoonMatchesKnownPhases(t *testing.T) {
	fulls := []time.Time{
		utc(2024, time.April, 23, 23),
		utc(2025, time.October, 7, 3),
		utc(2023, time.August, 31, 1),
	}
	for _, f := range fulls {
		if !IsFullMoon(f) {
			t.Errorf("%s should be a full moon, age %.1f", f.Format(time.DateOnly), Moon(f).Age)
		}
		if Moon(f).Name != "Full Moon" || Moon(f).Illumination < 0.95 {
			t.Errorf("%s: got %+v", f.Format(time.DateOnly), Moon(f))
		}
	}
	news := []time.Time{utc(2025, time.October, 21, 12), utc(2024, time.April, 8, 18)}
	for _, n := range news {
		if IsFullMoon(n) || Moon(n).Name != "New Moon" {
			t.Errorf("%s should be a new moon: %+v", n.Format(time.DateOnly), Moon(n))
		}
	}
}

func TestHalloweenCountdown(t *testing.T) {
	cases := map[time.Time]int{
		utc(2026, time.September, 23, 20): 38,
		utc(2026, time.October, 30, 23):   1,
		utc(2026, time.October, 31, 1):    0,
		utc(2026, time.November, 1, 0):    364,
		utc(2027, time.November, 1, 0):    365, // 2028 is a leap year
	}
	for d, want := range cases {
		if got := NightsUntilHalloween(d); got != want {
			t.Errorf("%s: got %d, want %d", d.Format(time.DateOnly), got, want)
		}
	}
}

func TestFriday13(t *testing.T) {
	if !IsFriday13(utc(2026, time.November, 13, 12)) {
		t.Error("13 Nov 2026 is a Friday")
	}
	if IsFriday13(utc(2026, time.October, 13, 12)) {
		t.Error("13 Oct 2026 is a Tuesday")
	}
}

func TestAnniversariesAndNotes(t *testing.T) {
	all := monsters.GetAllMonsters()
	d := utc(2026, time.November, 21, 21)
	a := Anniversaries(d, all)
	if len(a) != 1 || a[0].Monster.ID != "frankenstein" || a[0].Years != 95 {
		t.Fatalf("got %+v", a)
	}
	notes := strings.Join(Notes(d, all), "\n")
	if !strings.Contains(notes, "Frankenstein was released: 95 years ago") {
		t.Errorf("notes missing anniversary:\n%s", notes)
	}

	// Son of Dracula (5 November 1943) comes before The Invisible Man (13 November 1933).
	next, when, ok := NextAnniversary(utc(2026, time.September, 23, 20), all)
	if !ok || next.Monster.ID != "son-of-dracula" || when.Day() != 5 {
		t.Errorf("next anniversary after 23 Sep: %s on %s", next.Monster.ID, when)
	}

	if n := strings.Join(Notes(utc(2026, time.October, 5, 21), all), "\n"); !strings.Contains(n, "Night 5 of the 31 Nights of Fright") {
		t.Errorf("October notes:\n%s", n)
	}
	if n := Notes(utc(2026, time.March, 3, 21), all); len(n) != 0 && strings.Contains(strings.Join(n, ""), "Halloween") {
		t.Errorf("March shouldn't mention Halloween: %v", n)
	}
}

func TestDailySeedIsStable(t *testing.T) {
	if DailySeed(utc(2026, time.October, 31, 1)) != DailySeed(utc(2026, time.October, 31, 23)) {
		t.Error("same day, same seed")
	}
	if DailySeed(utc(2026, time.October, 31, 1)) == DailySeed(utc(2026, time.November, 1, 1)) {
		t.Error("different days, different seeds")
	}
}

func TestDateErrorIsReadable(t *testing.T) {
	_, err := Date("31/10")
	if err == nil || err.Error() != `dates look like 2026-10-31, not "31/10"` {
		t.Errorf("Date(31/10) = %v", err)
	}
}

func TestTonightsFright(t *testing.T) {
	all := monsters.GetAllMonsters()
	f, ok := TonightsFright(utc(2026, time.October, 13, 21), all)
	if !ok || f.Night != 13 || f.Monster.ID != all[12].ID || f.Fact != all[12].Facts[0] {
		t.Errorf("13 October should feature the 13th monster and its first fact, got %+v", f)
	}
	if _, ok := TonightsFright(utc(2026, time.March, 3, 21), all); ok {
		t.Error("March has no Night of Fright")
	}
	if _, ok := TonightsFright(utc(2026, time.October, 3, 21), nil); ok {
		t.Error("no monsters, no Night of Fright")
	}
	// A two-monster pack laps the vault: night 5 is the first monster's third fact.
	two := all[:2]
	f, ok = TonightsFright(utc(2026, time.October, 5, 21), two)
	if !ok || f.Monster.ID != two[0].ID || f.Fact != two[0].Facts[2%len(two[0].Facts)] {
		t.Errorf("night 5 of a two-monster pack: got %+v", f)
	}
}
