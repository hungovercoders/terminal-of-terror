# Night 29 · The Night Watch

> 📺 *"Nothing works on Windows. That's not an opinion; it's a law of nature, like sunrise, and the way to repeal it is to run your tests on Windows every single time. Tonight we hire a night watchman: a workflow that formats, vets, tests and smoke-tests the program on three operating systems whenever anyone opens a pull request."*

**Tonight you'll learn**
- What continuous integration is for
- A GitHub Actions workflow: triggers, jobs, steps, a matrix
- Separating fast checks from slow ones
- Keeping the runner's real config directory out of the tests
- A smoke test that runs every command
- Reading a red build

**Where we are:** tests pass on your machine.

## "Works on my machine"

Every project has the moment where a change that works for its author breaks for someone else: a different OS, a missing `go mod tidy`, a file that was never committed. **Continuous integration** is running the checks on a clean machine, automatically, for every change, so that moment happens in a pull request and not in a release.

GitHub Actions is the CI built into GitHub. A **workflow** is a YAML file in `.github/workflows/`; GitHub runs it on the events it names. The project's is [`ci.yml`](../../.github/workflows/ci.yml).

## Triggers and permissions

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read

concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true
```

It runs on every pull request and on pushes to `main`. `permissions: contents: read` is the least the workflow needs: it reads the code and writes nothing, so a compromised dependency in the build can't push to the repository. `concurrency` cancels an older run of the same branch when a newer push arrives, which saves minutes and money when you push three fixes in a row.

## Two jobs

```yaml
jobs:
  lint:
    name: Format, vet and tidy
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: gofmt
        run: |
          unformatted=$(gofmt -l .)
          if [ -n "$unformatted" ]; then
            echo "These files need gofmt:"
            echo "$unformatted"
            exit 1
          fi

      - name: go vet
        run: go vet ./...

      - name: go mod tidy
        run: |
          go mod tidy
          git diff --exit-code go.mod go.sum || (echo "Run 'go mod tidy' and commit the result." && exit 1)
```

Night 28's table stakes, as a job. Each step is a shell command; a non-zero exit fails the step, the job and the check on the pull request. `gofmt -l` prints nothing when everything's formatted, so the step turns "printed something" into a failure *with the list of files*. The tidy check runs `go mod tidy` and asks git whether anything changed. Every failure message says what to do, because the person reading it is not the person who wrote the workflow.

`go-version-file: go.mod` installs whatever Go version the module declares, so CI and the developer agree, and upgrading Go is one edit.

```yaml
  test:
    name: Test (${{ matrix.os }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    env:
      # Keep progress and community packs out of the runner's real config dir.
      TERMINAL_OF_TERROR_HOME: ${{ github.workspace }}/.tot-home
```

A **matrix** runs the job once per value: three operating systems, three runs, in parallel. `fail-fast: false` lets the other two finish when one fails, so a Windows-only problem shows as Windows-only. And that `env` line is Night 21's environment variable doing exactly what it was made for: the tests and the smoke test write progress and packs into the workspace, not the runner's config directory.

```yaml
      - name: Unit tests (with race detector)
        if: runner.os == 'Linux'
        run: go test -race ./...

      - name: Unit tests
        if: runner.os != 'Linux'
        run: go test ./...
```

The race detector on Linux only: it's the slowest step, and a race is a race on any OS. `if:` conditions pick which step runs.

## The smoke test

Unit tests exercise packages. Nothing so far runs the *binary*, and a binary can fail in ways packages can't: a command not registered, a flag that panics on parse, an `os.Exit` in the wrong place.

```yaml
      - name: Build
        run: go build -o terminal-of-terror${{ runner.os == 'Windows' && '.exe' || '' }} .

      - name: Smoke test the commands
        shell: bash
        run: |
          set -euo pipefail
          bin=./terminal-of-terror
          [ "$RUNNER_OS" = "Windows" ] && bin=./terminal-of-terror.exe
          $bin --help > /dev/null
          $bin --version
          $bin list > /dev/null
          $bin list --json > /dev/null
          $bin random --daily --date 2026-10-31 > /dev/null
          $bin monster dracula --json > /dev/null
          $bin quiz -n 3 --json > /dev/null
          $bin countdown --date 2026-10-13 > /dev/null
          $bin tonight --date 2026-10-31 > /dev/null
          $bin mash dracula mummy --fast > /dev/null
          $bin crypt > /dev/null
          $bin packs new ci-pack > /dev/null
          $bin list --pack ci-pack > /dev/null
          $bin packs > /dev/null
          # Unknown names should fail cleanly rather than crash.
          if $bin monster zombie-accountant 2> /dev/null; then
            echo "expected an error for an unknown monster"
            exit 1
          fi
```

Every command, once, in a mode that needs no terminal: `--json` for the interactive ones, `--fast` for the mash, `--date` so the output doesn't depend on the day. `set -euo pipefail` makes the script stop at the first failure. The last block checks that a *bad* input fails, because a command that exits 0 on garbage is a bug too. Notice how many of Week 4's flags exist partly so this list could be written: a program that can be smoke-tested is a program that was designed to be driven.

`shell: bash` on all three runners, so one script works on Windows too. The `.exe` dance is the price of a matrix; it's paid once, here.

## Reading a red build

When a check fails, the pull request shows a red cross, and clicking through gives the log of the failing step. The workflow is written so that the last lines of that log say what to do: "These files need gofmt", "Run 'go mod tidy' and commit the result", a Go test failure with its `t.Fatalf` message. Write your CI for the person who'll read it at 11pm having broken it.

The other habit: fix red builds *first*. A red `main` hides every new failure behind the old one.

## Run it

You can't run Actions locally without extra tools, but the smoke test is plain bash:

```bash
go build -o terminal-of-terror . && TERMINAL_OF_TERROR_HOME=/tmp/ci-home bash -c '
set -euo pipefail
bin=./terminal-of-terror
$bin --help > /dev/null && $bin list --json > /dev/null && $bin quiz -n 3 --json > /dev/null
echo smoke ok'
```

Then push a branch, open a pull request, and watch the four checks go green. (The fourth is GoReleaser's config check, which belongs to Night 31.)

## Try it

- Add a step that uploads the built binary as an artifact (`actions/upload-artifact@v4`). Download it from the run page.
- Break formatting on purpose, push, and read the failure. Fix it with `gofmt -w`.
- What does the smoke test *not* cover? Write down two things. (Hint: everything interactive.)

## 💀 Terrifying fact

`actions/checkout@v4` pins a *major* version, which floats to the latest v4. Some projects pin to a full commit SHA instead, because an action is code that runs with your repository's permissions, and a compromised tag could run anything. The `permissions: contents: read` line is the other half of that defence: even if an action misbehaved, it couldn't write. Least privilege isn't only for servers.

## 🕯️ Before dawn

Add a new command to the smoke test list *before* it exists (say, `$bin moon`). Push, watch it fail, then write the command. That's test-first at the CI level, and it's a surprisingly good way to decide what a command's non-interactive mode should be.

> 📺 *"Three operating systems, every change, every time. Tomorrow night we point a camera at the whole thing, because a README without a picture is a haunted house with the lights off."*

[← Night 28](night-28.md) · [Index](README.md) · [Night 30 →](night-30.md)
