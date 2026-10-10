# Changing the UI

1. `internal/ui/ui.go` holds the explorer model and key handling, `render.go` the views, `intro.go` the
   Channel 13 opening and `cards.go` the non-interactive output.
2. Update the model struct if needed, `Update()` for new interactions and the render functions for display changes.
3. Run `go test ./internal/ui` (it renders every page at several terminal sizes), and
   `DUMP=1 go test ./internal/ui -run Dump -v` to see sample screens.
4. Test interactivity in a real terminal, at small and large widths and heights.
5. If the change is visible in the README demos, re-record them with `docs/demos/render.sh` (see "Recording the
   Demos" in CONTRIBUTING.md) and check the results.

Styling comes from `internal/ui/styles.go`; per-monster colours come from each monster's `theme`. Avoid emoji
that need a variation selector (⚰️, 🎞️), because they break box alignment.
