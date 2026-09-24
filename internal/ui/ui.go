package ui

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hungovercoders/terminal-of-terror/internal/calendar"
	"github.com/hungovercoders/terminal-of-terror/internal/host"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
)

type screen int

const (
	screenIntro screen = iota
	screenGallery
	screenDetail
)

// Options configures the monster explorer.
type Options struct {
	ShowAll bool   // open on the gallery of every monster
	StartID string // open on this monster's page
	NoIntro bool   // skip the Channel 13 opening
	Seed    int64  // random seed; 0 means use the clock
	Now     time.Time
}

type tickMsg time.Time

// Animation speeds: the intro runs fast; the silent-film grain only needs
// an occasional flicker, so idle silent pages don't redraw constantly.
const (
	introTick = 70 * time.Millisecond
	grainTick = 600 * time.Millisecond
)

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// tickEvery is how often the current screen animates.
func (m model) tickEvery() time.Duration {
	if m.screen == screenIntro {
		return introTick
	}
	return grainTick
}

type model struct {
	monsters []monsters.Monster
	rng      *rand.Rand

	screen screen
	next   screen // where the intro leads

	index    int // monster shown on the detail page
	cursor   int // gallery selection
	tab      int // index into tabsFor(current monster)
	scroll   int
	revealed bool // myth verdicts visible
	showHelp bool

	searching bool
	query     string
	hits      []monsters.SearchHit
	hitCursor int

	width, height int

	seen map[string]bool // monster pages visited

	frame    int
	ticking  bool
	noise    string
	grain    string
	greeting string
	notes    []string // what's special about tonight
	signOff  string
	quitting bool
}

func newModel(opts Options) model {
	seed := opts.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(seed))
	m := model{
		monsters: monsters.GetAllMonsters(),
		rng:      r,
		greeting: host.Greeting(r),
		next:     screenDetail,
		seen:     map[string]bool{},
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	m.notes = calendar.Notes(now, m.monsters)
	if calendar.IsHalloween(now) {
		m.greeting = "Happy Halloween, creatures of the night! Every monster in the vault is awake tonight, and so are you. I'm Count Cathode, and this is the biggest Creature Feature of the year."
	}
	if opts.ShowAll {
		m.next = screenGallery
	}
	for i, mo := range m.monsters {
		if mo.ID == opts.StartID {
			m.index, m.cursor = i, i
			m.next = screenDetail
		}
	}
	// Jumping straight to a monster means you know what you want: no intro.
	if opts.NoIntro || opts.StartID != "" {
		m.enter(m.next)
	} else {
		m.screen = screenIntro
		m.noise = noise(r, 60, 9)
	}
	m.ticking = m.needsTick() // Init starts the loop
	return m
}

func (m model) current() monsters.Monster { return m.monsters[m.index] }

// needsTick reports whether anything on screen is animated.
func (m model) needsTick() bool {
	if m.screen == screenIntro {
		return !m.introDone()
	}
	return m.screen == screenDetail && m.current().IsSilent()
}

func (m *model) ensureTick() tea.Cmd {
	if m.ticking || !m.needsTick() {
		return nil
	}
	m.ticking = true
	return tick(m.tickEvery())
}

func (m model) Init() tea.Cmd {
	if m.ticking {
		return tick(m.tickEvery())
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.clampScroll()
		return m, nil

	case tickMsg:
		m.ticking = false
		m.frame++
		if m.screen == screenIntro && m.frame < staticFrames {
			m.noise = noise(m.rng, min(max(m.contentWidth()-4, 20), 70), 9)
		}
		if m.screen == screenDetail && m.current().IsSilent() {
			m.grain = noise(m.rng, artWidth(m.current().ASCII)+4, 1)
		}
		return m, m.ensureTick()

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m.quit()
		}
		if m.screen == screenIntro {
			if msg.String() == "q" {
				return m.quit()
			}
			if !m.introDone() {
				m.frame = introDoneFrame(m.greeting)
				return m, nil
			}
			m.enter(m.next)
			return m, m.ensureTick()
		}
		if m.searching {
			return m.updateSearch(msg)
		}
		switch msg.String() {
		case "q":
			return m.quit()
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "/":
			m.searching, m.query, m.hits, m.hitCursor = true, "", nil, 0
			return m, nil
		}
		if m.screen == screenGallery {
			return m.updateGallery(msg)
		}
		return m.updateDetail(msg)
	}
	return m, nil
}

func (m model) quit() (tea.Model, tea.Cmd) {
	m.quitting = true
	m.signOff = host.SignOff(m.rng)
	return m, tea.Quit
}

