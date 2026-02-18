package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gabrielfornes/teatime/internal/storage"
)

// screen represents which screen is currently active.
type screen int

const (
	screenProjectList screen = iota
	screenProjectView
	screenNoteList
	screenEdit
)

// Model is the root Bubble Tea model for teatime.
type Model struct {
	store *storage.Store

	// Terminal dimensions
	width  int
	height int

	// Current screen
	screen screen

	// Project list state
	projects      []string
	projectCursor int
	creatingNew   bool
	newNameInput  textarea.Model

	// Currently selected project
	currentProject string

	// Project view state (menu on the left, today's note on the right)
	menuCursor        int
	todayNote         string
	todayNoteRendered string
	reminders         []storage.Reminder

	// Note list state
	noteCategory        storage.Category
	notes               []storage.NoteFile
	noteCursor          int
	previewNote         string
	previewNoteRendered string
	lastRenderedPreview string
	lastRenderedWidth   int

	// Edit mode state
	editTextarea    textarea.Model
	editCategory    storage.Category
	editNoteName    string
	editDirty       bool
	editRef         string         // reference content from the level below
	editRefRendered string         // rendered version for the viewport
	editViewport    viewport.Model // scrollable right pane for reference content
	editFocusLeft   bool           // true = textarea focused, false = viewport focused

	// Status message (shown briefly)
	statusMsg string
	statusErr bool

	// Error state
	err error
}

// Menu items shown in the project view
var menuItems = []struct {
	key      string
	label    string
	category storage.Category
}{
	{"e", "Edit today", storage.CategoryDaily},
	{"d", "Daily notes", storage.CategoryDaily},
	{"w", "Weekly notes", storage.CategoryWeekly},
	{"m", "Monthly notes", storage.CategoryMonthly},
	{"Q", "Quarterly notes", storage.CategoryQuarterly},
	{"y", "Yearly notes", storage.CategoryYearly},
}

// NewModel creates and returns a new root model.
func NewModel(store *storage.Store) Model {
	ta := textarea.New()
	ta.Placeholder = "Enter project name..."
	ta.CharLimit = 64
	ta.SetWidth(30)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false

	editTa := textarea.New()
	editTa.Placeholder = "Start writing..."
	editTa.ShowLineNumbers = false
	editTa.SetWidth(60)
	editTa.SetHeight(15)

	return Model{
		store:        store,
		screen:       screenProjectList,
		width:        defaultTerminalWidth,
		height:       defaultTerminalHeight,
		newNameInput: ta,
		editTextarea: editTa,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("🍵 teatime"),
		m.loadProjects,
	)
}

// Layout calculation helpers
func (m Model) projectViewLayout() (int, int, int) {
	usableWidth := m.width - 8
	leftWidth := int(float64(usableWidth) * leftPaneWidthFraction)
	if leftWidth < minLeftPaneWidth {
		leftWidth = minLeftPaneWidth
	}
	rightWidth := usableWidth - leftWidth - 2
	if rightWidth < minRightPaneWidth {
		rightWidth = minRightPaneWidth
	}
	paneHeight := m.height - 8
	if paneHeight < 5 {
		paneHeight = 5
	}
	return leftWidth, rightWidth, paneHeight
}

