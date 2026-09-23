package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hungovercoders/terminal-of-terror/internal/host"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

type tab int

const (
	tabFacts tab = iota
	tabLegend
	tabFilm
	tabMyths
	tabQuotes
	tabStats
	tabLegacy
)

var tabNames = map[tab]string{
	tabFacts:  "Facts",
	tabLegend: "Legend",
	tabFilm:   "The Film",
	tabMyths:  "Myth vs Movie",
	tabQuotes: "Quotes",
	tabStats:  "Stat Card",
	tabLegacy: "Legacy",
}

// shortTabNames keep the tab bar on one line in narrow terminals.
var shortTabNames = map[tab]string{
	tabFilm:  "Film",
	tabMyths: "Myths",
	tabStats: "Stats",
}

// tabsFor lists the sections a monster has something to say in.
func tabsFor(m monsters.Monster) []tab {
	tabs := []tab{tabFacts, tabLegend}
	if m.Film != nil {
		tabs = append(tabs, tabFilm)
	}
	if len(m.Myths) > 0 {
		tabs = append(tabs, tabMyths)
	}
	if len(m.Quotes) > 0 {
		tabs = append(tabs, tabQuotes)
	}
	tabs = append(tabs, tabStats)
	if len(m.Legacy) > 0 {
		tabs = append(tabs, tabLegacy)
	}
	return tabs
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	switch m.screen {
	case screenIntro:
		return m.viewIntro()
	case screenGallery:
		return m.viewGallery()
	default:
		return m.viewDetail()
	}
}

// ---- shared chrome ----

func (m model) titleBar(right string) string {
	w := m.contentWidth()
	left := titleStyle.Render("🎃 TERMINAL OF TERROR 🎃")
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 2 {
		return left
	}
	return left + strings.Repeat(" ", gap) + helpStyle.Render(right)
}

var (
	galleryKeys = []string{"↑/↓ choose", "enter open", "/ search", "? help", "q quit"}
	detailKeys  = []string{"←/→ monster", "tab section", "↑/↓ scroll", "/ search", "g gallery", "? help", "q quit"}
)

func (m model) footer(keys []string) string {
	w := m.contentWidth()
	if m.searching {
		prompt := headingStyle.Render("/") + " " + m.query + "▌"
		return prompt + "\n" + helpStyle.Render(joinFit([]string{"type to search", "↑/↓ choose", "enter open", "esc cancel"}, " · ", w))
	}
	if m.showHelp {
		return helpStyle.Render(joinFit(fullHelp, "   ", w))
	}
	return helpStyle.Render(joinFit(keys, " · ", w))
}

var fullHelp = []string{
	"Monsters: ←/→ h/l p/n", "Sections: tab/shift+tab or 1-7", "Scroll: ↑/↓ j/k space pgup/pgdn home/end",
	"Myth vs Movie: r reveals verdicts", "Gallery: g or esc", "Search: /", "Help: ?", "Quit: q or ctrl+c",
}

// viewport shows the slice of lines starting at offset that fits height.
func viewport(lines []string, offset, height int) string {
	if height <= 0 || len(lines) <= height {
		return strings.Join(lines, "\n")
	}
	offset = min(max(offset, 0), len(lines)-height)
	visible := append([]string{}, lines[offset:offset+height]...)
	if offset > 0 {
		visible[0] = helpStyle.Render("  ▲ more above")
	}
	if offset+height < len(lines) {
		visible[len(visible)-1] = helpStyle.Render("  ▼ more below")
	}
	return strings.Join(visible, "\n")
}

// ---- gallery ----

func (m model) viewGallery() string {
	w := m.contentWidth()
	header := m.titleBar(fmt.Sprintf("%d monsters in the vault", len(m.monsters)))
	footer := m.footer(galleryKeys)

	var body string
	if m.searching {
		body = m.searchResults(w)
	} else {
		listW := w
		showPreview := w >= 100
		if showPreview {
			listW = w / 2
		}
		lines, cursorLine := m.galleryLines(listW)
		h := m.availableHeight(header, footer)
		offset := 0
		if h > 0 && cursorLine >= h-1 {
			offset = cursorLine - h + 3
		}
		list := viewport(lines, offset, h)
		if showPreview {
			list = lipgloss.NewStyle().Width(listW).Render(list)
			body = lipgloss.JoinHorizontal(lipgloss.Top, list, m.preview(m.monsters[m.cursor], w-listW-2))
		} else {
			body = list
		}
	}
	return header + "\n\n" + body + "\n\n" + footer
}

