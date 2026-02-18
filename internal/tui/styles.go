package tui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// renderMarkdown renders markdown content using glamour.
func renderMarkdown(width int, content string) string {
	if content == "" {
		return ""
	}

	// Use a fixed style to avoid slow terminal background detection.
	// We use "dark" as a sensible default for TUI applications.
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}

	out, err := r.Render(content)
	if err != nil {
		return content
	}

	return strings.TrimSpace(out)
}

// Colors
var (
	colorPrimary   = lipgloss.Color("#E0A458") // warm tea gold
	colorSecondary = lipgloss.Color("#A8D8B9") // soft green
	colorMuted     = lipgloss.Color("#666666")
	colorHighlight = lipgloss.Color("#FFFBE6") // cream
	colorDanger    = lipgloss.Color("#E06C75")
	colorBorder    = lipgloss.Color("#444444")
	colorSelected  = lipgloss.Color("#E0A458")
)

// Layout styles
var (
	// App-level wrapper
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	// Title bar
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	// Left pane (menu / list)
	leftPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	// Right pane (preview / content)
	rightPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	// Focused pane border
	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(1, 2)
)

// List item styles
var (
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorHighlight).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))

	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)
)

// Help bar
var (
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	helpBarStyle = lipgloss.NewStyle().
			MarginTop(1)
)

// Status messages
var (
	successStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorDanger)
)

// Reminders
var (
	reminderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5C07B")).
			Bold(true)

	reminderItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5C07B"))
)

// Misc
var (
	cursorStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)

	headerStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginBottom(1)

	// paneHeaderStyle is like headerStyle but without MarginBottom,
	// for use inside bordered panes where vertical space is tight.
	paneHeaderStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	previewHeaderStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Bold(true).
				MarginBottom(1)
)

// helpEntry renders a single "[key] description" help item.
func helpEntry(key, desc string) string {
	return helpKeyStyle.Render("["+key+"]") + " " + helpDescStyle.Render(desc)
}

// Constants for layout
const (
	leftPaneWidthFraction = 0.30
	minLeftPaneWidth      = 24
	minRightPaneWidth     = 30
	defaultTerminalWidth  = 80
	defaultTerminalHeight = 24
)
