// Command shelldocs writes what a package manager installs beside the
// binary: a man page per command under man/, and shell completions for
// bash, zsh, fish and PowerShell under completions/. GoReleaser runs it
// before packaging (see .goreleaser.yaml); run it by hand with
// `go run ./scripts/shelldocs -out .` and read a page with
// `man -l man/terminal-of-terror.1`.
//
// The man pages are written straight from the command tree in roff, the
// few macros a man page needs, rather than through a Markdown converter,
// which would be a dependency for forty lines of formatting.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hungovercoders/terminal-of-terror/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	out := flag.String("out", ".", "directory to write man/ and completions/ into")
	flag.Parse()
	if err := Write(cmd.Root(), *out); err != nil {
		log.Fatalf("shelldocs: %v", err)
	}
}

// Write writes the man pages and completions for root under dir.
func Write(root *cobra.Command, dir string) error {
	manDir, compDir := filepath.Join(dir, "man"), filepath.Join(dir, "completions")
	for _, d := range []string{manDir, compDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	for _, c := range Commands(root) {
		if err := os.WriteFile(filepath.Join(manDir, ManName(c)+".1"), []byte(ManPage(c)), 0o644); err != nil {
			return err
		}
	}
	name := root.Name()
	gens := map[string]func(io.Writer) error{
		name + ".bash": func(w io.Writer) error { return root.GenBashCompletionV2(w, true) },
		name + ".zsh":  root.GenZshCompletion,
		name + ".fish": func(w io.Writer) error { return root.GenFishCompletion(w, true) },
		name + ".ps1":  root.GenPowerShellCompletionWithDesc,
	}
	for file, gen := range gens {
		var b bytes.Buffer
		if err := gen(&b); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(compDir, file), b.Bytes(), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// Commands lists root and every command under it that a person can run,
// in tree order: hidden commands and help are left out.
func Commands(root *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Hidden || c.Name() == "help" {
			return
		}
		out = append(out, c)
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)
	return out
}

// ManName is the page's name: the command path with dashes,
// "terminal-of-terror-crypt-reset".
func ManName(c *cobra.Command) string {
	return strings.ReplaceAll(c.CommandPath(), " ", "-")
}

// ManPage renders one command as a section 1 man page.
func ManPage(c *cobra.Command) string {
	var b strings.Builder
	name := ManName(c)
	fmt.Fprintf(&b, ".TH %q 1 \"\" \"terminal-of-terror\" \"Terminal of Terror manual\"\n", strings.ToUpper(name))
	fmt.Fprintf(&b, ".SH NAME\n%s \\- %s\n", roff(name), roff(c.Short))
	fmt.Fprintf(&b, ".SH SYNOPSIS\n.B %s\n", roff(c.UseLine()))
	desc := c.Long
	if desc == "" {
		desc = c.Short
	}
	b.WriteString(".SH DESCRIPTION\n")
	for _, para := range strings.Split(strings.TrimSpace(desc), "\n\n") {
		fmt.Fprintf(&b, ".PP\n%s\n", roffLines(para))
	}
	flags(&b, "OPTIONS", c.LocalNonPersistentFlags())
	if c == c.Root() {
		flags(&b, "GLOBAL OPTIONS", c.PersistentFlags())
	} else {
		flags(&b, "GLOBAL OPTIONS", c.InheritedFlags())
	}
	if ex := strings.TrimSpace(c.Example); ex != "" {
		fmt.Fprintf(&b, ".SH EXAMPLES\n.PP\n.nf\n%s\n.fi\n", roffLines(ex))
	}
	var see []string
	if c.HasParent() {
		see = append(see, ManName(c.Parent())+"(1)")
	}
	for _, sub := range c.Commands() {
		if !sub.Hidden && sub.Name() != "help" {
			see = append(see, ManName(sub)+"(1)")
		}
	}
	if len(see) > 0 {
		fmt.Fprintf(&b, ".SH SEE ALSO\n%s\n", strings.Join(see, ", "))
	}
	b.WriteString(".SH WEBSITE\nhttps://terminalofterror.com/\n")
	return b.String()
}

func flags(b *strings.Builder, title string, fs *pflag.FlagSet) {
	if !fs.HasAvailableFlags() {
		return
	}
	fmt.Fprintf(b, ".SH %s\n", title)
	fs.VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		names := "\\-\\-" + f.Name
		if f.Shorthand != "" {
			names = "\\-" + f.Shorthand + ", " + names
		}
		if f.Value.Type() != "bool" {
			names += " " + strings.ToUpper(f.Value.Type())
		}
		usage := f.Usage
		if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" && f.DefValue != "[]" {
			usage += " (default " + f.DefValue + ")"
		}
		fmt.Fprintf(b, ".TP\n.B %s\n%s\n", names, roff(usage))
	})
}

// roff escapes text for a man page: backslashes, and a leading dot or
// quote, which roff would read as a macro.
func roff(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	if strings.HasPrefix(s, ".") || strings.HasPrefix(s, "'") {
		s = "\\&" + s
	}
	return s
}

func roffLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = roff(l)
	}
	return strings.Join(lines, "\n")
}