func (m model) galleryLines(width int) ([]string, int) {
	var lines []string
	cursorLine := 0
	pack := ""
	for i, mo := range m.monsters {
		if mo.Pack != pack {
			pack = mo.Pack
			if len(lines) > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, headingStyle.Render("── "+packName(pack)+" ──"))
		}
		p := paletteFor(mo)
		name := mo.Emoji + " " + mo.Name
		marker := "  "
		nameStyle := lipgloss.NewStyle().Foreground(p.primary)
		if i == m.cursor {
			marker = lipgloss.NewStyle().Foreground(colorBlood).Render("▶ ")
			nameStyle = nameStyle.Bold(true).Underline(true)
			cursorLine = len(lines)
		}
		year := ""
		if y := debutYear(mo); y != "" {
			year = helpStyle.Render(" (" + y + ")")
		}
		lines = append(lines, marker+nameStyle.Render(name)+year)
		lines = append(lines, "    "+hostStyle.Render(truncate(mo.Description, width-6)))
	}
	return lines, cursorLine
}

func (m model) preview(mo monsters.Monster, width int) string {
	p := paletteFor(mo)
	art := lipgloss.NewStyle().Foreground(p.primary).Render(mo.ASCII)
	text := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(p.primary).Render(mo.Emoji+" "+mo.Name),
		metaStyle.Render(wrap("First appearance: "+mo.FirstAppearance(), width-4)),
		"",
		art,
	)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.accent).
		Padding(0, 1).
		MaxWidth(width).
		Render(text)
}

// ---- detail ----

func (m model) viewDetail() string {
	header := m.detailHeader()
	footer := m.footer(detailKeys)
	var body string
	if m.searching {
		body = m.searchResults(m.contentWidth())
	} else {
		body = viewport(m.bodyLines(), m.scroll, m.availableHeight(header, footer))
	}
	return header + "\n" + body + "\n\n" + footer
}

func (m model) detailHeader() string {
	mo := m.current()
	p := paletteFor(mo)
	w := m.contentWidth()

	// Squeeze the header on short terminals so the content still fits.
	compact := m.height > 0 && m.height < 32
	tiny := m.height > 0 && m.height < 22

	right := fmt.Sprintf("Monster %d of %d · %s", m.index+1, len(m.monsters), packName(mo.Pack))
	lines := []string{m.titleBar(right)}
	if !compact {
		lines = append(lines, "")
	}

	name := lipgloss.NewStyle().Bold(true).Foreground(p.primary).Render(mo.Emoji + "  " + strings.ToUpper(mo.Name))
	if p.silent {
		name += "  " + lipgloss.NewStyle().Foreground(colorBlack).Background(p.accent).Render(fmt.Sprintf(" ▶ SILENT PICTURE · %d ", mo.Film.Year))
	}
	lines = append(lines, name)
	switch {
	case p.silent && !tiny:
		lines = append(lines, intertitle(mo.Description, min(w, 72), p, !compact))
	case tiny:
		lines = append(lines, lipgloss.NewStyle().Italic(true).Foreground(p.accent).Render(truncate(mo.Description, w)))
	default:
		lines = append(lines, lipgloss.NewStyle().Italic(true).Foreground(p.accent).Render(wrap(mo.Description, w)))
	}
	intro := "📺 " + host.Name + ": “" + host.Intro(mo) + "”"
	switch {
	case tiny:
	case compact:
		lines = append(lines, hostStyle.Render(truncate(intro, w)))
	default:
		lines = append(lines, hostStyle.Render(wrap(intro, w)))
	}
	if !compact {
		lines = append(lines, "")
	}
	lines = append(lines, m.tabBar(p))
	return strings.Join(lines, "\n")
}

func (m model) tabBar(p palette) string {
	active := lipgloss.NewStyle().Bold(true).Foreground(colorBlack).Background(p.primary).Padding(0, 1)
	inactive := lipgloss.NewStyle().Foreground(colorDim).Padding(0, 1)
	var parts []string
	for i, t := range tabsFor(m.current()) {
		name := tabNames[t]
		if short, ok := shortTabNames[t]; ok && m.contentWidth() < 100 {
			name = short
		}
		label := fmt.Sprintf("%d %s", i+1, name)
		if i == m.tab {
			parts = append(parts, active.Render(label))
		} else {
			parts = append(parts, inactive.Render(label))
		}
	}
	return joinFit(parts, " ", m.contentWidth())
}

