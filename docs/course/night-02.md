# Night 2 · The Skeleton Crew

> 📺 *"A monster needs bones before it needs a face. Tonight we give our program a skeleton: commands, flags, a help screen and a version number, courtesy of a library named after a snake. Don't worry. It's a friendly snake."*

**Tonight you'll learn**
- Adding a dependency with `go get`
- What Cobra is and how it structures a command-line tool
- A root command with `--help` and `--version`
- Splitting code into packages: `main` and `cmd`
- Two settings that make errors friendlier

**Where we are:** `main.go` prints a title card.

## Commands and subcommands

Real terminal tools work like this:

```
git commit -m "message"
docker run --rm ubuntu
terminal-of-terror monster dracula
```

A *root command* (`git`), a *subcommand* (`commit`), *flags* (`-m`) and *arguments* (`"message"`). You could parse all that yourself from `os.Args`, and on a bad night you might. Instead we'll use [Cobra](https://github.com/spf13/cobra), the library behind `kubectl`, `gh`, `hugo` and hundreds of others. It parses everything, generates the help screens and, on Night 31, gives us `--version` for free.

## Add the dependency

```bash
go get github.com/spf13/cobra@v1.10.1
```

`go get` downloads the library, records it in `go.mod` and writes a checksum into `go.sum`. Commit both files: `go.sum` is how Go guarantees that everyone building your program gets exactly the same library bytes.

Look at `go.mod` now. Cobra has brought two friends with it (`pflag` and `mousetrap`), marked `// indirect`. That's normal.

## The `cmd` package

We'll keep every command in its own file in a folder called `cmd`. Create `cmd/root.go`:

```go
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "terminal-of-terror",
	Short: "A terminal tool that terrifies you with universal monsters!",
	Long: `Terminal of Terror brings classic Universal monsters to your terminal.

Learn the real history behind Dracula, Frankenstein's Monster, the Wolf Man
and more, presented by your late-night horror host, Count Cathode.`,
	Version: "0.1.0",
}

// Execute runs the root command. main calls it and nothing else.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Cobra adds a "completion" subcommand by default; we don't need it yet.
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// When a command fails, print only the error, not the whole usage text.
	rootCmd.SilenceUsage = true
}
```

New Go here:

- **`package cmd`**: a second package. Files in the same folder share a package and can see each other's variables. Other packages see only names that start with a capital letter (`Execute`), which is Go's whole visibility system: capital means exported.
- **`&cobra.Command{ ... }`**: creating a struct value and taking a pointer to it. The `Use:`, `Short:` lines are *fields*. We'll write our own structs tomorrow.
- **`func init()`**: runs automatically when the package loads, before `main`. Cobra programs use it to wire commands together.
- **`if err := ...; err != nil`**: Go's error idiom. Functions return errors as ordinary values, and you check them right there. No exceptions, no try/catch.

`Execute` returns an error when the user types something Cobra can't parse. We print nothing extra because Cobra already has; we just exit with status 1 so scripts know it failed.

## Point `main` at it

Replace `main.go`:

```go
package main

import "github.com/hungovercoders/terminal-of-terror/cmd"

func main() {
	cmd.Execute()
}
```

The import path is your module name plus the folder. That's the whole reason the module name matters.

## Run it

```bash
go mod tidy
go run .
```

`go mod tidy` looks at what your code actually imports and sorts out `go.mod`: Cobra moves from the `// indirect` list to the direct one now that `cmd/root.go` imports it. Run it whenever you add or remove an import.

With no arguments, Cobra prints the description:

```
Terminal of Terror brings classic Universal monsters to your terminal.

Learn the real history behind Dracula, Frankenstein's Monster, the Wolf Man
and more, presented by your late-night horror host, Count Cathode.
```

No "Usage:" section yet. Cobra only shows one when there's something to use, a function to run or subcommands to list, and we have neither until tomorrow.

Now try:

```bash
go run . --version
go run . --nonsense
```

The first prints `terminal-of-terror version 0.1.0`. The second prints `Error: unknown flag: --nonsense` and exits with status 1, so a script calling us would know it failed. Without `SilenceUsage`, Cobra would also dump the whole usage text under the error, which buries the one line that matters.

## Try it

- Change `Short`. It doesn't show up anywhere yet: `Short` is the one-line summary a command gets in its parent's list, and the root has no parent. From tomorrow, every subcommand's `Short` appears in `--help`.
- Set `SilenceUsage` to `false` and run `go run . --nonsense` again. See the difference.
- Run `go doc github.com/spf13/cobra Command` to read the fields a command can have. `go doc` works for any package you've downloaded.

## 💀 Terrifying fact

Go has no `public` or `private` keywords. Capitalise a name to export it, lowercase it to keep it inside the package. That's it. It means you can tell at a glance, anywhere in a program, whether `execute()` is local or `Execute()` belongs to the outside world.

## 🕯️ Before dawn

Give the root command an `Example:` field with two lines showing how you'd run the tool (look at `go doc cobra.Command` for the field). Then find out where Cobra prints it.

> 📺 *"Bones in place. It stands up, it says its name, and it knows its own version. Tomorrow night, we let it meet the monsters."*

[← Night 1](night-01.md) · [Index](README.md) · [Night 3 →](night-03.md)
