package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/hungovercoders/terminal-of-terror/internal/crypt"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/store"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
)

// record saves an outcome to the player's progress and announces anything
// new. Progress problems are reported but never stop the fun.
func record(o crypt.Outcome) {
	p, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Couldn't read your progress, so this session won't be saved: %v\n", err)
		return
	}
	if o.Now.IsZero() {
		o.Now = time.Now()
	}
	// Badges like Master of the Crypt count every monster, not just a --pack selection.
	u := crypt.Apply(p, monsters.EveryMonster(), o)
	if err := p.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Couldn't save your progress: %v\n", err)
		return
	}
	if s := ui.RenderUnlocks(u); s != "" {
		fmt.Print("\n" + s)
	}
}
