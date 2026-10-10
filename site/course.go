package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Course is Count Cathode's Night School: the index page and its lessons.
type Course struct {
	Intro   template.HTML // the index, rendered from course/README.md
	Weeks   []Week
	Lessons []Lesson
}

// Week groups lessons the way the course index does.
type Week struct {
	Title   string // e.g. "Week 1 · The Vault"
	Theme   string // e.g. "foundations"
	Lessons []Lesson
}

// Lesson is one night of the course.
type Lesson struct {
	Night   int
	Slug    string // night-01
	Title   string // It's Alive!
	Summary string // what you'll learn, from the index table, as plain text
	Quote   string // the host's opening line
	Body    template.HTML
	Week    string
}

var (
	weekRe   = regexp.MustCompile(`^### (Week \d+ · [^—]+?)\s*—\s*(.+)$`)
	rowRe    = regexp.MustCompile(`^\| \[(\d+)\]\((night-\d+)\.md\) \| (.+?) \| (.+?) \|$`)
	h1Re     = regexp.MustCompile(`(?m)^# Night (\d+) · (.+)$`)
	quoteRe  = regexp.MustCompile(`(?m)^> 📺 \*"(.+?)"\*\s*$`)
	footerRe = regexp.MustCompile(`(?m)^(?:\[[^\]]*\]\((?:night-\d+|README)\.md\)(?: · )?)+\s*$`)
)

// loadCourse reads course/ and renders every page to HTML.
func loadCourse(dir string) (*Course, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		return nil, err
	}
	c := &Course{}
	var week *Week
	byNight := map[int]*Lesson{}
	for _, line := range strings.Split(string(raw), "\n") {
		if m := weekRe.FindStringSubmatch(line); m != nil {
			c.Weeks = append(c.Weeks, Week{Title: strings.TrimSpace(m[1]), Theme: strings.TrimSpace(m[2])})
			week = &c.Weeks[len(c.Weeks)-1]
			continue
		}
		if m := rowRe.FindStringSubmatch(line); m != nil && week != nil {
			var n int
			fmt.Sscanf(m[1], "%d", &n)
			week.Lessons = append(week.Lessons, Lesson{Night: n, Slug: m[2], Title: m[3], Summary: strings.ReplaceAll(m[4], "`", ""), Week: week.Title})
		}
	}
	for i := range c.Weeks {
		for j := range c.Weeks[i].Lessons {
			l := &c.Weeks[i].Lessons[j]
			if err := l.load(dir); err != nil {
				return nil, err
			}
			byNight[l.Night] = l
		}
	}
	for n := 1; n <= len(byNight); n++ {
		l, ok := byNight[n]
		if !ok {
			return nil, fmt.Errorf("the course index has no night %d", n)
		}
		c.Lessons = append(c.Lessons, *l)
	}
	intro, err := renderMarkdown(raw)
	if err != nil {
		return nil, err
	}
	c.Intro = intro
	return c, nil
}

// load reads and renders the lesson's markdown file.
func (l *Lesson) load(dir string) error {
	raw, err := os.ReadFile(filepath.Join(dir, l.Slug+".md"))
	if err != nil {
		return err
	}
	if m := h1Re.FindSubmatch(raw); m != nil {
		l.Title = string(m[2])
	}
	if m := quoteRe.FindSubmatch(raw); m != nil {
		l.Quote = string(m[1])
	}
	// The H1 and the Index/Night links at the foot are drawn by the page
	// itself, so they come out of the body.
	body := h1Re.ReplaceAll(raw, nil)
	body = footerRe.ReplaceAll(body, nil)
	html, err := renderMarkdown(body)
	if err != nil {
		return fmt.Errorf("%s: %w", l.Slug, err)
	}
	l.Body = html
	return nil
}

var markdown = goldmark.New(
	goldmark.WithExtensions(extension.GFM, extension.Typographer),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
		parser.WithASTTransformers(util.Prioritized(linkRewriter{}, 100)),
	),
)

func renderMarkdown(src []byte) (template.HTML, error) {
	var buf bytes.Buffer
	if err := markdown.Convert(src, &buf); err != nil {
		return "", err
	}
	// Code blocks scroll sideways on a phone, so keyboard users need to be
	// able to focus them to scroll.
	out := strings.ReplaceAll(buf.String(), "<pre><code", `<pre tabindex="0"><code`)
	return template.HTML(out), nil
}

// linkRewriter points the lessons' links at the website instead of the
// markdown files: night-02.md becomes night-02.html, README.md becomes
// index.html, and ../cmd/list.go goes to the file on GitHub.
type linkRewriter struct{}

func (linkRewriter) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Link:
			v.Destination = []byte(rewriteLink(string(v.Destination)))
		case *ast.Image:
			v.Destination = []byte(rewriteLink(string(v.Destination)))
		}
		return ast.WalkContinue, nil
	})
}

func rewriteLink(dest string) string {
	target, frag, _ := strings.Cut(dest, "#")
	if frag != "" {
		frag = "#" + frag
	}
	switch {
	case strings.Contains(target, "://"), strings.HasPrefix(dest, "#"), strings.HasPrefix(target, "mailto:"):
		return dest
	case target == "README.md":
		return "index.html" + frag
	case strings.HasPrefix(target, "../"):
		return repoURL + "/blob/main/" + strings.TrimPrefix(target, "../") + frag
	case strings.HasSuffix(target, ".md"):
		return strings.TrimSuffix(target, ".md") + ".html" + frag
	}
	return dest
}