func (m Model) editPaneLayout() (int, int, int) {
	usableWidth := m.width - 10
	leftWidth := usableWidth / 2
	if leftWidth < 30 {
		leftWidth = 30
	}
	rightWidth := usableWidth - leftWidth - 2
	if rightWidth < 20 {
		rightWidth = 20
	}
	paneHeight := m.height - 11
	if paneHeight < 5 {
		paneHeight = 5
	}
	return leftWidth, rightWidth, paneHeight
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Trigger re-rendering of whatever is currently shown if the width changed
		var cmds []tea.Cmd
		switch m.screen {
		case screenProjectView:
			_, rw, _ := m.projectViewLayout()
			m.todayNoteRendered = "" // Force "Rendering..." while re-rendering
			cmds = append(cmds, m.renderMarkdownCmd(m.todayNote, rw-4, "today"))
		case screenNoteList:
			_, rw, _ := m.projectViewLayout()
			m.previewNoteRendered = ""
			cmds = append(cmds, m.renderMarkdownCmd(m.previewNote, rw-4, "preview"))
		case screenEdit:
			if m.editCategory != storage.CategoryDaily {
				_, rw, _ := m.editPaneLayout()
				m.editRefRendered = ""
				cmds = append(cmds, m.renderMarkdownCmd(m.editRef, rw-4, "edit"))
			}
		}
		return m, tea.Batch(cmds...)

	case projectsLoadedMsg:
		m.projects = msg.projects
		m.err = msg.err
		return m, nil

	case remindersLoadedMsg:
		m.reminders = msg.reminders
		// Clamp cursor if reminders list shrank (e.g. after saving a summary)
		if total := m.totalProjectViewItems(); m.menuCursor >= total && total > 0 {
			m.menuCursor = total - 1
		}
		// Silently ignore errors — reminders are best-effort
		return m, nil

	case refContentLoadedMsg:
		if msg.err != nil {
			m.editRef = "(error loading reference: " + msg.err.Error() + ")"
		} else {
			m.editRef = msg.content
		}
		_, rw, ph := m.editPaneLayout()
		vpWidth := rw - 4
		if vpWidth < 20 {
			vpWidth = 20
		}
		vpHeight := ph - 3
		if vpHeight < 3 {
			vpHeight = 3
		}
		m.editViewport = viewport.New(vpWidth, vpHeight)
		m.editRefRendered = ""
		return m, m.renderMarkdownCmd(m.editRef, vpWidth, "edit")

	case markdownRenderedMsg:
		switch msg.target {
		case "today":
			m.todayNoteRendered = msg.content
		case "preview":
			m.previewNoteRendered = msg.content
		case "edit":
			m.editRefRendered = msg.content
			m.editViewport.SetContent(m.editRefRendered)
		}
		return m, nil

	case noteLoadedMsg:
		if msg.err != nil {
			m.statusMsg = "Error loading note: " + msg.err.Error()
			m.statusErr = true
		} else {
			switch msg.target {
			case "today":
				m.todayNote = msg.content
				_, rw, _ := m.projectViewLayout()
				m.todayNoteRendered = ""
				return m, m.renderMarkdownCmd(m.todayNote, rw-4, "today")
			case "preview":
				m.previewNote = msg.content
				_, rw, _ := m.projectViewLayout()
				m.previewNoteRendered = ""
				return m, m.renderMarkdownCmd(m.previewNote, rw-4, "preview")
			case "edit":
				m.editTextarea.SetValue(msg.content)
				m.editDirty = false
			}
		}
		return m, nil

	case notesListedMsg:
		if msg.err != nil {
			m.statusMsg = "Error listing notes: " + msg.err.Error()
			m.statusErr = true
		} else {
			m.notes = msg.notes
			m.noteCursor = 0
			m.previewNote = ""
			m.previewNoteRendered = ""
			if len(m.notes) > 0 {
				return m, m.loadNoteContent(m.currentProject, m.noteCategory, m.notes[0].Name, "preview")
			}
		}
		return m, nil

	case noteSavedMsg:
		if msg.err != nil {
			m.statusMsg = "Error saving: " + msg.err.Error()
			m.statusErr = true
			return m, nil
		}
		m.statusMsg = "Saved ✓"
		m.statusErr = false
		m.editDirty = false
		// Reload today's note and reminders so the project view shows the latest content
		return m, tea.Batch(m.loadTodayNote(), m.loadReminders())

	case projectCreatedMsg:
		if msg.err != nil {
			m.statusMsg = "Error creating project: " + msg.err.Error()
			m.statusErr = true
		} else {
			m.statusMsg = "Project created ✓"
			m.statusErr = false
			m.creatingNew = false
		}
		return m, m.loadProjects
	}

	// Delegate to the active screen
	switch m.screen {
	case screenProjectList:
		return m.updateProjectList(msg)
	case screenProjectView:
		return m.updateProjectView(msg)
	case screenNoteList:
		return m.updateNoteList(msg)
	case screenEdit:
		return m.updateEdit(msg)
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	var content string

	switch m.screen {
	case screenProjectList:
		content = m.viewProjectList()
	case screenProjectView:
		content = m.viewProjectView()
	case screenNoteList:
		content = m.viewNoteList()
	case screenEdit:
		content = m.viewEdit()
	}

	return appStyle.MaxWidth(m.width).MaxHeight(m.height).Render(content)
}

