package ui

import (
	"context"
	"fmt"
	"strings"

	"f6n/internal/provider"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputMode represents the current input mode
type InputMode int

const (
	NormalMode InputMode = iota
	FilterMode
	CommandMode
)

// Model represents the application state
type Model struct {
	table         table.Model
	viewport      viewport.Model
	textInput     textinput.Model
	functions     []provider.FunctionInfo
	allFunctions  []provider.FunctionInfo // Unfiltered list
	provider      provider.Provider
	currentView   ViewType
	selectedFunc  *provider.FunctionInfo
	environment   string
	inputMode     InputMode
	filterActive  bool   // Whether a filter is currently applied
	activeFilter  string // The current filter text
	width         int
	height        int
	loading       bool
	err           error
}

// functionsLoadedMsg is sent when functions are loaded from the provider
type functionsLoadedMsg struct {
	functions []provider.FunctionInfo
	err       error
}

// NewModel creates a new TUI model
func NewModel(prov provider.Provider, environment string) Model {
	columns := []table.Column{
		{Title: "Function Name", Width: 40},
		{Title: "Runtime", Width: 15},
		{Title: "Memory", Width: 10},
		{Title: "Timeout", Width: 10},
		{Title: "Last Modified", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#07646bff")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#07646bff")).
		Bold(true)
	t.SetStyles(s)

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#07646bff")).
		Padding(1, 2)

	// Initialize text input for filter/command mode
	ti := textinput.New()
	ti.Placeholder = "Filter functions..."
	ti.CharLimit = 100
	ti.Width = 50

	return Model{
		table:        t,
		viewport:     vp,
		textInput:    ti,
		provider:     prov,
		currentView:  ListView,
		environment:  environment,
		inputMode:    NormalMode,
		loading:      true,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	if m.err != nil {
		return tea.Quit
	}
	return tea.Batch(
		m.fetchFunctions(),
		tea.EnterAltScreen,
	)
}

// fetchFunctions loads functions from the cloud provider
func (m Model) fetchFunctions() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		functions, err := m.provider.ListFunctions(ctx)
		if err != nil {
			return functionsLoadedMsg{err: err}
		}
		return functionsLoadedMsg{functions: functions}
	}
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case functionsLoadedMsg:
		return m.handleFunctionsLoaded(msg)

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	var cmd tea.Cmd
	if m.currentView == ListView {
		m.table, cmd = m.table.Update(msg)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

// handleWindowSize handles window resize events
func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	// Calculate available height for table (subtract top padding, header, info rows, shortcuts, help)
	// Top padding: 5, ASCII art: 6, Info: 3, Shortcuts: 3, Help: 2, Extra spacing: 3 = 22 total
	availableHeight := msg.Height - 22
	if availableHeight < 5 {
		availableHeight = 5
	}

	m.table.SetHeight(availableHeight)

	// Update table column widths to span entire width
	totalWidth := msg.Width - 4
	m.table.SetColumns([]table.Column{
		{Title: "Function Name", Width: int(float64(totalWidth) * 0.35)},
		{Title: "Runtime", Width: int(float64(totalWidth) * 0.15)},
		{Title: "Memory", Width: int(float64(totalWidth) * 0.12)},
		{Title: "Timeout", Width: int(float64(totalWidth) * 0.12)},
		{Title: "Last Modified", Width: int(float64(totalWidth) * 0.26)},
	})

	m.viewport.Width = msg.Width - 4
	m.viewport.Height = msg.Height - 8
	return m, nil
}

// handleFunctionsLoaded handles the functions loaded message
func (m Model) handleFunctionsLoaded(msg functionsLoadedMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.err = msg.err
		return m, nil
	}

	// Store both filtered and unfiltered lists
	m.allFunctions = msg.functions
	m.functions = msg.functions
	m.updateTable()
	return m, nil
}

// updateTable updates the table with current functions list
func (m *Model) updateTable() {
	rows := []table.Row{}
	for _, fn := range m.functions {
		rows = append(rows, table.Row{
			fn.Name,
			fn.Runtime,
			fmt.Sprintf("%d MB", fn.Memory),
			fmt.Sprintf("%d s", fn.Timeout),
			fn.LastModified,
		})
	}
	m.table.SetRows(rows)
}

// filterFunctions filters functions based on the current filter text
func (m *Model) filterFunctions() {
	filterText := strings.ToLower(strings.TrimSpace(m.textInput.Value()))
	if filterText == "" {
		m.functions = m.allFunctions
	} else {
		m.functions = []provider.FunctionInfo{}
		for _, fn := range m.allFunctions {
			// Filter by function name, runtime, or description
			if strings.Contains(strings.ToLower(fn.Name), filterText) ||
				strings.Contains(strings.ToLower(fn.Runtime), filterText) ||
				strings.Contains(strings.ToLower(fn.Description), filterText) {
				m.functions = append(m.functions, fn)
			}
		}
	}
	m.updateTable()
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle input modes
	if m.inputMode == FilterMode || m.inputMode == CommandMode {
		return m.handleInputMode(msg)
	}

	// Normal mode key handling
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
		
	case "q":
		if m.currentView == ListView {
			return m, tea.Quit
		}
		return m, nil

	case "\\":
		// Enter filter mode
		if m.currentView == ListView {
			m.inputMode = FilterMode
			m.textInput.Placeholder = "Filter functions..."
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, textinput.Blink
		}
		return m, nil

	case ":":
		// Enter command mode
		m.inputMode = CommandMode
		m.textInput.Placeholder = "Enter command (:q to quit)..."
		m.textInput.SetValue(":")
		m.textInput.Focus()
		m.textInput.CursorEnd()
		return m, textinput.Blink

	case "enter":
		if m.currentView == ListView && len(m.functions) > 0 {
			selectedIdx := m.table.Cursor()
			if selectedIdx < len(m.functions) {
				m.selectedFunc = &m.functions[selectedIdx]
				m.currentView = DetailView
				m.viewport.SetContent(formatFunctionDetails(m.selectedFunc))
			}
		}
		return m, nil

	case "esc":
		if m.currentView != ListView {
			m.currentView = ListView
		} else if m.filterActive {
			// Clear active filter when in list view
			m.filterActive = false
			m.activeFilter = ""
			m.functions = m.allFunctions
			m.updateTable()
		}
		return m, nil

	case "r":
		if m.currentView == ListView {
			m.loading = true
			return m, m.fetchFunctions()
		}
		return m, nil
	}

	var cmd tea.Cmd
	if m.currentView == ListView {
		m.table, cmd = m.table.Update(msg)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

// handleInputMode handles keys when in filter or command mode
func (m Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyEsc:
		currentMode := m.inputMode
		// Exit input mode
		m.inputMode = NormalMode
		m.textInput.Blur()
		if currentMode == FilterMode {
			// Reset filter when escaping from filter mode
			m.filterActive = false
			m.activeFilter = ""
			m.functions = m.allFunctions
			m.updateTable()
		}
		return m, nil

	case tea.KeyEnter:
		if m.inputMode == FilterMode {
			// Apply filter and exit filter mode
			filterText := strings.TrimSpace(m.textInput.Value())
			if filterText != "" {
				m.filterActive = true
				m.activeFilter = filterText
			} else {
				m.filterActive = false
				m.activeFilter = ""
			}
			m.inputMode = NormalMode
			m.textInput.Blur()
			return m, nil
		} else if m.inputMode == CommandMode {
			// Execute command
			command := strings.TrimSpace(m.textInput.Value())
			m.inputMode = NormalMode
			m.textInput.Blur()
			return m.executeCommand(command)
		}

	case tea.KeyCtrlC:
		return m, tea.Quit
	}

	// Update text input
	m.textInput, cmd = m.textInput.Update(msg)
	
	// If in filter mode, update filter in real-time
	if m.inputMode == FilterMode {
		m.filterFunctions()
	}
	
	return m, cmd
}

// executeCommand executes a vim-like command
func (m Model) executeCommand(command string) (tea.Model, tea.Cmd) {
	switch command {
	case ":q", ":quit":
		return m, tea.Quit
	case ":r", ":refresh":
		m.loading = true
		return m, m.fetchFunctions()
	default:
		// Unknown command, just ignore
		return m, nil
	}
}

// View renders the UI
func (m Model) View() string {
	return renderView(m)
}