func (m model) updateGallery(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.cursor = (m.cursor - 1 + len(m.monsters)) % len(m.monsters)
	case "down", "j":
		m.cursor = (m.cursor + 1) % len(m.monsters)
	case "home":
		m.cursor = 0
	case "end":
		m.cursor = len(m.monsters) - 1
	case "enter", " ", "right", "l":
		m.open(m.cursor)
		return m, m.ensureTick()
	}
	return m, nil
}

func (m model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	tabs := tabsFor(m.current())
	switch key := msg.String(); key {
	case "right", "l", "n":
		m.open((m.index + 1) % len(m.monsters))
		return m, m.ensureTick()
	case "left", "h", "p":
		m.open((m.index - 1 + len(m.monsters)) % len(m.monsters))
		return m, m.ensureTick()
	case "tab":
		m.setTab((m.tab + 1) % len(tabs))
	case "shift+tab":
		m.setTab((m.tab - 1 + len(tabs)) % len(tabs))
	case "down", "j":
		m.scroll++
	case "up", "k":
		m.scroll--
	case "pgdown", " ":
		m.scroll += max(m.bodyHeight()-2, 1)
	case "pgup":
		m.scroll -= max(m.bodyHeight()-2, 1)
	case "home":
		m.scroll = 0
	case "end":
		m.scroll = 1 << 30
	case "r":
		m.revealed = !m.revealed
	case "g", "esc", "backspace":
		m.cursor = m.index
		m.screen = screenGallery
		return m, nil
	default:
		if n, err := strconv.Atoi(key); err == nil && n >= 1 && n <= len(tabs) {
			m.setTab(n - 1)
		}
	}
	m.clampScroll()
	return m, nil
}

func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.searching = false
		return m, nil
	case tea.KeyEnter:
		if len(m.hits) > 0 {
			id := m.hits[m.hitCursor].Monster.ID
			m.searching = false
			for i, mo := range m.monsters {
				if mo.ID == id {
					m.open(i)
				}
			}
			return m, m.ensureTick()
		}
		return m, nil
	case tea.KeyUp:
		if m.hitCursor > 0 {
			m.hitCursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.hitCursor < len(m.hits)-1 {
			m.hitCursor++
		}
		return m, nil
	case tea.KeyBackspace:
		if r := []rune(m.query); len(r) > 0 {
			m.query = string(r[:len(r)-1])
		}
	case tea.KeyRunes, tea.KeySpace:
		if msg.Type == tea.KeySpace {
			m.query += " "
		} else {
			m.query += string(msg.Runes)
		}
	default:
		return m, nil
	}
	m.hits = monsters.Search(m.query)
	m.hitCursor = 0
	return m, nil
}

// open shows monster i on the detail page, keeping the same kind of tab
// where the new monster has one.
func (m *model) open(i int) {
	prev := tabFacts
	if m.screen == screenDetail {
		if tabs := tabsFor(m.current()); m.tab < len(tabs) {
			prev = tabs[m.tab]
		}
	}
	m.index, m.cursor = i, i
	m.enter(screenDetail)
	m.tab, m.scroll, m.revealed = 0, 0, false
	for j, t := range tabsFor(m.current()) {
		if t == prev {
			m.tab = j
		}
	}
}

// enter switches screen, noting monster pages as visited.
func (m *model) enter(s screen) {
	m.screen = s
	if s == screenDetail {
		m.seen[m.current().ID] = true
	}
}

func (m *model) setTab(i int) {
	m.tab, m.scroll = i, 0
}

func (m *model) clampScroll() {
	if m.screen != screenDetail {
		return
	}
	maxScroll := 0
	if h := m.bodyHeight(); h > 0 {
		maxScroll = max(len(m.bodyLines())-h, 0)
	}
	m.scroll = min(max(m.scroll, 0), maxScroll)
}

// contentWidth is the usable width, assuming 80 columns until told otherwise.
func (m model) contentWidth() int {
	if m.width <= 0 {
		return 80
	}
	return m.width
}

// RunUI starts the interactive explorer and returns the ids of the
// monster pages visited.
func RunUI(opts Options) ([]string, error) {
	p := tea.NewProgram(newModel(opts), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm, ok := final.(model)
	if !ok {
		return nil, nil
	}
	if fm.signOff != "" {
		fmt.Println(hostStyle.Render("📺 " + fm.signOff))
	}
	var seen []string
	for id := range fm.seen {
		seen = append(seen, id)
	}
	return seen, nil
}