// --- Screen: Project List ---

func (m Model) updateProjectList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.creatingNew {
		return m.updateCreateProject(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.projectCursor > 0 {
				m.projectCursor--
			}
		case "down", "j":
			if m.projectCursor < len(m.projects)-1 {
				m.projectCursor++
			}
		case "enter":
			if len(m.projects) > 0 {
				m.currentProject = m.projects[m.projectCursor]
				m.screen = screenProjectView
				m.menuCursor = 0
				m.statusMsg = ""
				m.reminders = nil
				return m, tea.Batch(m.loadTodayNote(), m.loadReminders())
			}
		case "n":
			m.creatingNew = true
			m.newNameInput.Reset()
			m.newNameInput.Focus()
			return m, m.newNameInput.Cursor.BlinkCmd()
		}
	}

	return m, nil
}

func (m Model) updateCreateProject(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.creatingNew = false
			return m, nil
		case "enter":
			name := m.newNameInput.Value()
			if name != "" {
				m.creatingNew = false
				return m, m.createProject(name)
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.newNameInput, cmd = m.newNameInput.Update(msg)
	return m, cmd
}

func (m Model) viewProjectList() string {
	var s string
	s += titleStyle.Render("🍵 teatime") + "\n\n"

	if m.err != nil {
		s += errorStyle.Render("Error: "+m.err.Error()) + "\n\n"
	}

	if len(m.projects) == 0 && !m.creatingNew {
		s += mutedStyle.Render("No projects yet. Press [n] to create one.") + "\n"
	}

	for i, p := range m.projects {
		if i == m.projectCursor {
			s += selectedItemStyle.Render("  > "+p) + "\n"
		} else {
			s += normalItemStyle.Render("    "+p) + "\n"
		}
	}

	s += "\n"

	if m.creatingNew {
		s += "Project name: " + m.newNameInput.View() + "\n"
		s += helpBarStyle.Render(helpEntry("enter", "create") + "  " + helpEntry("esc", "cancel"))
	} else {
		if m.statusMsg != "" {
			if m.statusErr {
				s += errorStyle.Render(m.statusMsg) + "\n"
			} else {
				s += successStyle.Render(m.statusMsg) + "\n"
			}
		}
		s += helpBarStyle.Render(
			helpEntry("↑/↓", "navigate") + "  " +
				helpEntry("enter", "select") + "  " +
				helpEntry("n", "new project") + "  " +
				helpEntry("q", "quit"),
		)
	}

	return s
}

// --- Screen: Project View ---

// totalProjectViewItems returns the total number of navigable items
// (reminders + menu items) in the project view.
func (m Model) totalProjectViewItems() int {
	return len(m.reminders) + len(menuItems)
}

func (m Model) updateProjectView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "b", "esc":
			m.screen = screenProjectList
			m.statusMsg = ""
			return m, nil
		case "up", "k":
			if m.menuCursor > 0 {
				m.menuCursor--
			}
		case "down", "j":
			if m.menuCursor < m.totalProjectViewItems()-1 {
				m.menuCursor++
			}
		case "enter":
			return m.handleMenuSelect()
		case "e":
			return m.enterEditMode(storage.CategoryDaily, storage.TodayName())
		case "d":
			return m.enterNoteList(storage.CategoryDaily)
		case "w":
			return m.enterNoteList(storage.CategoryWeekly)
		case "m":
			return m.enterNoteList(storage.CategoryMonthly)
		case "Q":
			return m.enterNoteList(storage.CategoryQuarterly)
		case "y":
			return m.enterNoteList(storage.CategoryYearly)
		}
	}

	return m, nil
}

