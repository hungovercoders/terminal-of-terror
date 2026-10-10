// Command site builds the Terminal of Terror website: a static site with a
// landing page, one page per monster (from the built-in packs) and Count
// Cathode's Night School (from docs/course). Run it from the repository
// root:
//
//	go run ./site                 # writes site/dist
//	go run ./site -out /tmp/www   # somewhere else
//
// .github/workflows/pages.yml runs it on every push to main and publishes
// the result to GitHub Pages.
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hungovercoders/terminal-of-terror/internal/crypt"
	"github.com/hungovercoders/terminal-of-terror/internal/host"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

//go:embed templates static
var siteFS embed.FS

const (
	repoURL     = "https://github.com/hungovercoders/terminal-of-terror"
	defaultBase = "https://hungovercoders.github.io/terminal-of-terror/"
)

func main() {
	out := flag.String("out", "site/dist", "directory to write the site into")
	base := flag.String("base", defaultBase, "public URL of the site, for the sitemap and social cards")
	course := flag.String("course", "docs/course", "directory holding the Night School lessons")
	assets := flag.String("assets", "docs/assets", "directory holding the README GIFs and screenshots")
	flag.Parse()

	n, err := Build(Config{Out: *out, BaseURL: *base, CourseDir: *course, AssetsDir: *assets})
	if err != nil {
		fmt.Fprintln(os.Stderr, "site:", err)
		os.Exit(1)
	}
	fmt.Printf("📺 Channel 13 is on the air: %d pages written to %s\n", n, *out)
}

// Config says where the site's inputs are and where to write it.
type Config struct {
	Out       string
	BaseURL   string
	CourseDir string
	AssetsDir string
}

// Site is what every template can see.
type Site struct {
	Name         string
	Tagline      string
	Repo         string
	BaseURL      string
	MonsterCount int
	PackCount    int
	BadgeCount   int
	LessonCount  int
	Host         string
	Channel      string
}

// Page is the data handed to a template: the site, the page's own fields
// and Root, the path back to the site's top level ("" or "../"). The 404
// page is the exception: GitHub Pages serves it at whatever address was
// missing, so its Root is the site's absolute path instead.
type Page struct {
	Site        Site
	Root        string
	Path        string // e.g. "monsters/dracula.html"
	Title       string
	Description string
	Section     string // which nav item is lit: home, monsters, course
	Theme       *Theme // per-monster colours, when the page has them
	Data        any
}

// URL is the page's absolute address.
func (p Page) URL() string { return p.Site.BaseURL + p.Path }

// Theme is a page's colour scheme.
type Theme struct {
	Primary string
	Accent  string
	Silent  bool
}

// Build writes the whole site and returns how many HTML pages it made.
func Build(cfg Config) (int, error) {
	if err := os.MkdirAll(cfg.Out, 0o755); err != nil {
		return 0, err
	}
	if !strings.HasSuffix(cfg.BaseURL, "/") {
		cfg.BaseURL += "/"
	}
	packs := monsters.AllPacks()
	all := monsters.EveryMonster()

	course, err := loadCourse(cfg.CourseDir)
	if err != nil {
		return 0, fmt.Errorf("course: %w", err)
	}

	site := Site{
		Name:         "Terminal of Terror",
		Tagline:      "Channel 13's Creature Feature, live in your terminal",
		Repo:         repoURL,
		BaseURL:      cfg.BaseURL,
		MonsterCount: len(all),
		PackCount:    len(packs),
		BadgeCount:   len(crypt.Badges),
		LessonCount:  len(course.Lessons),
		Host:         host.Name,
		Channel:      host.Channel,
	}

	base, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return 0, fmt.Errorf("base URL: %w", err)
	}
	b := &builder{cfg: cfg, site: site, basePath: base.Path}
	if err := b.copyStatic(); err != nil {
		return 0, err
	}
	if err := b.copyAssets(); err != nil {
		return 0, err
	}

	b.render("index.html", Page{
		Title:       site.Name,
		Description: "Meet the classic Universal monsters, world folklore, cryptids and literary monsters in your terminal: an interactive explorer, the Midnight Quiz, Guess the Monster, the Monster Mash and nightly rituals, all hosted by Count Cathode.",
		Section:     "home",
		Data:        homeData(packs, all, course),
	})
	b.render("monsters/index.html", Page{
		Title:       "The Roster",
		Description: fmt.Sprintf("Every one of the %d monsters in the vault, from Dracula to Krampus, with the real history behind each.", len(all)),
		Section:     "monsters",
		Data:        rosterData(packs),
	})
	for i, m := range all {
		d := monsterData(packs, all, i)
		b.render("monsters/"+m.ID+".html", Page{
			Title:       m.Name,
			Description: m.Description + ". " + excerpt(m.Legend, 140),
			Section:     "monsters",
			Theme:       d.Theme,
			Data:        d,
		})
	}
	b.render("night-school/index.html", Page{
		Title:       "Count Cathode's Night School",
		Description: "Build Terminal of Terror from nothing in 31 nights: a course in writing a real Go command-line application.",
		Section:     "course",
		Data:        course,
	})
	for i := range course.Lessons {
		l := course.Lessons[i]
		b.render("night-school/"+l.Slug+".html", Page{
			Title:       l.Title,
			Description: l.Summary,
			Section:     "course",
			Data:        lessonData{Course: course, Lesson: l, Index: i},
		})
	}
	b.render("404.html", Page{Title: "Technical difficulties", Section: "home"})

	if b.err != nil {
		return 0, b.err
	}
	if err := b.writeExtras(); err != nil {
		return 0, err
	}
	return len(b.pages), nil
}