// bodyLines renders the scrollable part of the detail page.
func (m model) bodyLines() []string {
	mo := m.current()
	p := paletteFor(mo)
	w := m.contentWidth()
	tabs := tabsFor(mo)
	t := tabs[min(m.tab, len(tabs)-1)]

	var body string
	if mo.ASCII != "" && w >= 100 {
		box := m.artBox(mo, p)
		textW := w - lipgloss.Width(box) - 3
		body = lipgloss.JoinHorizontal(lipgloss.Top, box, "   ", m.tabContent(t, mo, p, textW))
	} else if mo.ASCII != "" && t == tabFacts && w >= artWidth(mo.ASCII)+6 {
		// Narrow screens: facts first, the portrait underneath.
		body = m.tabContent(t, mo, p, w) + "\n\n" + m.artBox(mo, p)
	} else {
		body = m.tabContent(t, mo, p, w)
	}
	return strings.Split("\n"+body, "\n")
}

func (m model) artBox(mo monsters.Monster, p palette) string {
	art := lipgloss.NewStyle().Foreground(p.primary).Render(mo.ASCII)
	if p.silent && m.grain != "" {
		// Flickering film grain above and below, like a worn reel.
		g := helpStyle.Render(m.grain)
		art = g + "\n" + art + "\n" + g
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.accent).
		Padding(1, 2).
		Render(art)
}