func (m Model) handleMenuSelect() (tea.Model, tea.Cmd) {
	// Cursor is in the reminders section
	if m.menuCursor < len(m.reminders) {
		r := m.reminders[m.menuCursor]
		return m.enterEditMode(r.Category, r.Name)
	}
	// Cursor is in the menu items section
	menuIdx := m.menuCursor - len(m.reminders)
	if menuIdx >= 0 && menuIdx < len(menuItems) {
		item := menuItems[menuIdx]
		if item.key == "e" {
			return m.enterEditMode(storage.CategoryDaily, storage.TodayName())
		}
		return m.enterNoteList(item.category)
	}
	return m, nil
}

func (m Model) viewProjectView() string {
	leftWidth, rightWidth, paneHeight := m.projectViewLayout()

	// Left pane: menu
	leftContent := headerStyle.Render(m.currentProject) + "\n\n"

	// Reminders (navigable)
	if len(m.reminders) > 0 {
		leftContent += reminderStyle.Render("⚠ Missing summaries:") + "\n"
		for i, r := range m.reminders {
			if i == m.menuCursor {
				leftContent += selectedItemStyle.Render("  > • "+r.Label) + "\n"
			} else {
				leftContent += reminderItemStyle.Render("    • "+r.Label) + "\n"
			}
		}
		leftContent += "\n"
	}

	// Menu items (cursor offset by number of reminders)
	for i, item := range menuItems {
		line := "[" + item.key + "] " + item.label
		idx := i + len(m.reminders)
		if idx == m.menuCursor {
			leftContent += selectedItemStyle.Render("  > "+line) + "\n"
		} else {
			leftContent += normalItemStyle.Render("    "+line) + "\n"
		}
	}

	leftPane := leftPaneStyle.
		Width(leftWidth).
		Height(paneHeight).
		Render(leftContent)

	// Right pane: today's note preview
	rightContent := previewHeaderStyle.Render("📅 "+storage.TodayName()) + "\n\n"
	if m.todayNote == "" {
		rightContent += mutedStyle.Render("No entry for today yet.\nPress [e] to start writing.")
	} else if m.todayNoteRendered == "" {
		rightContent += mutedStyle.Render("Rendering...")
	} else {
		rightContent += m.todayNoteRendered
	}

	rightPane := rightPaneStyle.
		Width(rightWidth).
		Height(paneHeight).
		Render(rightContent)

	// Join panes side by side
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Title
	title := titleStyle.Render("🍵 teatime")

	// Status
	status := ""
	if m.statusMsg != "" {
		if m.statusErr {
			status = errorStyle.Render(m.statusMsg)
		} else {
			status = successStyle.Render(m.statusMsg)
		}
	}

	// Help
	help := helpBarStyle.Render(
		helpEntry("↑/↓", "navigate") + "  " +
			helpEntry("enter", "select") + "  " +
			helpEntry("b", "back") + "  " +
			helpEntry("q", "quit"),
	)

	return lipgloss.JoinVertical(lipgloss.Left, title, body, status, help)
}

// --- Screen: Note List ---

func (m Model) enterNoteList(category storage.Category) (tea.Model, tea.Cmd) {
	m.screen = screenNoteList
	m.noteCategory = category
	m.noteCursor = 0
	m.previewNote = ""
	m.statusMsg = ""
	return m, m.listNotes(m.currentProject, category)
}

func (m Model) updateNoteList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "b", "esc":
			m.screen = screenProjectView
			m.statusMsg = ""
			return m, nil
		case "up", "k":
			if m.noteCursor > 0 {
				m.noteCursor--
				if len(m.notes) > 0 {
					return m, m.loadNoteContent(m.currentProject, m.noteCategory, m.notes[m.noteCursor].Name, "preview")
				}
			}
		case "down", "j":
			if m.noteCursor < len(m.notes)-1 {
				m.noteCursor++
				return m, m.loadNoteContent(m.currentProject, m.noteCategory, m.notes[m.noteCursor].Name, "preview")
			}
		case "enter", "e":
			if len(m.notes) > 0 {
				note := m.notes[m.noteCursor]
				return m.enterEditMode(m.noteCategory, note.Name)
			}
		case "n":
			name := storage.DefaultNameForCategory(m.noteCategory)
			return m.enterEditMode(m.noteCategory, name)
		}
	}

	return m, nil
}

