# Changing the website

1. `site/main.go` builds the pages and `site/course.go` renders the Night School markdown from `course/` with
   goldmark. Lesson links of the form `../cmd/list.go` are rewritten to the file on GitHub.
2. Templates are in `site/templates/` (Go html/template, one file per page kind plus `base.html`), styles in
   `site/static/style.css`, behaviour in `site/static/site.js`.
3. Monster pages are generated from the packs and the course from `course/`, so content changes belong there,
   not in the templates.
4. Run `go test ./site` (it builds the site and checks every internal link) and `go run ./site`, then look at
   `site/dist` in a browser at desktop and phone widths.
5. The README's install, games and rituals copy is repeated on the landing page template: update both.

`npm run build` writes the site to `dist/` with a localhost base URL, which is what slopstopper's browser checks
serve on port 8080 (`slopstopper serve`). `.github/workflows/pages.yml` publishes the real site on every push to `main`.
