#!/usr/bin/env bash
# Renders the README's demo GIFs and screenshots from docs/demos/*.tape.
#
#   docs/demos/render.sh            # everything
#   docs/demos/render.sh hero quiz  # just these
#
# Needs: go, vhs (https://github.com/charmbracelet/vhs), ffmpeg, and ttyd 1.7.7
# or newer. Set TTYD=/path/to/ttyd to pick a specific ttyd binary.
#
# Two workarounds live here rather than in the tapes:
#  - ttyd's browser terminal defaults to Unicode 6 widths, which draws emoji
#    one cell wide and knocks every box border out of line. A small ttyd
#    wrapper turns on Unicode 11.
#  - vhs writes raw frames and we encode them with ffmpeg ourselves, which
#    also lets us keep GIFs small (15fps, diffed palette).
set -euo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
cd "$root"
demos=docs/demos
assets=docs/assets
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# Stills keep only their final frame, as a PNG.
stills=" tonight countdown random list crypt "
bg="#0d0a12"

for tool in go vhs ffmpeg; do
  command -v "$tool" > /dev/null || { echo "missing $tool" >&2; exit 1; }
done
real_ttyd=${TTYD:-$(command -v ttyd || true)}
[ -n "$real_ttyd" ] || { echo "missing ttyd (1.7.7+)" >&2; exit 1; }

# The recorded shell gets a fake home, so saved paths read ~/.config/...
config="$work/home/.config/terminal-of-terror"
mkdir -p "$work/bin" "$config"
cat > "$work/bin/ttyd" <<EOF
#!/bin/sh
exec "$real_ttyd" -t unicodeVersion=11 "\$@"
EOF
chmod +x "$work/bin/ttyd"

# ttyd starts "bash" from PATH; this one swaps vhs's plain prompt for a pumpkin.
real_bash=$(command -v bash)
printf '%s\n' "PS1='\\[\\e[38;2;255;122;26m\\]🎃 \\[\\e[0m\\]'" \
  "export HOME='$work/home' TERMINAL_OF_TERROR_HOME='$config'" > "$work/rc"
printf '#!%s\nexec "%s" --noprofile --rcfile "%s" -i +o history\n' \
  "$real_bash" "$real_bash" "$work/rc" > "$work/bin/bash"
chmod +x "$work/bin/bash"

go build -o "$work/bin/terminal-of-terror" .

# A crypt with some history, so the screenshot has something to show.
cat > "$config/progress.json" <<'EOF'
{
  "version": 1,
  "knowledge": {"dracula": 3, "frankenstein": 3, "mummy": 3, "phantom": 3, "wolf-man": 2, "creature": 1, "invisible-man": 2},
  "captured": {
    "dracula": "2026-10-01T21:13:00Z",
    "frankenstein": "2026-10-03T22:40:00Z",
    "mummy": "2026-10-07T23:05:00Z",
    "phantom": "2026-10-13T00:31:00Z"
  },
  "badges": {
    "first-fright": "2026-10-01T21:13:00Z",
    "grave-robber": "2026-10-01T21:13:00Z",
    "eagle-eye": "2026-10-04T20:02:00Z",
    "promoter": "2026-10-05T19:44:00Z",
    "night-owl": "2026-10-13T00:31:00Z",
    "friday-13": "2026-11-13T21:00:00Z"
  },
  "seen": {"dracula": true, "frankenstein": true, "mummy": true, "phantom": true, "wolf-man": true},
  "quizzesPlayed": 9,
  "bestQuizScore": 90,
  "guessesPlayed": 4,
  "bestGuessScore": 420,
  "mashesPlayed": 6
}
EOF

# vhs moves its frames into place with a plain rename, which fails silently
# across filesystems or into a missing directory, so keep its temp files
# next to the output.
mkdir -p "$demos/.frames/tmp"
export TMPDIR="$root/$demos/.frames/tmp"

export PATH="$work/bin:$PATH"
export VHS_NO_SANDBOX=${VHS_NO_SANDBOX:-true}

names=("$@")
if [ ${#names[@]} -eq 0 ]; then
  for tape in "$demos"/*.tape; do names+=("$(basename "$tape" .tape)"); done
fi

for name in "${names[@]}"; do
  echo "🎬 $name"
  frames="$demos/.frames/$name"
  rm -rf "$frames"
  vhs "$demos/$name.tape" > "$work/$name.log" 2>&1 || { cat "$work/$name.log"; exit 1; }
  [ -e "$frames/frame-text-00001.png" ] || { cat "$work/$name.log"; echo "no frames for $name" >&2; exit 1; }

  inputs=(-framerate 50 -start_number 1 -i "$frames/frame-text-%05d.png"
          -framerate 50 -start_number 1 -i "$frames/frame-cursor-%05d.png")
  pad="pad=iw+48:ih+48:24:24:color=$bg"
  if [[ "$stills" == *" $name "* ]]; then
    last=$(find "$frames" -name 'frame-text-*.png' | sort | tail -1)
    ffmpeg -loglevel error -y -i "$last" -vf "$pad" "$assets/$name.png"
    echo "   → $assets/$name.png"
  else
    ffmpeg -loglevel error -y "${inputs[@]}" -filter_complex \
      "[0][1]overlay,$pad,fps=15,split[a][b];[a]palettegen=max_colors=192:stats_mode=diff[p];[b][p]paletteuse=dither=bayer:bayer_scale=5:diff_mode=rectangle" \
      "$assets/$name.gif"
    echo "   → $assets/$name.gif ($(du -h "$assets/$name.gif" | cut -f1))"
  fi
  rm -rf "$frames"
done
rm -rf "$demos/.frames"