func (m Model) viewNoteList() string {
	leftWidth, rightWidth, paneHeight := m.projectViewLayout()

	// Left pane: note list
	leftContent := headerStyle.Render(storage.CategoryLabel(m.noteCategory)) + "\n\n"
	if len(m.notes) == 0 {
		leftContent += mutedStyle.Render("No notes yet.\nPress [n] to create one.")
	} else {
		for i, note := range m.notes {
			if i == m.noteCursor {
				leftContent += selectedItemStyle.Render("  > "+note.Name) + "\n"
			} else {
				leftContent += normalItemStyle.Render("    "+note.Name) + "\n"
			}
		}
	}

	leftPane := leftPaneStyle.
		Width(leftWidth).
		Height(paneHeight).
		Render(leftContent)

	// Right pane: preview
	rightContent := ""
	if len(m.notes) > 0 && m.noteCursor < len(m.notes) {
		rightContent += previewHeaderStyle.Render("📄 "+m.notes[m.noteCursor].Name) + "\n\n"
		if m.previewNote == "" {
			rightContent += mutedStyle.Render("(empty)")
		} else if m.previewNoteRendered == "" {
			rightContent += mutedStyle.Render("Rendering...")
		} else {
			rightContent += m.previewNoteRendered
		}
	} else {
		rightContent += mutedStyle.Render("Select a note to preview")
	}

	rightPane := rightPaneStyle.
		Width(rightWidth).
		Height(paneHeight).
		Render(rightContent)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	title := titleStyle.Render("🍵 teatime — " + m.currentProject)

	status := ""
	if m.statusMsg != "" {
		if m.statusErr {
			status = errorStyle.Render(m.statusMsg)
		} else {
			status = successStyle.Render(m.statusMsg)
		}
	}

	help := helpBarStyle.Render(
		helpEntry("↑/↓", "navigate") + "  " +
			helpEntry("enter", "edit") + "  " +
			helpEntry("n", "new note") + "  " +
			helpEntry("b", "back") + "  " +
			helpEntry("q", "quit"),
	)

	return lipgloss.JoinVertical(lipgloss.Left, title, body, status, help)
}

// --- Screen: Edit ---

