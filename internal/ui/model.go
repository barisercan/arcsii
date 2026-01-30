package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/barisercan/arcsii/internal/commands"
	"github.com/barisercan/arcsii/internal/watcher"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			Background(lipgloss.Color("#1A1A2E")).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4ECDC4")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98D8C8")).
			Background(lipgloss.Color("#1A1A2E")).
			Padding(0, 1)

	// Live event styles
	createStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	modifyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	deleteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	renameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")).
			Bold(true)

	filePathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4"))

	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	pulseColors = []string{"#FF6B6B", "#FF8E8E", "#FFB0B0", "#FF8E8E", "#FF6B6B"}

	// Default commands to cycle through
	defaultCommands = []string{"/watch", "/tree", "/uml", "/ascii", "/deps", "/changes", "/stats", "/funcs", "/help"}
)

// EventDisplay wraps a file event with display state
type EventDisplay struct {
	Event     watcher.FileEvent
	Age       int // for animation
	Highlight bool
}

type Model struct {
	targetDir    string
	input        textinput.Model
	viewport     viewport.Model
	content      string
	status       string
	width        int
	height       int
	ready        bool
	cmdRegistry  *commands.Registry
	history      []string
	historyIndex int

	// Live watch mode
	watcher       *watcher.Watcher
	events        []EventDisplay
	watchMode     bool
	tick          int
	pulseIndex    int
	gitAnimation  string // Current git animation type
	gitAnimTick   int    // Animation frame counter
}

// Messages
type fileEventMsg watcher.FileEvent
type tickMsg time.Time

func NewModel(targetDir string) Model {
	ti := textinput.New()
	ti.Placeholder = "Type a command (e.g., /help, /tree, /uml) or watch live changes..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4")).Bold(true)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Start file watcher
	w, _ := watcher.New(targetDir)

	return Model{
		targetDir:    targetDir,
		input:        ti,
		content:      "", // Will be set in Init
		status:       "Watching",
		cmdRegistry:  commands.NewRegistry(targetDir),
		history:      []string{},
		historyIndex: -1,
		watcher:      w,
		events:       []EventDisplay{},
		watchMode:    true,
		tick:         0,
		pulseIndex:   0,
	}
}

