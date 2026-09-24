// Package host gives a voice to Count Cathode, Terminal of Terror's
// late-night horror host, broadcasting from beyond the static.
package host

import (
	"fmt"
	"math/rand"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

// Name is the host's on-air name.
const Name = "Count Cathode"

// Channel is the fictional station the host broadcasts on.
const Channel = "Channel 13"

var greetings = []string{
	"Good evening, creatures of the night, and welcome to the Channel 13 Creature Feature. I'm your host, Count Cathode, broadcasting from beyond the static.",
	"Pull up a coffin and dim the lights. The rabbit ears are up, the popcorn is stale, and the monsters are restless.",
	"Ah, you're back. I knew you would be. Nobody ever really leaves the Creature Feature.",
	"Welcome, insomniacs and night owls. Adjust your antenna, lock the cellar door, and let's meet tonight's guests.",
}

var signOffs = []string{
	"That's all for tonight. Sleep tight, and check under the bed. Count Cathode, signing off.",
	"Don't touch that dial... actually, do. It's late. Good night, fiends!",
	"The test card is coming up, which means it's time for bed. Pleasant screams!",
	"Until next time, keep your stakes sharp and your garlic fresh. Goodnight from Channel 13.",
}

var quips = []string{
	"Write that one down. There'll be a test. Probably at midnight.",
	"Now you know. And knowing is half the haunting.",
	"Tell that one at your next séance.",
	"Frightfully educational, isn't it?",
	"Chilling, simply chilling.",
}

// Greeting returns a random opening line.
func Greeting(r *rand.Rand) string { return pick(r, greetings) }

// SignOff returns a random closing line.
func SignOff(r *rand.Rand) string { return pick(r, signOffs) }

// Quip returns a short reaction to a fact.
func Quip(r *rand.Rand) string { return pick(r, quips) }

// Intro is the host's introduction for a monster.
func Intro(m monsters.Monster) string {
	if m.HostIntro != "" {
		return m.HostIntro
	}
	return fmt.Sprintf("Tonight's feature: %s. Viewer discretion is advised.", m.Name)
}

func pick(r *rand.Rand, lines []string) string {
	if r == nil {
		return lines[rand.Intn(len(lines))]
	}
	return lines[r.Intn(len(lines))]
}