func (m Model) enterEditMode(category storage.Category, name string) (tea.Model, tea.Cmd) {
	m.screen = screenEdit
	m.editCategory = category
	m.editNoteName = name
	m.editDirty = false
	m.editRef = ""
	m.editFocusLeft = true
	m.statusMsg = ""

	hasSplitPane := category != storage.CategoryDaily

	// Height budget for the edit screen (split-pane mode):
	//
	// lipgloss .Width(w)/.Height(h) set the size INCLUDING padding but
	// EXCLUDING borders. So for a pane with Padding(1,2) and RoundedBorder:
	//   content width  = w - leftPad(2) - rightPad(2) = w - 4
	//   content height = h - topPad(1) - bottomPad(1) = h - 2
	//   outer width    = w + leftBorder(1) + rightBorder(1) = w + 2
	//   outer height   = h + topBorder(1) + bottomBorder(1) = h + 2
	//
	// Vertical layout of the edit screen (inside appStyle padding):
	//   title + marginBottom(1)      = 2 lines
	//   subtitle                     = 1 line
	//   blank ""                     = 1 line
	//   pane outer (paneH + 2)       = paneH + 2 lines
	//   status ""                    = 1 line
	//   help marginTop(1) + text     = 2 lines
	//                          total = paneH + 9
	//
	// appStyle Padding(1,2) → vertical padding = 2; available = m.height - 2
	// paneH + 9 = m.height - 2  →  paneH = m.height - 11
	//
	// Inside the pane content area (paneH - 2 after vertical padding):
	//   pane header label            = 1 line
	//   textarea                     = taHeight lines
	//   taHeight = paneH - 2 - 1    = paneH - 3

	// Size the textarea
	if hasSplitPane {
		// Split layout: textarea on the left
		leftWidth, _, _ := m.editPaneLayout()
		// Subtract horizontal padding (2+2) so textarea fits inside pane content area
		taWidth := leftWidth - 4
		if taWidth < 20 {
			taWidth = 20
		}
		m.editTextarea.SetWidth(taWidth)
	} else {
		// Full width for daily notes
		taWidth := m.width - 10
		if taWidth < 40 {
			taWidth = 40
		}
		m.editTextarea.SetWidth(taWidth)
	}

	var taHeight int
	if hasSplitPane {
		_, _, ph := m.editPaneLayout()
		taHeight = ph - 3 // subtract vertical padding (2) + header (1)
	} else {
		taHeight = m.height - 8 // no pane border/padding in full-width mode
	}
	if taHeight < 5 {
		taHeight = 5
	}
	m.editTextarea.SetHeight(taHeight)
	m.editTextarea.Focus()

	cmds := []tea.Cmd{
		m.loadNoteContent(m.currentProject, category, name, "edit"),
		m.editTextarea.Cursor.BlinkCmd(),
	}

	// Load reference content for summary notes
	if hasSplitPane {
		cmds = append(cmds, m.loadReferenceContent(m.currentProject, category, name))
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	hasSplitPane := m.editCategory != storage.CategoryDaily

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Save and go back; loadTodayNote will be triggered by noteSavedMsg
			content := m.editTextarea.Value()
			m.screen = screenProjectView
			return m, m.saveNote(m.currentProject, m.editCategory, m.editNoteName, content)
		case "ctrl+c":
			// Abort without saving
			m.screen = screenProjectView
			m.statusMsg = "Edit cancelled"
			m.statusErr = false
			return m, nil
		case "tab":
			if hasSplitPane {
				m.editFocusLeft = !m.editFocusLeft
				if m.editFocusLeft {
					m.editTextarea.Focus()
				} else {
					m.editTextarea.Blur()
				}
				return m, nil
			}
		}
	}

	if m.editFocusLeft {
		var cmd tea.Cmd
		m.editTextarea, cmd = m.editTextarea.Update(msg)
		m.editDirty = true
		return m, cmd
	}

	// Right pane (viewport) is focused — forward scroll events
	var cmd tea.Cmd
	m.editViewport, cmd = m.editViewport.Update(msg)
	return m, cmd
}