func (m Model) Init() tea.Cmd {
	// Start the watcher and tick timer
	if m.watcher != nil {
		m.watcher.Start()
	}
	return tea.Batch(
		textinput.Blink,
		listenForEvents(m.watcher),
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func listenForEvents(w *watcher.Watcher) tea.Cmd {
	if w == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case event := <-w.Events:
			return fileEventMsg(event)
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tickMsg:
		m.tick++
		m.pulseIndex = (m.pulseIndex + 1) % len(pulseColors)

		// Handle git animation
		if m.gitAnimation != "" {
			m.gitAnimTick++
			if m.gitAnimTick > 50 { // 5 seconds
				m.gitAnimation = ""
				m.gitAnimTick = 0
			}
		}

		// Age events and remove old highlights
		for i := range m.events {
			m.events[i].Age++
			if m.events[i].Age > 30 { // 3 seconds
				m.events[i].Highlight = false
			}
		}

		// Update viewport content if in watch mode
		if m.watchMode {
			m.content = m.renderLiveView()
			m.viewport.SetContent(m.content)
		}

		return m, tea.Batch(tickCmd(), listenForEvents(m.watcher))

	case fileEventMsg:
		event := watcher.FileEvent(msg)

		// Check for git operations and trigger animation
		if event.IsGitOp && event.GitOp != "" {
			m.gitAnimation = event.GitOp
			m.gitAnimTick = 0
		}

		// Add new event at the beginning
		m.events = append([]EventDisplay{{
			Event:     event,
			Age:       0,
			Highlight: true,
		}}, m.events...)

		// Keep only last 50 events
		if len(m.events) > 50 {
			m.events = m.events[:50]
		}

		if event.IsGitOp {
			m.status = fmt.Sprintf("Git %s detected!", event.GitOp)
		} else {
			m.status = fmt.Sprintf("File %s: %s", event.Operation, event.Name)
		}

		return m, listenForEvents(m.watcher)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.watcher != nil {
				m.watcher.Stop()
			}
			return m, tea.Quit
		case "enter":
			cmd := strings.TrimSpace(m.input.Value())
			if cmd != "" {
				// Add to history
				m.history = append(m.history, cmd)
				m.historyIndex = len(m.history)

				// Check for special commands
				cmdLower := strings.ToLower(strings.TrimPrefix(cmd, "/"))
				if cmdLower == "watch" || cmdLower == "live" || cmdLower == "w" {
					m.watchMode = true
					m.content = m.renderLiveView()
					m.status = "Watching"
				} else {
					m.watchMode = false
					m.content, m.status = m.cmdRegistry.Execute(cmd)
				}

				m.input.Reset()
				m.viewport.SetContent(m.content)
				m.viewport.GotoTop()
			}
		case "up":
			// Combine history with default commands for cycling
			allCommands := append(m.history, defaultCommands...)
			if len(allCommands) > 0 {
				if m.historyIndex <= 0 {
					m.historyIndex = len(allCommands) - 1
				} else {
					m.historyIndex--
				}
				m.input.SetValue(allCommands[m.historyIndex])
				m.input.CursorEnd()
			}
			return m, nil
		case "down":
			allCommands := append(m.history, defaultCommands...)
			if len(allCommands) > 0 {
				if m.historyIndex >= len(allCommands)-1 {
					m.historyIndex = 0
				} else {
					m.historyIndex++
				}
				m.input.SetValue(allCommands[m.historyIndex])
				m.input.CursorEnd()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 5
		footerHeight := 4
		vpHeight := m.height - headerHeight - footerHeight

		if !m.ready {
			m.viewport = viewport.New(m.width-4, vpHeight)
			m.content = m.renderLiveView()
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = m.width - 4
			m.viewport.Height = vpHeight
		}

		m.input.Width = m.width - 10
	}

	m.input, tiCmd = m.input.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) renderLiveView() string {
	var sb strings.Builder

	// Check if we should show git animation
	if m.gitAnimation != "" {
		sb.WriteString(m.renderGitAnimation())
		sb.WriteString("\n\n")
	}

	// Animated header
	pulseColor := pulseColors[m.pulseIndex]
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(pulseColor)).
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(pulseColor)).
		Padding(0, 2)

	// Spinning animation
	spinners := []string{"◐", "◓", "◑", "◒"}
	spinner := spinners[m.tick%len(spinners)]

	sb.WriteString(headerStyle.Render(fmt.Sprintf("%s LIVE FILE MONITOR", spinner)))
	sb.WriteString("\n\n")

	if len(m.events) == 0 && m.gitAnimation == "" {
		// Waiting animation
		dots := strings.Repeat(".", (m.tick/5)%4)
		waiting := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true).
			Render(fmt.Sprintf("    Watching for changes%s", dots))

		sb.WriteString(waiting)
		sb.WriteString("\n\n")

		// Show helpful tip
		tip := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4")).
			Render("    💡 Make changes to any file and watch them appear here!")
		sb.WriteString(tip)
		sb.WriteString("\n\n")

		// ASCII art pulse
		art := m.renderWaitingAnimation()
		sb.WriteString(art)
	} else {
		// Render events
		for i, ed := range m.events {
			if i >= 20 {
				break // Show max 20 events
			}
			sb.WriteString(m.renderEvent(ed))
			sb.WriteString("\n")
		}
	}

	// Footer with instructions
	sb.WriteString("\n")
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Render("    Type /help for commands, /tree for file structure")
	sb.WriteString(footer)

	return sb.String()
}

func (m Model) renderEvent(ed EventDisplay) string {
	var opStyle lipgloss.Style
	var icon string

	switch ed.Event.Operation {
	case "created":
		opStyle = createStyle
		icon = "✚"
	case "modified":
		opStyle = modifyStyle
		icon = "✎"
	case "deleted":
		opStyle = deleteStyle
		icon = "✖"
	case "renamed":
		opStyle = renameStyle
		icon = "↻"
	default:
		opStyle = modifyStyle
		icon = "•"
	}

	// Highlight effect for new events
	if ed.Highlight {
		opStyle = opStyle.Background(lipgloss.Color("#1F2937"))
	}

	// Format time
	ago := time.Since(ed.Event.Time)
	var timeStr string
	if ago < time.Second {
		timeStr = "just now"
	} else if ago < time.Minute {
		timeStr = fmt.Sprintf("%ds ago", int(ago.Seconds()))
	} else {
		timeStr = fmt.Sprintf("%dm ago", int(ago.Minutes()))
	}

	// Get file extension for icon
	fileIcon := getFileIcon(ed.Event.Name)

	// Build the line
	line := fmt.Sprintf("    %s %s  %s  %s  %s",
		opStyle.Render(icon),
		opStyle.Render(fmt.Sprintf("%-10s", ed.Event.Operation)),
		fileIcon,
		filePathStyle.Render(ed.Event.Path),
		timeStyle.Render(timeStr),
	)

	return line
}

