# Night 3 · Meet the Monsters

> 📺 *"Tonight's guests need no introduction, but they'll get one anyway, because introductions are what tonight is about. We're going to describe a monster to a computer. It won't scream. Computers never do; that's what makes them unsettling."*

**Tonight you'll learn**
- Structs: describing a thing with named fields
- Slices: lists that grow
- `for … range` loops
- `fmt.Printf` and format verbs
- The `internal` folder, and your first subcommand

**Where we are:** a root command with `--help` and `--version`.

## Describing a monster

A monster has a name, a description, an origin and some facts. In Go, a thing with named parts is a **struct**. Create `internal/monsters/monsters.go`:

```go
// Package monsters knows every creature in the vault.
package monsters

// Monster is a single creature and what we know about it.
type Monster struct {
	Name        string
	Description string
	Origin      string
	Facts       []string
}
```

`string` is text. `[]string` is a **slice** of strings: an ordered list that can hold any number of them. Every field starts with a capital letter, so other packages can read them.

The comment above `package monsters` is the package's documentation, and the one above `Monster` documents the type. `go doc ./internal/monsters` prints them. Go's convention is that a doc comment starts with the name it describes; it reads oddly at first and then becomes second nature.

### Why `internal`?

Go treats a folder named `internal` specially: packages inside it can only be imported by code in the same module. Our `cmd` package can import `internal/monsters`; a stranger's program can't. It's how a program keeps its insides private while still splitting into packages.

## The vault

Below the struct, add the data. For now it lives in the code; on Night 5 it moves out.

```go
// The vault. Tonight it lives in code; on Night 5 it moves to JSON.
var monsters = []Monster{
	{
		Name:        "Dracula",
		Description: "The legendary vampire count from Transylvania",
		Origin:      "Bram Stoker's novel (1897)",
		Facts: []string{
			"He casts no reflection, as Jonathan Harker discovers while shaving",
			"Stoker took the name from Vlad III of Wallachia, but little else",
			"Bela Lugosi played the Count on Broadway in 1927 before the film",
		},
	},
	{
		Name:        "Frankenstein's Monster",
		Description: "The tragic creature created by Dr. Victor Frankenstein",
		Origin:      "Mary Shelley's novel (1818)",
		Facts: []string{
			"The creature is never named in the novel; Frankenstein is his creator",
			"He teaches himself to read by secretly watching a cottage family",
			"Boris Karloff was credited only as '?' in the 1931 film's opening titles",
		},
	},
	{
		Name:        "The Wolf Man",
		Description: "A man cursed to transform into a werewolf",
		Origin:      "The Wolf Man (1941 film)",
		Facts: []string{
			"Lon Chaney Jr. played Larry Talbot in all five of his Universal films",
			"In the 1941 film he is killed with a silver-headed cane, not a silver bullet",
			"The film opened just five days after the attack on Pearl Harbor",
		},
	},
}

// GetAllMonsters returns every monster in the vault.
func GetAllMonsters() []Monster {
	return monsters
}
```

Things to notice:

- `[]Monster{ {...}, {...} }` is a slice of three structs, written out in full. Inside a slice of `Monster`, Go lets you omit the type name from each element.
- Every field is `Name: value,` with a **trailing comma**, even after the last one. Go requires it when the closing brace is on its own line. `gofmt` will remind you.
- The variable `monsters` is lowercase, so it's private to the package. `GetAllMonsters` is the public door. Later we'll want to load monsters from files and filter them, and everything outside the package will keep working because it only ever used the door.

## The `list` command

Create `cmd/list.go`:

```go
package cmd

import (
	"fmt"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List every monster in the vault",
	Run: func(cmd *cobra.Command, args []string) {
		all := monsters.GetAllMonsters()
		fmt.Println("Universal Monsters")
		fmt.Println("==================")
		for i, m := range all {
			fmt.Printf("%2d. %-24s %s\n", i+1, m.Name, m.Description)
		}
		fmt.Printf("\n%d monsters lurk in the vault.\n", len(all))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
```

- **`Run:`** is the function Cobra calls when someone types `list`. It receives the command and any arguments; we'll use `args` on Night 8.
- **`:=`** declares a new variable and works out its type from the right-hand side. `all` is a `[]Monster`.
- **`for i, m := range all`** loops over the slice. `i` is the position (from 0), `m` is a copy of the element. If you don't need the index, write `for _, m := range all`: the underscore means "I know there's a value here, throw it away". Go refuses to compile an unused variable, so you'll see `_` a lot.
- **`Printf`** formats with verbs: `%d` for a number, `%s` for a string, `%2d` pads the number to 2 characters, `%-24s` pads the name to 24 characters, left-aligned (the `-`). `\n` ends the line, because unlike `Println`, `Printf` doesn't add one.
- **`rootCmd.AddCommand(listCmd)`** in `init` hangs `list` under the root. The root and `list` are in the same package, so `rootCmd` is visible here.

## Run it

```bash
go run . list
```

```
Universal Monsters
==================
 1. Dracula                  The legendary vampire count from Transylvania
 2. Frankenstein's Monster   The tragic creature created by Dr. Victor Frankenstein
 3. The Wolf Man             A man cursed to transform into a werewolf

3 monsters lurk in the vault.
```

And now `go run .` alone shows a proper usage block, because the root has something to list:

```
Usage:
  terminal-of-terror [command]

Available Commands:
  help        Help about any command
  list        List every monster in the vault
```

There's `Short` from last night, doing its job.

## Try it

- Add a fourth monster. The Mummy (1932 film, played by Boris Karloff) is waiting.
- Change `%-24s` to `%24s` and see the names jump to the right.
- Print the facts too: inside the loop, add a second loop over `m.Facts` that prints each one indented with `"    - "`.

## 💀 Terrifying fact

`m` in `for _, m := range all` is a *copy* of each monster. Changing `m.Name` inside the loop changes the copy and leaves the slice alone. Go copies structs when you assign them, pass them to functions or range over them. It's why Go code that needs to *change* a thing passes a pointer, which you'll see on Night 8.

## 🕯️ Before dawn

Add a `Year int` field to `Monster` for the year of the monster's first appearance, fill it in for each monster, and print it in `list` after the name in brackets, like `Dracula (1897)`. Then sort the list by year: `sort.Slice(all, func(i, j int) bool { return all[i].Year < all[j].Year })` from the `sort` package. Notice that sorting `all` sorts the vault's own slice, because a slice is a *view* onto shared storage, not a copy. Tomorrow's terrifying fact explains.

> 📺 *"Three guests, properly introduced. They're a little stiff, but so was Karloff until the electrodes went in. Tomorrow night, we let chance pick one."*

[← Night 2](night-02.md) · [Index](README.md) · [Night 4 →](night-04.md)
