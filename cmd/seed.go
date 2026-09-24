package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// envSeed fixes every random choice (questions, portraits, fights, host
// lines) so a run can be repeated exactly. docs/demos/render.sh sets it so
// re-recording the README demos gives the same games every time.
const envSeed = "TERMINAL_OF_TERROR_SEED"

// seed returns the seed from TERMINAL_OF_TERROR_SEED, or the clock.
func seed() int64 {
	if s := os.Getenv(envSeed); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
		fmt.Fprintf(os.Stderr, "⚠️  ignoring %s=%q: it should be a whole number\n", envSeed, s)
	}
	return time.Now().UnixNano()
}

// newRand returns a random source seeded by seed().
func newRand() *rand.Rand {
	return rand.New(rand.NewSource(seed()))
}
