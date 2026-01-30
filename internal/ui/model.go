package ui

import (
	"strings"

	"github.com/barisercan/arcsii/internal/commands"
	"github.com/barisercan/arcsii/internal/renderer"
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

	contentStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#45B7D1")).
			Padding(1, 2)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98D8C8")).
			Background(lipgloss.Color("#1A1A2E")).
			Padding(0, 1)
)

type Model struct {
	targetDir   string
	input       textinput.Model
	viewport    viewport.Model
	content     string
	status      string
	width       int
	height      int
	ready       bool
	cmdRegistry *commands.Registry
}

func NewModel(targetDir string) Model {
	ti := textinput.New()
	ti.Placeholder = "Type a command (e.g., /help, /tree, /uml)..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4")).Bold(true)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	return Model{
		targetDir:   targetDir,
		input:       ti,
		content:     renderer.RenderWelcome(),
		status:      "Ready",
		cmdRegistry: commands.NewRegistry(targetDir),
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			cmd := strings.TrimSpace(m.input.Value())
			if cmd != "" {
				m.content, m.status = m.cmdRegistry.Execute(cmd)
				m.input.Reset()
				m.viewport.SetContent(m.content)
				m.viewport.GotoTop()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 5
		footerHeight := 4
		vpHeight := m.height - headerHeight - footerHeight

		if !m.ready {
			m.viewport = viewport.New(m.width-4, vpHeight)
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

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// Header with logo
	header := titleStyle.Render("◈ ARCSII") + "  " + helpStyle.Render("Terminal Architecture Visualizer")

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
