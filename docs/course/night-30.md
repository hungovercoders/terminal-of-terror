# Night 30 · Lights, Camera

> 📺 *"Nobody installs a terminal program because of a paragraph. They install it because of a GIF of the thing doing something, twenty seconds long, in the README. Tonight we build a camera: a script that records every demo from a tape, with the same seed, the same theme and the same fake home directory, so that the pictures are as reproducible as the tests."*

**Tonight you'll learn**
- VHS: recording a terminal from a script instead of a screen
- Why demos must be reproducible, and how the seed makes them so
- A recording environment: fake home, fixed prompt, sample progress
- Working around tools: Unicode widths in ttyd, encoding with ffmpeg
- A README that reads like a TV listing

**Where we are:** CI is green. The program has no pictures.

## Tapes

[VHS](https://github.com/charmbracelet/vhs), from the same people as Bubble Tea, records a terminal from a *tape*: a text file of settings and keystrokes. [`docs/demos/quiz.tape`](../../docs/demos/quiz.tape):

```
Output docs/demos/.frames/quiz/
Set Shell bash
Set FontSize 15
Set Width 1000
Set Height 560
Set Padding 24
Set TypingSpeed 60ms
Set Theme @theme.json
Set Framerate 50

Sleep 400ms
Type "terminal-of-terror quiz -n 3"
Sleep 300ms
Enter
Sleep 2.5s
Type "1"
Sleep 3s
Enter
...
```

A tape is a test script for humans: it types a command, waits for the screen, presses a key, waits again. Because it's text, it's in version control, it's diffable, and anyone can re-record it. There are nine tapes: a long *hero* GIF for the top of the README, one per game, and five *stills* (`tonight`, `countdown`, `random`, `list`, `crypt`) that keep only their final frame as a PNG.

## The tape presses `1`. Is `1` right?

Here's the problem with recording a game: the quiz tape answers question one with `1`, and the recording wants that to be *correct*, then a wrong answer, then the results. If the questions were random, the tape would be right by luck. Night 26's seed is the answer. [`render.sh`](../../docs/demos/render.sh) sets it for every recording:

```bash
# Recordings use a fixed TERMINAL_OF_TERROR_SEED (DEMO_SEED, default 13), so
# re-recording gives the same questions, portraits and fights every time.
```

With seed 13, `quiz -n 3` asks the same three questions in the same order every time, and the tape's keystrokes were chosen to match. The CONTRIBUTING note says the rest: *if you change the seed or the monster data, check that each tape still shows what it should*. The tapes are coupled to the data through the seed, and the coupling is written down.

Nothing in the program was built "for the demos". The seed exists for reproducibility and the `--date` flags for testing; the demos are what those decisions were for, all along.

## A room to film in

`render.sh` builds a small world for the recording to happen in:

```bash
# The recorded shell gets a fake home, so saved paths read ~/.config/...
config="$work/home/.config/terminal-of-terror"
mkdir -p "$work/bin" "$config"
...
printf '%s\n' "PS1='\\[\\e[38;2;255;122;26m\\]🎃 \\[\\e[0m\\]'" \
  "export HOME='$work/home' TERMINAL_OF_TERROR_HOME='$config'" \
  "export TERMINAL_OF_TERROR_SEED='${DEMO_SEED:-13}'" > "$work/rc"
```

A temporary home directory, so the crypt screenshot says `~/.config/terminal-of-terror/progress.json` and not the author's real path. A pumpkin for a prompt. And a `progress.json` with a plausible history, four captures and six badges, written into that home before recording, so the crypt has something to show. A demo of an empty crypt is a demo of nothing. `go build` puts the binary first on `PATH`, so the tape types `terminal-of-terror` like a user would.

## Working around the tools

Two workarounds live in the script, with comments saying why:

```bash
#  - ttyd's browser terminal defaults to Unicode 6 widths, which draws emoji
#    one cell wide and knocks every box border out of line. A small ttyd
#    wrapper turns on Unicode 11.
#  - vhs writes raw frames and we encode them with ffmpeg ourselves, which
#    also lets us keep GIFs small (15fps, diffed palette).
```

VHS records through ttyd, a terminal in a browser, and that browser terminal measured emoji as one column, so every card border was one column off (Night 9's terrifying fact, arriving on cue). The fix is a shell script named `ttyd` earlier on the `PATH` that adds `-t unicodeVersion=11` and hands over to the real one. The script doesn't patch VHS or ttyd; it wraps them, and the wrapper is nine lines.

The encoding: VHS is asked for raw frames (`Output` is a directory), and ffmpeg makes the GIF: overlay the cursor frames, pad with the background colour, drop to 15 frames a second, and build a palette from the *differences* between frames. The hero GIF came out under 2 MB, where VHS's default was several times that. A README with a 10 MB GIF at the top is a README nobody scrolls.

```bash
    ffmpeg -loglevel error -y "${inputs[@]}" -filter_complex \
      "[0][1]overlay,$pad,fps=15,split[a][b];[a]palettegen=max_colors=192:stats_mode=diff[p];[b][p]paletteuse=dither=bayer:bayer_scale=5:diff_mode=rectangle" \
      "$assets/$name.gif"
```

Nobody remembers an ffmpeg filter chain; that's what the script is for.

## The theme lives once

Every tape has `Set Theme @theme.json`, and VHS doesn't support that line. The script replaces it with the contents of [`theme.json`](../../docs/demos/theme.json) as one line of JSON before recording, so the colour scheme is in one file, not nine. When a tool doesn't support the shape you want, a two-line `awk` in the wrapper often does.

## The README

With pictures to hand, [`README.md`](../../README.md) was rewritten as a *programme*: a listing table at the top ("9:00 PM Summon it, 9:05 PM Your first night..."), the hero GIF, then sections in the order a new viewer meets things: install, five commands to try, the explorer, the games, the crypt, the rituals, the roster, packs, and the command reference last. Each game section has its GIF; each ritual its still.

Some habits from that rewrite that apply to any README:

- **Show before tell.** The GIF is above the fold; the feature list is below it.
- **The first commands are copy-pasteable** and each has a comment saying what it does.
- **Alt text describes the picture** for screen readers and for the day the image doesn't load: "Guess the Monster: a portrait slowly clears from the fog as clues appear".
- **Callouts** (`> [!TIP]`, `> [!WARNING]`) for the one thing per section a reader must not miss.
- **Keep the voice.** The README is hosted by Count Cathode too. A project's tone is part of its interface.

## Run it

If you have `vhs`, `ffmpeg` and ttyd 1.7.7 or newer:

```bash
docs/demos/render.sh quiz
```

and look at `docs/assets/quiz.gif`. Without them, read `render.sh` top to bottom; it's 130 lines and every decision has a comment.

## Try it

- Write a tape for `mash dracula mummy --fast`. Add it to `render.sh`'s list (it's a GIF, not a still). Where does its `Sleep` need to be long?
- Change `DEMO_SEED` and re-record the quiz. Watch the tape answer the wrong question.
- Reduce `fps=15` to 8 and compare the file size and the feel.

## 💀 Terrifying fact

The script sets `TMPDIR` to a directory *inside the repository* before running VHS:

```bash
# vhs moves its frames into place with a plain rename, which fails silently
# across filesystems or into a missing directory, so keep its temp files
# next to the output.
```

Night 21 said `os.Rename` across filesystems is a copy, or a failure; VHS's version failed *silently* when `/tmp` was a different filesystem from the checkout, producing no frames and no error. A day was lost to it, and the comment is there so the next person doesn't lose one. When a tool does nothing and says nothing, suspect a rename.

## 🕯️ Before dawn

Look at the README's programme table and find the section you'd read first if you'd never seen the project. Then find the one nobody would read. Is it needed? Tomorrow is the last night, and it ends with a release.

> 📺 *"Lights, camera, and a seed of 13. The pictures are in the can. Tomorrow night, Halloween, we sign off: a version number, a changelog, a tag, and a release for the whole world to download."*

[← Night 29](night-29.md) · [Index](README.md) · [Night 31 →](night-31.md)