type builder struct {
	cfg      Config
	site     Site
	basePath string // the path part of BaseURL, e.g. "/terminal-of-terror/"
	tmpls    map[string]*template.Template
	pages    []string
	err      error
}

// template parses base.html plus templates/<name> once and caches it.
func (b *builder) template(name string) (*template.Template, error) {
	if t, ok := b.tmpls[name]; ok {
		return t, nil
	}
	t, err := template.New("").Funcs(funcs).ParseFS(siteFS, "templates/base.html", "templates/"+name)
	if err != nil {
		return nil, err
	}
	if b.tmpls == nil {
		b.tmpls = map[string]*template.Template{}
	}
	b.tmpls[name] = t
	return t, nil
}

// render executes templates/<kind>.html inside the base layout and writes
// the result to <out>/<rel>. The template is chosen from the page's path:
// monsters/<id>.html uses monster.html, night-school/night-NN.html uses
// lesson.html, and each index uses its own file.
func (b *builder) render(rel string, p Page) {
	if b.err != nil {
		return
	}
	p.Site = b.site
	p.Path = rel
	p.Root = strings.Repeat("../", strings.Count(rel, "/"))
	if rel == "404.html" {
		p.Root = b.basePath
	}

	tmpl, err := b.template(templateFor(rel))
	if err != nil {
		b.err = err
		return
	}
	full := filepath.Join(b.cfg.Out, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		b.err = err
		return
	}
	f, err := os.Create(full)
	if err != nil {
		b.err = err
		return
	}
	defer f.Close()
	if err := tmpl.ExecuteTemplate(f, "base", p); err != nil {
		b.err = fmt.Errorf("%s: %w", rel, err)
		return
	}
	b.pages = append(b.pages, rel)
}

func templateFor(rel string) string {
	switch {
	case rel == "index.html":
		return "home.html"
	case rel == "404.html":
		return "404.html"
	case rel == "monsters/index.html":
		return "roster.html"
	case strings.HasPrefix(rel, "monsters/"):
		return "monster.html"
	case rel == "night-school/index.html":
		return "course.html"
	default:
		return "lesson.html"
	}
}

// copyStatic writes the embedded CSS and JS to <out>/static.
func (b *builder) copyStatic() error {
	return fs.WalkDir(siteFS, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := siteFS.ReadFile(p)
		if err != nil {
			return err
		}
		return writeFile(filepath.Join(b.cfg.Out, filepath.FromSlash(p)), data)
	})
}

// copyAssets brings the README's GIFs and screenshots along to <out>/assets.
func (b *builder) copyAssets() error {
	entries, err := os.ReadDir(b.cfg.AssetsDir)
	if err != nil {
		return fmt.Errorf("assets: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(b.cfg.AssetsDir, e.Name()))
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(b.cfg.Out, "assets", e.Name()), data); err != nil {
			return err
		}
	}
	return nil
}

// writeExtras adds the files a static host wants beside the pages.
func (b *builder) writeExtras() error {
	var sm strings.Builder
	sm.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sm.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, p := range b.pages {
		if p == "404.html" {
			continue
		}
		loc := b.site.BaseURL + strings.TrimSuffix(p, "index.html")
		fmt.Fprintf(&sm, "  <url><loc>%s</loc></url>\n", template.HTMLEscapeString(loc))
	}
	sm.WriteString("</urlset>\n")
	files := map[string]string{
		"sitemap.xml": sm.String(),
		"robots.txt":  "User-agent: *\nAllow: /\nSitemap: " + b.site.BaseURL + "sitemap.xml\n",
		".nojekyll":   "",
	}
	for name, body := range files {
		if err := writeFile(filepath.Join(b.cfg.Out, name), []byte(body)); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(name string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return err
	}
	return os.WriteFile(name, data, 0o644)
}

// ---- page data ----

// homeData feeds the landing page.
type homePage struct {
	Packs    []monsters.Pack
	Monsters []monsters.Monster
	Badges   []crypt.Badge
	Greeting string
	SignOff  string
	Facts    []factEntry
	Course   *Course
	Commands []command
	Sections []explorerSection
}

type factEntry struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	Fact  string `json:"fact"`
}

type command struct{ Use, Does string }

