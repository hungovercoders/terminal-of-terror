# Night 1 · It's Alive!

> 📺 *"Good evening, and welcome to the first night of Night School. Tonight we do what Victor Frankenstein did: assemble something from parts and shout when it moves. Our parts are a text editor, a terminal and the Go toolchain. Our shout will be quieter."*

**Tonight you'll learn**
- What Go is and why it suits command-line tools
- How to install Go and check it works
- What a *module* is, and `go mod init`
- Your first program: `package main`, `func main`, and `fmt`
- `go run` versus `go build`

**Where we are:** an empty folder.

## Why Go for a terminal program?

Go compiles to a single file with no dependencies. You build `terminal-of-terror`, hand it to a friend, and it runs: no runtime to install, no `node_modules`, no virtual environments. It builds for Windows, macOS and Linux from one machine, it starts in milliseconds, and its standard library already covers JSON, files, testing and time. Nearly every popular terminal tool of the last decade is written in it.

Go is also a small language. There are 25 keywords. You can learn all the syntax in a week and spend the rest of your life learning how to use it well, which is the right way round.

## Install Go

Download the installer for your system from [go.dev/dl](https://go.dev/dl/) and run it. Then open a *new* terminal and check:

```bash
go version
```

You should see `go version go1.24.x` or newer. If the command isn't found, the installer didn't put Go on your `PATH`; the [install page](https://go.dev/doc/install) has the fix for each system.

## A place for the monsters

Make a folder and step into it:

```bash
mkdir terminal-of-terror
cd terminal-of-terror
```

Now tell Go this folder is a **module**, a unit of code with a name and a list of the libraries it uses:

```bash
go mod init github.com/hungovercoders/terminal-of-terror
```

This creates `go.mod`:

```
module github.com/hungovercoders/terminal-of-terror

go 1.24
```

The module name looks like a web address because Go modules are usually published on sites like GitHub, and the name is where other people would fetch them from. Nothing is being published tonight. Using this exact name means every `import` in the course matches your code; on Night 31 you'll see how to rename it.

## The first program

Create `main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("🎃 It's alive!")
}
```

Three things to know:

- **`package main`** marks this as a program rather than a library. Every Go file starts with a `package` line.
- **`import "fmt"`** brings in the *format* package from the standard library, which does printing.
- **`func main()`** is where the program starts. Go looks for exactly this function in `package main`.

Go is fussy about formatting and has a tool that settles every argument: `gofmt`. Editors with a Go extension run it on save. Tabs for indentation, braces on the same line, no semicolons. You'll never think about it again.

Run it:

```bash
go run .
```

```
🎃 It's alive!
```

`go run .` compiles the program in the current folder (`.`) and runs it in one step. It's the command you'll use most while developing.

## Build a real binary

```bash
go build -o terminal-of-terror
./terminal-of-terror
```

On Windows, the file is `terminal-of-terror.exe` and you run `.\terminal-of-terror.exe`.

That file is the whole program. Copy it to another computer with the same operating system and it runs. It's about two megabytes, most of which is the Go runtime (memory management, the scheduler) that every Go program carries.

## Save your work

Tell Git to ignore the binary, then make your first commit:

```bash
git init
printf 'terminal-of-terror\nterminal-of-terror.exe\n' > .gitignore
git add .
git commit -m "feat: it's alive"
```

The commit message follows a convention called [Conventional Commits](https://www.conventionalcommits.org): a type (`feat`, `fix`, `docs`, `chore`), a colon, and a short description. On Night 31 a tool will turn these messages into a changelog, so it pays to start now.

## Try it

- Change the message. Run `go run .` again.
- Break something: delete the closing `}` and run it. Read the error. Go's compiler errors say the file, the line and what it expected; they're your friend for the next 30 nights.
- Add a second line: `fmt.Println("Tonight's feature:", "Dracula")`. Notice `Println` puts a space between its arguments.

## 💀 Terrifying fact

`go run .` doesn't leave a binary behind, so where did the program go? Into a cache directory, along with every package you'll ever compile. `go env GOCACHE` shows where. Go's compiler is fast partly because it never compiles the same package twice.

## 🕯️ Before dawn

Make `main.go` print a three-line title card:

```
==========================
   TERMINAL OF TERROR
==========================
```

Hint: a string can contain `\n` for a new line, and `strings.Repeat("=", 26)` from the `strings` package makes the rule. You'll need to import `"strings"` as well as `"fmt"`: group the imports in parentheses, one per line.

> 📺 *"It moved. Dr Frankenstein would be proud, and possibly arrested. Tomorrow night we give it a skeleton."*

[Index](README.md) · [Night 2 →](night-02.md)
