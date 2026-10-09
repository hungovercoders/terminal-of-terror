package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

// TestBuild generates the whole site and checks that every page is there
// and that no internal link points at a file that wasn't written.
func TestBuild(t *testing.T) {
	out := t.TempDir()
	n, err := Build(Config{Out: out, BaseURL: "https://example.test/tot/", CourseDir: "../docs/course", AssetsDir: "../docs/assets"})
	if err != nil {
		t.Fatal(err)
	}
	all := monsters.EveryMonster()
	want := 1 + 1 + len(all) + 1 + 31 + 1 // home, roster, monsters, course index, lessons, 404
	if n != want {
		t.Errorf("wrote %d pages, want %d", n, want)
	}
	for _, f := range []string{"index.html", "monsters/index.html", "monsters/dracula.html", "night-school/index.html", "night-school/night-31.html", "404.html", "static/style.css", "static/site.js", "assets/hero.gif", "sitemap.xml", "robots.txt", ".nojekyll"} {
		if _, err := os.Stat(filepath.Join(out, f)); err != nil {
			t.Errorf("missing %s", f)
		}
	}

	var pages []string
	err = filepath.WalkDir(out, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".html") {
			pages = append(pages, p)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pages {
		checkLinks(t, out, "/tot/", p)
	}

	// GitHub Pages serves 404.html at the missing address, so its links
	// must not be relative.
	notFound, err := os.ReadFile(filepath.Join(out, "404.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`href="/tot/static/style.css"`, `href="/tot/index.html"`, `src="/tot/static/site.js"`} {
		if !strings.Contains(string(notFound), want) {
			t.Errorf("404.html lacks %s", want)
		}
	}
	if strings.Contains(string(notFound), `href="static/`) || strings.Contains(string(notFound), `href="../`) {
		t.Error("404.html still has relative links")
	}
}

var (
	hrefRe        = regexp.MustCompile(`(?:href|src)="([^"]+)"`)
	relMarkdownRe = regexp.MustCompile(`href="(?:[^":]*/)?[^":/]*\.md(?:#[^"]*)?"`)
)

// checkLinks follows every href and src on the page that points into the
// site, relative or absolute under basePath.
func checkLinks(t *testing.T, root, basePath, page string) {
	t.Helper()
	body, err := os.ReadFile(page)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range hrefRe.FindAllStringSubmatch(string(body), -1) {
		link := m[1]
		if strings.Contains(link, "://") || strings.HasPrefix(link, "#") || strings.HasPrefix(link, "mailto:") || strings.HasPrefix(link, "data:") {
			continue
		}
		link, _, _ = strings.Cut(link, "#")
		link, _, _ = strings.Cut(link, "?")
		if link == "" {
			continue
		}
		target := filepath.Join(filepath.Dir(page), filepath.FromSlash(link))
		if strings.HasPrefix(link, "/") {
			if !strings.HasPrefix(link, basePath) {
				rel, _ := filepath.Rel(root, page)
				t.Errorf("%s links to %q, which is outside the site", rel, m[1])
				continue
			}
			target = filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(link, basePath)))
		}
		if strings.HasSuffix(link, "/") {
			target = filepath.Join(target, "index.html")
		}
		if _, err := os.Stat(target); err != nil {
			rel, _ := filepath.Rel(root, page)
			t.Errorf("%s links to %q, which doesn't exist", rel, m[1])
		}
	}
}

func TestMonsterPageContent(t *testing.T) {
	out := t.TempDir()
	if _, err := Build(Config{Out: out, BaseURL: "https://example.test/", CourseDir: "../docs/course", AssetsDir: "../docs/assets"}); err != nil {
		t.Fatal(err)
	}
	for _, m := range monsters.EveryMonster() {
		body, err := os.ReadFile(filepath.Join(out, "monsters", m.ID+".html"))
		if err != nil {
			t.Fatal(err)
		}
		page := string(body)
		for _, want := range []string{m.Name, m.Facts[0], string([]rune(m.Legend)[:20]), "terminal-of-terror monster " + m.ID} {
			if !strings.Contains(page, htmlEscape(want)) {
				t.Errorf("%s: page lacks %q", m.ID, want)
			}
		}
		if m.IsSilent() != strings.Contains(page, `class="monster silent"`) {
			t.Errorf("%s: silent styling is wrong", m.ID)
		}
	}
}

func TestRewriteLink(t *testing.T) {
	cases := map[string]string{
		"night-02.md":            "night-02.html",
		"README.md":              "index.html",
		"README.md#before-night": "index.html#before-night",
		"../../cmd/list.go":      repoURL + "/blob/main/cmd/list.go",
		"https://go.dev/dl/":     "https://go.dev/dl/",
		"#try-it":                "#try-it",
	}
	for in, want := range cases {
		if got := rewriteLink(in); got != want {
			t.Errorf("rewriteLink(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLessonLinksAreRewritten(t *testing.T) {
	c, err := loadCourse("../docs/course")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Lessons) != 31 {
		t.Fatalf("got %d lessons, want 31", len(c.Lessons))
	}
	if len(c.Weeks) != 4 {
		t.Errorf("got %d weeks, want 4", len(c.Weeks))
	}
	for _, l := range c.Lessons {
		for _, m := range relMarkdownRe.FindAllString(string(l.Body), -1) {
			t.Errorf("%s still links to a markdown file: %s", l.Slug, m)
		}
		if l.Title == "" || l.Summary == "" || l.Quote == "" {
			t.Errorf("%s: title %q, summary %q, quote %q", l.Slug, l.Title, l.Summary, l.Quote)
		}
	}
	if c.Lessons[0].Title != "It's Alive!" {
		t.Errorf("night 1 is %q", c.Lessons[0].Title)
	}
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&#34;", "'", "&#39;")
	return r.Replace(s)
}