func (m model) tabContent(t tab, mo monsters.Monster, p palette, w int) string {
	body := lipgloss.NewStyle().Foreground(p.body)
	var b strings.Builder
	heading := func(s string) { b.WriteString(headingStyle.Render(s) + "\n\n") }

	switch t {
	case tabFacts:
		heading("Terrifying Facts")
		for _, f := range mo.Facts {
			b.WriteString(bullet(f, w, body) + "\n")
		}
		b.WriteString("\n" + metaStyle.Render(wrap("Origin: "+mo.Origin, w)))
		b.WriteString("\n" + metaStyle.Render(wrap("First appearance: "+mo.FirstAppearance(), w)))

	case tabLegend:
		heading("The Legend Behind the Monster")
		b.WriteString(body.Render(wrap(mo.Legend, w)) + "\n\n")
		b.WriteString(metaStyle.Render(wrap("Origin: "+mo.Origin, w)))

	case tabFilm:
		f := mo.Film
		heading(fmt.Sprintf("🎬 %s (%d)", f.Title, f.Year))
		row := func(label, value string) {
			if value != "" {
				b.WriteString(metaStyle.Render(fmt.Sprintf("%-19s", label)) + body.Render(wrap(value, max(w-19, 10))) + "\n")
			}
		}
		row("Studio", f.Studio)
		row("Director", f.Director)
		row("Monster played by", f.Star)
		row("Makeup", f.Makeup)
		row("Released", formatDate(f.ReleaseDate))
		if f.Silent {
			row("Format", "Silent picture")
		}
		if f.Note != "" {
			b.WriteString("\n" + body.Render(wrap(f.Note, w)))
		}

	case tabMyths:
		heading("Myth vs Movie")
		if !m.revealed {
			b.WriteString(hostStyle.Render(wrap("True or false? Make your guesses, then press r to reveal the verdicts.", w)) + "\n\n")
		}
		for i, my := range mo.Myths {
			b.WriteString(body.Render(wrap(fmt.Sprintf("%d. %s", i+1, my.Claim), w)) + "\n")
			switch {
			case !m.revealed:
				b.WriteString(helpStyle.Render("   ? ? ?") + "\n\n")
			case my.True:
				b.WriteString(verdict(trueStyle.Render("✔ TRUE"), my.Explanation, w) + "\n\n")
			default:
				b.WriteString(verdict(mythStyle.Render("✘ MYTH"), my.Explanation, w) + "\n\n")
			}
		}

	case tabQuotes:
		heading("In Their Own Words")
		quote := lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(p.primary).
			PaddingLeft(1)
		for _, q := range mo.Quotes {
			text := body.Italic(true).Render(wrap("“"+q.Text+"”", w-4)) + "\n" +
				helpStyle.Render(wrap("— "+q.Speaker+", "+q.Source, w-4))
			b.WriteString(quote.Render(text) + "\n\n")
		}

	case tabStats:
		heading("Stat Card")
		s := mo.Stats
		for _, st := range []struct {
			name string
			v    int
		}{{"Strength", s.Strength}, {"Speed", s.Speed}, {"Cunning", s.Cunning}, {"Dread", s.Dread}} {
			b.WriteString(metaStyle.Render(fmt.Sprintf("%-9s", st.name)) + statBar(st.v, p) + "\n")
		}
		b.WriteString("\n" + headingStyle.Render("Powers") + "\n")
		for _, pw := range mo.Powers {
			b.WriteString(bullet(pw, w, body) + "\n")
		}
		b.WriteString("\n" + headingStyle.Render("Weaknesses") + "\n")
		for _, wk := range mo.Weaknesses {
			b.WriteString(bullet(wk, w, body) + "\n")
		}
		b.WriteString("\n" + helpStyle.Render("Stats are just for fun."))

	case tabLegacy:
		heading("Sequels, Remakes & Legacy")
		for _, l := range mo.Legacy {
			b.WriteString(bullet(l, w, body) + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// ---- search ----

func (m model) searchResults(w int) string {
	if strings.TrimSpace(m.query) == "" {
		return hostStyle.Render(wrap("Search names, nicknames, facts, films and legends. Try \"lugosi\", \"silver\" or \"paris\".", w))
	}
	if len(m.hits) == 0 {
		return hostStyle.Render("Nothing stirs in the dark... try another search.")
	}
	var lines []string
	for i, h := range m.hits {
		marker := "  "
		name := lipgloss.NewStyle().Foreground(paletteFor(h.Monster).primary).Render(h.Monster.Emoji + " " + h.Monster.Name)
		if i == m.hitCursor {
			marker = lipgloss.NewStyle().Foreground(colorBlood).Render("▶ ")
		}
		lines = append(lines, marker+name)
		lines = append(lines, "    "+helpStyle.Render(truncate("("+h.Field+") "+h.Snippet, w-6)))
	}
	return strings.Join(lines, "\n")
}

// ---- helpers ----

// availableHeight is how many body lines fit between header and footer.
func (m model) availableHeight(header, footer string) int {
	if m.height <= 0 {
		return 0
	}
	return max(m.height-lipgloss.Height(header)-lipgloss.Height(footer)-3, 3)
}

// bodyHeight is availableHeight for the current detail page.
func (m model) bodyHeight() int {
	footer := m.footer(detailKeys)
	return m.availableHeight(m.detailHeader(), footer)
}

func bullet(text string, w int, style lipgloss.Style) string {
	lines := strings.Split(wrap(text, max(w-4, 10)), "\n")
	for i := range lines {
		prefix := "    "
		if i == 0 {
			prefix = "  • "
		}
		lines[i] = prefix + style.Render(strings.TrimRight(lines[i], " "))
	}
	return strings.Join(lines, "\n")
}

// verdict lays out a myth verdict with its explanation indented beside it.
func verdict(label, explanation string, w int) string {
	lead := "   " + label + " "
	return lipgloss.JoinHorizontal(lipgloss.Top, lead, hostStyle.Render(wrap(explanation, max(w-lipgloss.Width(lead), 10))))
}

// joinFit joins items with sep, starting a new line rather than splitting an item.
func joinFit(items []string, sep string, width int) string {
	var lines []string
	line := ""
	for _, it := range items {
		switch {
		case line == "":
			line = it
		case lipgloss.Width(line+sep+it) <= width:
			line += sep + it
		default:
			lines = append(lines, line)
			line = it
		}
	}
	return strings.Join(append(lines, line), "\n")
}

func statBar(v int, p palette) string {
	v = min(max(v, 0), 10)
	return lipgloss.NewStyle().Foreground(p.primary).Render(strings.Repeat("█", v)) +
		helpStyle.Render(strings.Repeat("░", 10-v)) +
		fmt.Sprintf(" %2d", v)
}

func truncate(s string, w int) string {
	if w <= 1 || lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r))+1 > w {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

func formatDate(iso string) string {
	if iso == "" {
		return ""
	}
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return iso
	}
	return t.Format("2 January 2006")
}

func debutYear(mo monsters.Monster) string {
	if mo.Debut.Year != 0 {
		return fmt.Sprint(mo.Debut.Year)
	}
	return mo.Debut.Era
}

func packName(id string) string {
	for _, p := range monsters.GetPacks() {
		if p.ID == id {
			return p.Name
		}
	}
	return id
}