func (m Model) viewEdit() string {
	hasSplitPane := m.editCategory != storage.CategoryDaily

	title := titleStyle.Render("🍵 teatime — " + m.currentProject + " — " + m.editNoteName + " [edit]")

	catLabel := storage.CategoryLabel(m.editCategory)
	subtitle := mutedStyle.Render(catLabel)

	var body string
	if hasSplitPane && m.editRef != "" {
		// Split layout: editor on left, reference on right
		leftWidth, rightWidth, paneHeight := m.editPaneLayout()

		// Left pane: editor
		var leftPane string
		leftLabel := paneHeaderStyle.Render("✏️  Editor")
		leftContent := leftLabel + "\n" + m.editTextarea.View()
		if m.editFocusLeft {
			leftPane = focusedBorderStyle.
				Width(leftWidth).
				Height(paneHeight).
				Render(leftContent)
		} else {
			leftPane = leftPaneStyle.
				Width(leftWidth).
				Height(paneHeight).
				Render(leftContent)
		}

		// Right pane: reference content (scrollable viewport)
		refLabel := paneHeaderStyle.Render(referenceLabel(m.editCategory))
		rightContent := refLabel + "\n" + m.editViewport.View()
		var rightPane string
		if !m.editFocusLeft {
			rightPane = focusedBorderStyle.
				Width(rightWidth).
				Height(paneHeight).
				Render(rightContent)
		} else {
			rightPane = rightPaneStyle.
				Width(rightWidth).
				Height(paneHeight).
				Render(rightContent)
		}

		body = lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	} else {
		// Full width for daily notes
		body = m.editTextarea.View()
	}

	dirtyMarker := ""
	if m.editDirty {
		dirtyMarker = " (unsaved)"
	}

	status := ""
	if m.statusMsg != "" {
		if m.statusErr {
			status = errorStyle.Render(m.statusMsg)
		} else {
			status = successStyle.Render(m.statusMsg)
		}
	}

	maxHelpWidth := m.width - 4 // account for appStyle horizontal padding
	if maxHelpWidth < 20 {
		maxHelpWidth = 20
	}

	var help string
	if hasSplitPane {
		focusHint := "ref"
		if !m.editFocusLeft {
			focusHint = "editor"
		}
		help = helpBarStyle.MaxWidth(maxHelpWidth).Render(
			helpEntry("tab", focusHint) + "  " +
				helpEntry("esc", "save"+dirtyMarker) + "  " +
				helpEntry("ctrl+c", "discard"),
		)
	} else {
		help = helpBarStyle.MaxWidth(maxHelpWidth).Render(
			helpEntry("esc", "save"+dirtyMarker) + "  " +
				helpEntry("ctrl+c", "discard"),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle, "", body, status, help)
}

// referenceLabel returns a label for the reference pane based on category.
func referenceLabel(cat storage.Category) string {
	switch cat {
	case storage.CategoryWeekly:
		return "📋 Daily entries"
	case storage.CategoryMonthly:
		return "📋 Weekly summaries"
	case storage.CategoryQuarterly:
		return "📋 Monthly summaries"
	case storage.CategoryYearly:
		return "📋 Quarterly summaries"
	default:
		return "📋 Reference"
	}
}

// --- Commands (async operations) ---

type projectsLoadedMsg struct {
	projects []string
	err      error
}

type noteLoadedMsg struct {
	content string
	target  string // "today", "preview", or "edit"
	err     error
}

type refContentLoadedMsg struct {
	content string
	err     error
}

type markdownRenderedMsg struct {
	content string
	target  string // "today", "preview", or "edit"
}

type notesListedMsg struct {
	notes []storage.NoteFile
	err   error
}

func (m Model) renderMarkdownCmd(content string, width int, target string) tea.Cmd {
	return func() tea.Msg {
		rendered := renderMarkdown(width, content)
		return markdownRenderedMsg{content: rendered, target: target}
	}
}

type noteSavedMsg struct {
	err error
}

type projectCreatedMsg struct {
	err error
}

func (m Model) loadProjects() tea.Msg {
	projects, err := m.store.ListProjects()
	return projectsLoadedMsg{projects: projects, err: err}
}

func (m Model) loadTodayNote() tea.Cmd {
	return func() tea.Msg {
		content, err := m.store.ReadNote(m.currentProject, storage.CategoryDaily, storage.TodayName())
		return noteLoadedMsg{content: content, target: "today", err: err}
	}
}

func (m Model) loadNoteContent(project string, category storage.Category, name string, target string) tea.Cmd {
	return func() tea.Msg {
		content, err := m.store.ReadNote(project, category, name)
		return noteLoadedMsg{content: content, target: target, err: err}
	}
}

func (m Model) loadReferenceContent(project string, category storage.Category, name string) tea.Cmd {
	return func() tea.Msg {
		content, err := m.store.GatherReferenceContent(project, category, name)
		return refContentLoadedMsg{content: content, err: err}
	}
}

func (m Model) listNotes(project string, category storage.Category) tea.Cmd {
	return func() tea.Msg {
		notes, err := m.store.ListNotes(project, category)
		return notesListedMsg{notes: notes, err: err}
	}
}

func (m Model) saveNote(project string, category storage.Category, name string, content string) tea.Cmd {
	return func() tea.Msg {
		err := m.store.WriteNote(project, category, name, content)
		return noteSavedMsg{err: err}
	}
}

func (m Model) createProject(name string) tea.Cmd {
	return func() tea.Msg {
		err := m.store.CreateProject(name)
		return projectCreatedMsg{err: err}
	}
}

type remindersLoadedMsg struct {
	reminders []storage.Reminder
	err       error
}

func (m Model) loadReminders() tea.Cmd {
	return func() tea.Msg {
		reminders, err := m.store.CheckMissingSummaries(m.currentProject)
		return remindersLoadedMsg{reminders: reminders, err: err}
	}
}
