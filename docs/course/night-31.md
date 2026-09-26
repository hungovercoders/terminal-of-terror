# Night 31 · Halloween: Sign Off

> 📺 *"Halloween. The big night. Thirty nights ago you had an empty folder; tonight you have a program with nineteen monsters, three games, a crypt, a moon, a camera and a night watchman. One thing left: give it a number, write down what it does, and let it out. Then we sign off. Don't touch that dial."*

**Tonight you'll learn**
- Semantic versioning, and the module-path rule for v2
- Where a version number comes from: ldflags, build info, or "dev"
- A changelog people can read
- GoReleaser: one config, six binaries
- A release workflow that refuses to publish without notes
- Tagging, and why a tag is not a commit

**Where we are:** everything works, everything is tested, everything is pictured. Ship it.

## A number

[Semantic versioning](https://semver.org): `MAJOR.MINOR.PATCH`. Fixes bump the patch (1.0.1), new features the minor (1.1.0), breaking changes the major. For a program people run rather than a library they import, "breaking" means a command or flag stops working the way it did.

Go adds a rule of its own: a module at v2 or later must have `/v2` at the end of its module path. Without it, `go install ...@latest` *ignores* v2 tags entirely, and users get 1.x forever. So the project's rule, in CONTRIBUTING.md and AGENTS.md: stay below 2.0.0 unless you also change the module path. It's the kind of rule you write down because it's surprising and it only bites once a year.

## Where the number lives

There are three ways the program can learn its own version, and [`cmd/version.go`](../../cmd/version.go) tries them in order:

```go
// version is set at release time by GoReleaser:
//
//	-ldflags "-X github.com/hungovercoders/terminal-of-terror/cmd.version=1.0.0"
var version string

// appVersion is the version to report. Release builds carry it in version.
// Other builds use Go's build info: `go install ...@v1.0.0` reports 1.0.0,
// and a local `go build` reports a pseudo-version naming the commit. "dev"
// is the last resort when neither is available.
func appVersion() string {
	if version != "" {
		return strings.TrimPrefix(version, "v")
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return strings.TrimPrefix(v, "v")
		}
	}
	return "dev"
}
```

**`-X`** is a linker flag that sets a package-level string variable at build time, so a release build is stamped with its tag without a source change. **`debug.ReadBuildInfo`** reads what the Go toolchain embedded in the binary: for `go install ...@v1.0.0` that's the version; for a plain `go build` in a checkout it's `(devel)`. And `"dev"` for anything else. `rootCmd.Version` is set from this, which is what makes `--version` work; Cobra adds the flag when `Version` is non-empty.

Three sources, one function, and the comment explains which build gets which. Version reporting is a small feature that gets asked about in every bug report, so it's worth doing completely.

## The changelog

[`CHANGELOG.md`](../../CHANGELOG.md) follows [Keep a Changelog](https://keepachangelog.com): a section per version, newest first, headed `## [1.0.0] - 2026-09-24`, with what changed grouped for the *reader*, not the committer. The 1.0.0 section is grouped as The Creature Feature, Games and Nightly rituals, with a quote from the host, because that's how a player thinks of the program. `git cliff` can draft a section from the conventional commit messages you've been writing since Night 1 (`feat:`, `fix:`, `docs:`); then you edit it until it reads well. A generated changelog is a starting point, not a changelog.

The changelog is also *enforced*: the release workflow reads the section for the tag and refuses to publish without one.

## GoReleaser

Building for Linux, macOS and Windows, on Intel and ARM, is six `go build` invocations with the right `GOOS` and `GOARCH`, six archives, a checksum file and a GitHub release with the archives attached. [GoReleaser](https://goreleaser.com) does it from one file, [`.goreleaser.yaml`](../../.goreleaser.yaml):

```yaml
builds:
  - main: .
    binary: terminal-of-terror
    env:
      - CGO_ENABLED=0
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    flags:
      - -trimpath
    ldflags:
      - -s -w -X github.com/hungovercoders/terminal-of-terror/cmd.version={{ .Version }}
    mod_timestamp: "{{ .CommitTimestamp }}"
```

`CGO_ENABLED=0` makes a static binary that runs anywhere without a C library. `-s -w` strips debug symbols for a smaller file; `-trimpath` removes the build machine's paths; `mod_timestamp` makes builds reproducible, so two people building the same tag get byte-identical binaries. And there's the `-X` flag with `{{ .Version }}` filled in from the tag.

```yaml
archives:
  - formats: [tar.gz]
    format_overrides:
      - goos: windows
        formats: [zip]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    files:
      - README.md
      - LICENSE
      - CHANGELOG.md

changelog:
  disable: true
```

Tarballs, except a zip for Windows, named so a user can find theirs, with the README, licence and changelog inside. GoReleaser's own changelog is disabled because the notes come from CHANGELOG.md. Night 29's fourth CI check is `goreleaser check`, which validates this file on every pull request, so a typo here is found before release day.

## The release workflow

[`release.yml`](../../.github/workflows/release.yml) runs on a tag push:

```yaml
on:
  push:
    tags: ["v*"]

permissions:
  contents: write
```

`contents: write`, because this one *does* publish. Then the tests, then the notes:

```yaml
      - name: Release notes from CHANGELOG.md
        run: |
          version="${GITHUB_REF_NAME#v}"
          awk -v heading="## [$version]" '
            index($0, heading) == 1 { found = 1; next }
            found && /^## \[/ { exit }
            found { print }
          ' CHANGELOG.md > "$RUNNER_TEMP/notes.md"
          if [ ! -s "$RUNNER_TEMP/notes.md" ]; then
            echo "::error::CHANGELOG.md has no '## [$version]' section. Update it before tagging."
            exit 1
          fi
```

An `awk` program that prints the lines between this version's heading and the next one. If that's empty, the release stops with an error that says what to do, and nothing is published: no release with no notes. Then `goreleaser release --clean --release-notes notes.md` does the rest with the repository's own token.

## Tagging

A **tag** is a name for a commit. `git tag v1.0.0` on the merge commit of the release preparation, then `git push origin v1.0.0`, and the workflow runs. Tags aren't pushed by `git push` on their own, which is why the step is explicit, and why the project's rule is that a tag is never pushed without the maintainer's go-ahead: pushing a tag *is* publishing.

The CONTRIBUTING.md release checklist, in full:

1. Pick the version.
2. `git cliff --unreleased --tag v1.1.0 --prepend CHANGELOG.md`, then edit.
3. Commit (`chore(release): prepare for v1.1.0`) and merge to `main`.
4. `git tag v1.1.0 && git push origin v1.1.0` from an up-to-date `main`.

Try it without publishing: `goreleaser release --snapshot --clean` builds everything into `dist/` and touches nothing on GitHub.

## Run it

```bash
go build -o terminal-of-terror . && ./terminal-of-terror --version
go install github.com/hungovercoders/terminal-of-terror@latest && terminal-of-terror --version
```

The first says `dev` (or a pseudo-version); the second, the tagged release. Then the last command of the course:

```bash
terminal-of-terror monster
```

Watch the static. Read the greeting. It's yours.

## Try it

- Run `goreleaser release --snapshot --clean` and look in `dist/`. Unpack the Windows zip on a Mac. It's just files.
- Change `formats: [tar.gz]` to add `zip` for everyone. Some users prefer it.
- Make the changelog check *also* require a date in the heading.

## 💀 Terrifying fact

`go install github.com/hungovercoders/terminal-of-terror@latest` doesn't ask GitHub; it asks the Go module proxy, which caches modules and *never forgets a version*. Once a tag has been fetched through the proxy, deleting or moving the tag on GitHub doesn't remove it: users will keep getting the original. A published version is immutable in practice, which is why the release rules are strict, and why a mistake is fixed by publishing 1.0.1, never by re-tagging 1.0.0.

## 🕯️ Before dawn

Thirty-one nights. Here's what you built, in the order you built it: a module and a command; a data vault with tests; a search; a styled list and a card; an interactive explorer with pages, scrolling and search; a host and an intro; a silent-film mode; a gallery; a question generator and two games; saved progress, captures and badges; a fight simulator; a moon; three rituals; JSON and seeds and packs from disk; a test suite that runs on three operating systems; a camera; and a release.

The techniques are the point: Cobra commands, embedded data, the Elm architecture, Lip Gloss layout, table-driven tests, seeded randomness, atomic writes, a CI matrix, GoReleaser. None of them is about monsters. All of them are yours for the next program.

Your homework is a program of your own, on any subject, hosted by anyone you like. Thirty-one nights from now, tag it v1.0.0.

> 📺 *"That's all for tonight, and that's all for the course. Sleep tight, check under the bed, and keep your garlic fresh. From all of us at Channel 13: Happy Halloween, creatures of the night. Count Cathode, signing off."*

[← Night 30](night-30.md) · [Index](README.md)