func (m Model) renderGitAnimation() string {
	var art string
	frame := m.gitAnimTick

	switch m.gitAnimation {
	case "commit":
		art = m.renderCommitAnimation(frame)
	case "push":
		art = m.renderPushAnimation(frame)
	case "pull", "fetch":
		art = m.renderPullAnimation(frame)
	case "merge":
		art = m.renderMergeAnimation(frame)
	case "checkout":
		art = m.renderCheckoutAnimation(frame)
	case "rebase":
		art = m.renderRebaseAnimation(frame)
	case "stash":
		art = m.renderStashAnimation(frame)
	default:
		return ""
	}

	return art
}

func (m Model) renderCommitAnimation(frame int) string {
	colors := []string{"#10B981", "#34D399", "#6EE7B7", "#34D399", "#10B981"}
	color := colors[frame%len(colors)]

	frames := []string{
		`
    ╔═══════════════════════════════════════════════════════╗
    ║                                                       ║
    ║      ██████╗ ██████╗ ███╗   ███╗███╗   ███╗██╗████████╗║
    ║     ██╔════╝██╔═══██╗████╗ ████║████╗ ████║██║╚══██╔══╝║
    ║     ██║     ██║   ██║██╔████╔██║██╔████╔██║██║   ██║   ║
    ║     ██║     ██║   ██║██║╚██╔╝██║██║╚██╔╝██║██║   ██║   ║
    ║     ╚██████╗╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║██║   ██║   ║
    ║      ╚═════╝ ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝╚═╝   ╚═╝   ║
    ║                                                       ║
    ║              [  ✓  ]  Changes saved!                  ║
    ╚═══════════════════════════════════════════════════════╝`,
		`
    ╔═══════════════════════════════════════════════════════╗
    ║                    * * *                              ║
    ║      ██████╗ ██████╗ ███╗   ███╗███╗   ███╗██╗████████╗║
    ║     ██╔════╝██╔═══██╗████╗ ████║████╗ ████║██║╚══██╔══╝║
    ║     ██║     ██║   ██║██╔████╔██║██╔████╔██║██║   ██║   ║
    ║     ██║     ██║   ██║██║╚██╔╝██║██║╚██╔╝██║██║   ██║   ║
    ║     ╚██████╗╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║██║   ██║   ║
    ║      ╚═════╝ ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝╚═╝   ╚═╝   ║
    ║                   * * * *                             ║
    ║              [ ✓✓✓ ]  Changes saved!                  ║
    ╚═══════════════════════════════════════════════════════╝`,
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(frames[frame/3%len(frames)])
}

func (m Model) renderPushAnimation(frame int) string {
	colors := []string{"#3B82F6", "#60A5FA", "#93C5FD", "#60A5FA", "#3B82F6"}
	color := colors[frame%len(colors)]

	// Animated arrow going up
	arrows := []string{
		"        ▲        ",
		"       ▲▲▲       ",
		"      ▲▲▲▲▲      ",
		"     ▲▲▲▲▲▲▲     ",
		"    ▲▲▲▲▲▲▲▲▲    ",
	}
	arrowFrame := arrows[frame/2%len(arrows)]

	art := fmt.Sprintf(`
    ╔═══════════════════════════════════════════════════════╗
    ║                                                       ║
    ║     ██████╗ ██╗   ██╗███████╗██╗  ██╗██╗███╗   ██╗██╗ ║
    ║     ██╔══██╗██║   ██║██╔════╝██║  ██║██║████╗  ██║██║ ║
    ║     ██████╔╝██║   ██║███████╗███████║██║██╔██╗ ██║██║ ║
    ║     ██╔═══╝ ██║   ██║╚════██║██╔══██║██║██║╚██╗██║╚═╝ ║
    ║     ██║     ╚██████╔╝███████║██║  ██║██║██║ ╚████║██╗ ║
    ║     ╚═╝      ╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝╚═╝ ║
    ║                                                       ║
    ║                  %s                   ║
    ║              Pushing to remote...                     ║
    ╚═══════════════════════════════════════════════════════╝`, arrowFrame)

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(art)
}

func (m Model) renderPullAnimation(frame int) string {
	colors := []string{"#8B5CF6", "#A78BFA", "#C4B5FD", "#A78BFA", "#8B5CF6"}
	color := colors[frame%len(colors)]

	// Animated arrow going down
	arrows := []string{
		"    ▼▼▼▼▼▼▼▼▼    ",
		"     ▼▼▼▼▼▼▼     ",
		"      ▼▼▼▼▼      ",
		"       ▼▼▼       ",
		"        ▼        ",
	}
	arrowFrame := arrows[frame/2%len(arrows)]

	art := fmt.Sprintf(`
    ╔═══════════════════════════════════════════════════════╗
    ║                                                       ║
    ║     ██████╗ ██╗   ██╗██╗     ██╗     ██╗███╗   ██╗██╗ ║
    ║     ██╔══██╗██║   ██║██║     ██║     ██║████╗  ██║██║ ║
    ║     ██████╔╝██║   ██║██║     ██║     ██║██╔██╗ ██║██║ ║
    ║     ██╔═══╝ ██║   ██║██║     ██║     ██║██║╚██╗██║╚═╝ ║
    ║     ██║     ╚██████╔╝███████╗███████╗██║██║ ╚████║██╗ ║
    ║     ╚═╝      ╚═════╝ ╚══════╝╚══════╝╚═╝╚═╝  ╚═══╝╚═╝ ║
    ║                                                       ║
    ║                  %s                   ║
    ║              Pulling from remote...                   ║
    ╚═══════════════════════════════════════════════════════╝`, arrowFrame)

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(art)
}

func (m Model) renderMergeAnimation(frame int) string {
	colors := []string{"#F59E0B", "#FBBF24", "#FCD34D", "#FBBF24", "#F59E0B"}
	color := colors[frame%len(colors)]

	// Animated merge lines
	mergeFrames := []string{
		"    \\     /    ",
		"     \\   /     ",
		"      \\ /      ",
		"       Y       ",
		"       |       ",
	}
	mergeFrame := mergeFrames[frame/2%len(mergeFrames)]

	art := fmt.Sprintf(`
    ╔═══════════════════════════════════════════════════════╗
    ║                                                       ║
    ║     ███╗   ███╗███████╗██████╗  ██████╗ ███████╗██╗   ║
    ║     ████╗ ████║██╔════╝██╔══██╗██╔════╝ ██╔════╝██║   ║
    ║     ██╔████╔██║█████╗  ██████╔╝██║  ███╗█████╗  ██║   ║
    ║     ██║╚██╔╝██║██╔══╝  ██╔══██╗██║   ██║██╔══╝  ╚═╝   ║
    ║     ██║ ╚═╝ ██║███████╗██║  ██║╚██████╔╝███████╗██╗   ║
    ║     ╚═╝     ╚═╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝   ║
    ║                                                       ║
    ║                  %s                    ║
    ║              Merging branches...                      ║
    ╚═══════════════════════════════════════════════════════╝`, mergeFrame)

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(art)
}

func (m Model) renderCheckoutAnimation(frame int) string {
	colors := []string{"#EC4899", "#F472B6", "#F9A8D4", "#F472B6", "#EC4899"}
	color := colors[frame%len(colors)]

	art := `
    ╔═══════════════════════════════════════════════════════╗
    ║                                                       ║
    ║      ██████╗██╗  ██╗███████╗ ██████╗██╗  ██╗ ██████╗  ║
    ║     ██╔════╝██║  ██║██╔════╝██╔════╝██║ ██╔╝██╔═══██╗ ║
    ║     ██║     ███████║█████╗  ██║     █████╔╝ ██║   ██║ ║
    ║     ██║     ██╔══██║██╔══╝  ██║     ██╔═██╗ ██║   ██║ ║
    ║     ╚██████╗██║  ██║███████╗╚██████╗██║  ██╗╚██████╔╝ ║
    ║      ╚═════╝╚═╝  ╚═╝╚══════╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝  ║
    ║                                                       ║
    ║              ◇────────────────────◆                   ║
    ║              Switching branches...                    ║
    ╚═══════════════════════════════════════════════════════╝`

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(art)
}

func (m Model) renderRebaseAnimation(frame int) string {
	colors := []string{"#EF4444", "#F87171", "#FCA5A5", "#F87171", "#EF4444"}
	color := colors[frame%len(colors)]

	// Animated rebase blocks
	blocks := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	blockFrame := blocks[frame%len(blocks)]

	art := fmt.Sprintf(`
    ╔═══════════════════════════════════════════════════════╗
    ║                                                       ║
    ║     ██████╗ ███████╗██████╗  █████╗ ███████╗███████╗  ║
    ║     ██╔══██╗██╔════╝██╔══██╗██╔══██╗██╔════╝██╔════╝  ║
    ║     ██████╔╝█████╗  ██████╔╝███████║███████╗█████╗    ║
    ║     ██╔══██╗██╔══╝  ██╔══██╗██╔══██║╚════██║██╔══╝    ║
    ║     ██║  ██║███████╗██████╔╝██║  ██║███████║███████╗  ║
    ║     ╚═╝  ╚═╝╚══════╝╚═════╝ ╚═╝  ╚═╝╚══════╝╚══════╝  ║
    ║                                                       ║
    ║          %s%s%s%s%s%s%s%s  Rebasing...              ║
    ╚═══════════════════════════════════════════════════════╝`,
		blockFrame, blockFrame, blockFrame, blockFrame, blockFrame, blockFrame, blockFrame, blockFrame)

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(art)
}

func (m Model) renderStashAnimation(frame int) string {
	colors := []string{"#14B8A6", "#2DD4BF", "#5EEAD4", "#2DD4BF", "#14B8A6"}
	color := colors[frame%len(colors)]

	art := `
    ╔═══════════════════════════════════════════════════════╗
    ║                                                       ║
    ║     ███████╗████████╗ █████╗ ███████╗██╗  ██╗██╗      ║
    ║     ██╔════╝╚══██╔══╝██╔══██╗██╔════╝██║  ██║██║      ║
    ║     ███████╗   ██║   ███████║███████╗███████║██║      ║
    ║     ╚════██║   ██║   ██╔══██║╚════██║██╔══██║╚═╝      ║
    ║     ███████║   ██║   ██║  ██║███████║██║  ██║██╗      ║
    ║     ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝      ║
    ║                                                       ║
    ║              📦 Changes stashed away!                 ║
    ╚═══════════════════════════════════════════════════════╝`

	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(art)
}

func (m Model) renderWaitingAnimation() string {
	frames := []string{
		`
        ┌─────────────────────────────┐
        │    ░░░░░░░░░░░░░░░░░░░░    │
        │    ░░                ░░    │
        │    ░░   WATCHING     ░░    │
        │    ░░                ░░    │
        │    ░░░░░░░░░░░░░░░░░░░░    │
        └─────────────────────────────┘`,
		`
        ┌─────────────────────────────┐
        │    ▒▒░░░░░░░░░░░░░░░░▒▒    │
        │    ▒▒                ▒▒    │
        │    ▒▒   WATCHING     ▒▒    │
        │    ▒▒                ▒▒    │
        │    ▒▒░░░░░░░░░░░░░░░░▒▒    │
        └─────────────────────────────┘`,
		`
        ┌─────────────────────────────┐
        │    ▓▓▒▒░░░░░░░░░░░░▒▒▓▓    │
        │    ▓▓                ▓▓    │
        │    ▓▓   WATCHING     ▓▓    │
        │    ▓▓                ▓▓    │
        │    ▓▓▒▒░░░░░░░░░░░░▒▒▓▓    │
        └─────────────────────────────┘`,
		`
        ┌─────────────────────────────┐
        │    ██▓▓▒▒░░░░░░░░▒▒▓▓██    │
        │    ██                ██    │
        │    ██   WATCHING     ██    │
        │    ██                ██    │
        │    ██▓▓▒▒░░░░░░░░▒▒▓▓██    │
        └─────────────────────────────┘`,
	}

	frame := frames[(m.tick/3)%len(frames)]
	return lipgloss.NewStyle().Foreground(lipgloss.Color(pulseColors[m.pulseIndex])).Render(frame)
}

func getFileIcon(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go":
		return "🔷"
	case ".js", ".ts", ".jsx", ".tsx":
		return "🟨"
	case ".py":
		return "🐍"
	case ".rs":
		return "🦀"
	case ".md":
		return "📝"
	case ".json":
		return "📋"
	case ".yaml", ".yml":
		return "⚙️"
	case ".html":
		return "🌐"
	case ".css", ".scss":
		return "🎨"
	case ".sql":
		return "🗄️"
	case ".sh":
		return "💻"
	default:
		return "📄"
	}
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// Header with logo
	var modeIndicator string
	if m.watchMode {
		modeIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true).
			Render(" ● LIVE")
	} else {
		modeIndicator = ""
	}

	header := titleStyle.Render("◈ ARCSII") + modeIndicator + "  " + helpStyle.Render("Terminal Architecture Visualizer")

	// Content viewport
	content := m.viewport.View()

	// Input area
	input := inputStyle.Render(m.input.View())

	// Status bar
	status := statusStyle.Render("⚡ " + m.status + " │ " + m.targetDir + " │ ↑↓ scroll │ esc quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
		"",
		input,
		status,
	)
}