type explorerSection struct{ Name, Learn string }

func homeData(packs []monsters.Pack, all []monsters.Monster, course *Course) homePage {
	var facts []factEntry
	for _, m := range all {
		for _, f := range m.Facts {
			facts = append(facts, factEntry{ID: m.ID, Name: m.Name, Emoji: m.Emoji, Fact: f})
		}
	}
	return homePage{
		Packs:    packs,
		Monsters: all,
		Badges:   crypt.Badges,
		// A fixed seed, so rebuilding the site without changes is a no-op.
		Greeting: host.Greeting(rand.New(rand.NewSource(13))),
		SignOff:  "That's all for tonight. Sleep tight, and check under the bed.",
		Facts:    facts,
		Course:   course,
		Commands: []command{
			{"monster [name]", "The interactive explorer (--all, --no-intro, --json)"},
			{"list", "Every monster, grouped by pack (--json)"},
			{"random", "A random fact (--daily, --date YYYY-MM-DD, --json)"},
			{"quiz [monster]", "The Midnight Quiz (-n, --json)"},
			{"guess", "Guess the Monster (--rounds)"},
			{"mash [monster] [monster]", "The Monster Mash (--fast)"},
			{"crypt", "Your captured monsters, badges and records (--json); crypt reset starts again"},
			{"tonight", "Tonight's double feature (--shuffle, --date)"},
			{"countdown", "Nights until Halloween (--date)"},
			{"packs", "List packs (--json); packs new <id> makes one"},
		},
		Sections: []explorerSection{
			{"Facts", "Terrifying facts, origin and first appearance"},
			{"Legend", "The real folklore and history behind the monster"},
			{"The Film", "The classic film's director, star, makeup artist and release date"},
			{"Myth vs Movie", "True or false? Make your guesses, then press r for the verdicts"},
			{"Quotes", "Lines from the original public-domain novels"},
			{"Stat Card", "Strength, speed, cunning and dread, plus powers and weaknesses"},
			{"Legacy", "Sequels, remakes and crossovers"},
		},
	}
}

// rosterData feeds the roster page: every pack with its monsters.
type rosterPage struct {
	Packs []monsters.Pack
}

func rosterData(packs []monsters.Pack) rosterPage { return rosterPage{Packs: packs} }

// monsterPage feeds one monster's page.
type monsterPage struct {
	Monster monsters.Monster
	Pack    monsters.Pack
	Theme   *Theme
	Prev    *monsters.Monster
	Next    *monsters.Monster
	Number  int
	Total   int
	Stats   []stat
	Art     string
	Silent  bool
}

type stat struct {
	Name  string
	Value int
}

func monsterData(packs []monsters.Pack, all []monsters.Monster, i int) monsterPage {
	m := all[i]
	d := monsterPage{
		Monster: m,
		Number:  i + 1,
		Total:   len(all),
		Silent:  m.IsSilent(),
		Art:     m.ASCII,
		Stats: []stat{
			{"Strength", m.Stats.Strength},
			{"Speed", m.Stats.Speed},
			{"Cunning", m.Stats.Cunning},
			{"Dread", m.Stats.Dread},
		},
	}
	for _, p := range packs {
		if p.ID == m.Pack {
			d.Pack = p
		}
	}
	if i > 0 {
		d.Prev = &all[i-1]
	}
	if i+1 < len(all) {
		d.Next = &all[i+1]
	}
	t := &Theme{Primary: m.Theme.Primary, Accent: m.Theme.Accent, Silent: d.Silent}
	if d.Silent {
		t.Primary, t.Accent = "#F5F5F5", "#BDBDBD"
	}
	if t.Primary == "" {
		t.Primary = "#FFD700"
	}
	if t.Accent == "" {
		t.Accent = "#FF6347"
	}
	d.Theme = t
	return d
}

type lessonData struct {
	Course *Course
	Lesson Lesson
	Index  int
}

// Prev and Next give the neighbouring lessons, if any.
func (d lessonData) Prev() *Lesson {
	if d.Index == 0 {
		return nil
	}
	return &d.Course.Lessons[d.Index-1]
}

func (d lessonData) Next() *Lesson {
	if d.Index+1 >= len(d.Course.Lessons) {
		return nil
	}
	return &d.Course.Lessons[d.Index+1]
}

// ---- template helpers ----

var funcs = template.FuncMap{
	"pct": func(v int) int { return v * 10 },
	"json": func(v any) (template.JS, error) {
		b, err := json.Marshal(v)
		return template.JS(b), err
	},
	"firstAppearance": func(m monsters.Monster) string { return m.FirstAppearance() },
}

// excerpt shortens a paragraph to about n characters on a word boundary,
// for meta descriptions. Splitting on ". " would cut "St. Nicholas" in half.
func excerpt(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := strings.LastIndex(s[:n], " ")
	if cut <= 0 {
		cut = n
	}
	return strings.TrimRight(s[:cut], ",;:") + "…"
}
