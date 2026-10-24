package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// captureJSON runs a command's RunE under --json and returns what it wrote.
func captureJSON(t *testing.T, c *cobra.Command, args []string) []byte {
	t.Helper()
	var buf bytes.Buffer
	stdout, jsonOut = &buf, true
	defer func() { stdout, jsonOut = nil, false }()
	if err := c.RunE(c, args); err != nil {
		t.Fatalf("%s --json: %v", c.Name(), err)
	}
	if !json.Valid(buf.Bytes()) {
		t.Fatalf("%s --json is not JSON:\n%s", c.Name(), buf.String())
	}
	return buf.Bytes()
}

// Every command's --json parses and carries the keys a script would read.
func TestEveryCommandHasAJSONShape(t *testing.T) {
	t.Setenv("TERMINAL_OF_TERROR_HOME", t.TempDir())
	t.Setenv("TERMINAL_OF_TERROR_SEED", "13")
	tonightDate, countdownDate, randomDate = "2026-10-31", "2026-10-13", "2026-10-31"
	quizQuestions, guessRounds = 2, 2
	defer func() { tonightDate, countdownDate, randomDate = "", "", "" }()
	cases := []struct {
		cmd  *cobra.Command
		args []string
		keys []string
	}{
		{listCmd, nil, []string{`"id": "dracula"`}},
		{monsterCmd, []string{"dracula"}, []string{`"name": "Dracula"`}},
		{randomCmd, nil, []string{`"monster":`, `"fact":`}},
		{quizCmd, nil, []string{`"prompt":`, `"answer":`}},
		{guessCmd, nil, []string{`"monster":`, `"options":`, `"answer":`}},
		{mashCmd, []string{"dracula", "mummy"}, []string{`"a":`, `"rounds":`, `"winner":`}},
		{cryptCmd, nil, []string{`"captured":`, `"badges":`}},
		{tonightCmd, nil, []string{`"date": "2026-10-31"`, `"time": "9:00 PM"`, `"time": "10:45 PM"`}},
		{countdownCmd, nil, []string{`"nightsUntilHalloween": 18`, `"night": 13`, `"moon":`}},
		{packsCmd, nil, []string{`"id": "universal"`, `"monsters":`}},
	}
	for _, c := range cases {
		out := string(captureJSON(t, c.cmd, c.args))
		for _, k := range c.keys {
			if !strings.Contains(out, k) {
				t.Errorf("%s --json lacks %s:\n%s", c.cmd.Name(), k, out)
			}
		}
	}
}

// Errors carry a stable code under --json, and anything uncoded is "usage".
func TestErrorsHaveCodes(t *testing.T) {
	_, unknown := resolveMonster("draculla")
	_, ambiguous := resolveMonster("an")
	_, badDate := parseDate("31/10")
	cases := []struct {
		err  error
		code string
	}{
		{unknown, codeUnknownMonster},
		{ambiguous, codeAmbiguousMonster},
		{badDate, codeBadDate},
		{requireTerminal("x", "y"), codeNoTerminal},
		{fail(codeBadFlag, "", "--rounds must be at least 1"), codeBadFlag},
		{errors.New("something uncoded"), codeUsage},
	}
	for _, c := range cases {
		if c.err == nil {
			t.Fatalf("expected an error for code %s", c.code)
		}
		var buf bytes.Buffer
		writeErrorJSON(&buf, c.err)
		var got struct {
			Error jsonError `json:"error"`
		}
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("not one JSON object: %v\n%s", err, buf.String())
		}
		if got.Error.Code != c.code || got.Error.Message == "" {
			t.Errorf("%v -> %+v, want code %s", c.err, got.Error, c.code)
		}
		if strings.Count(buf.String(), "\n") != 1 {
			t.Errorf("the error object should be one line:\n%s", buf.String())
		}
	}
	if !wantsJSON([]string{"monster", "nope", "--json"}) || wantsJSON([]string{"monster", "nope"}) {
		t.Error("wantsJSON should read --json from the raw arguments")
	}
}

// In October every ritual stars the same monster: the Night of Fright's.
func TestOctoberRitualsShareTonightsMonster(t *testing.T) {
	t.Setenv("TERMINAL_OF_TERROR_HOME", t.TempDir())
	tonightDate, countdownDate, randomDate = "2026-10-24", "2026-10-24", "2026-10-24"
	defer func() { tonightDate, countdownDate, randomDate = "", "", ""; tonightShuffle = false }()

	var night struct {
		NightOfFright nightJSON `json:"nightOfFright"`
	}
	if err := json.Unmarshal(captureJSON(t, countdownCmd, nil), &night); err != nil {
		t.Fatal(err)
	}
	if night.NightOfFright.Night != 24 || night.NightOfFright.Monster == "" || night.NightOfFright.Fact == "" {
		t.Fatalf("countdown: %+v", night.NightOfFright)
	}
	var fact factJSON
	if err := json.Unmarshal(captureJSON(t, randomCmd, nil), &fact); err != nil {
		t.Fatal(err)
	}
	if fact.ID != night.NightOfFright.Monster || fact.Fact != night.NightOfFright.Fact {
		t.Errorf("the Fact of the Night should be tonight's monster's: %+v vs %+v", fact, night.NightOfFright)
	}
	var bill tonightJSON
	if err := json.Unmarshal(captureJSON(t, tonightCmd, nil), &bill); err != nil {
		t.Fatal(err)
	}
	if len(bill.Features) != 2 || bill.Features[0].Monster != night.NightOfFright.Monster || bill.Features[1].Monster == bill.Features[0].Monster {
		t.Errorf("tonight's bill should open with tonight's monster and a different support: %+v", bill.Features)
	}
	tonightShuffle = true
	if err := json.Unmarshal(captureJSON(t, tonightCmd, nil), &bill); err != nil {
		t.Fatal(err)
	}
	if len(bill.Features) != 2 || bill.Features[0].Monster == bill.Features[1].Monster {
		t.Errorf("--shuffle should still give two different monsters: %+v", bill.Features)
	}
}
