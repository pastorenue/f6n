package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"f6n/internal/charts"
	"f6n/internal/logger"
	"f6n/internal/provider"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
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
	table           table.Model
	viewport        viewport.Model
	textInput       textinput.Model
	textarea        textarea.Model
	functions       []provider.FunctionInfo
	allFunctions    []provider.FunctionInfo // Unfiltered list
	provider        provider.Provider
	accountID       string
	currentView     ViewType
	selectedFunc    *provider.FunctionInfo
	environment     string
	inputMode       InputMode
	editMode        bool   // Whether CodeView is in edit mode
	originalContent string // Store original content for cancel
	filterActive    bool   // Whether a filter is currently applied
	activeFilter    string // The current filter text
	width           int
	height          int
	loading         bool
	err             error
	// Log streaming fields
	streamingLogs bool               // Whether we're currently streaming logs
	streamCancel  context.CancelFunc // Function to cancel log streaming
	realTimeLogs  []string           // Buffer for real-time logs
	logStreamErr  error              // Error from log streaming
}

type functionsLoadedMsg struct {
	functions []provider.FunctionInfo
	err       error
}

type accountIDLoadedMsg struct {
	accountID string
	err       error
}

func (m Model) fetchAccountID() tea.Cmd {
	return func() tea.Msg {
		accountID, err := m.provider.GetAccountID(context.Background())
		if err != nil {
			return accountIDLoadedMsg{err: err}
		}
		return accountIDLoadedMsg{accountID: accountID}
	}
}

type functionCodeLoadedMsg struct {
	code string
	err  error
}

type functionLogsLoadedMsg struct {
	logs []string
	err  error
}

type logStreamStartedMsg struct {
	functionName string
}

type newLogEntryMsg struct {
	entry provider.LogEntry
}

type logStreamErrorMsg struct {
	err error
}

type functionMetricsLoadedMsg struct {
	metrics *provider.FunctionMetrics
	err     error
}

type functionCodeDownloadedMsg struct {
	path string
	err  error
}

type downloadingMsg struct {
	functionName string
}

type codeFilesLoadedMsg struct {
	content string
	err     error
}

type loadingCodeFilesMsg struct {
	functionName string
}

type editSavedMsg struct {
	success bool
	err     error
}

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

func (m Model) fetchFunctionCode(name string) tea.Cmd {
	logger.Logger.Printf("Fetching function code for: %s", name)
	return func() tea.Msg {
		code, err := m.provider.GetFunctionCode(context.Background(), name)
		if err != nil {
			logger.Logger.Printf("Error fetching function code: %v", err)
			return functionCodeLoadedMsg{err: err}
		}
		logger.Logger.Printf("Function code loaded successfully")
		return functionCodeLoadedMsg{code: code}
	}
}

func (m Model) fetchFunctionLogs(name string) tea.Cmd {
	return func() tea.Msg {
		logs, err := m.provider.GetFunctionLogs(context.Background(), name, 200)
		if err != nil {
			println("Error fetching function logs:", err.Error())
			return functionLogsLoadedMsg{err: err}
		}
		return functionLogsLoadedMsg{logs: logs}
	}
}

func (m Model) startLogStreaming(name string) tea.Cmd {
	return func() tea.Msg {
		return logStreamStartedMsg{functionName: name}
	}
}

func (m Model) streamLogs(ctx context.Context, name string) tea.Cmd {
	return func() tea.Msg {
		logChan, errChan := m.provider.StreamFunctionLogs(ctx, name)

		// Listen for logs or errors
		select {
		case entry, ok := <-logChan:
			if !ok {
				// Channel closed, streaming ended
				return logStreamErrorMsg{err: fmt.Errorf("log stream ended")}
			}
			return newLogEntryMsg{entry: entry}
		case err, ok := <-errChan:
			if !ok {
				// Error channel closed
				return logStreamErrorMsg{err: fmt.Errorf("error stream ended")}
			}
			return logStreamErrorMsg{err: err}
		case <-ctx.Done():
			return logStreamErrorMsg{err: ctx.Err()}
		}
	}
}

func (m Model) fetchFunctionMetrics(name string) tea.Cmd {
	return func() tea.Msg {
		// Get metrics for the last hour
		endTime := time.Now()
		startTime := endTime.Add(-1 * time.Hour)

		metrics, err := m.provider.GetFunctionMetrics(context.Background(), name, startTime, endTime)
		if err != nil {
			logger.Logger.Printf("Error fetching metrics for %s: %v", name, err)
			return functionMetricsLoadedMsg{err: err}
		}
		return functionMetricsLoadedMsg{metrics: metrics}
	}
}

func (m Model) downloadFunctionCode(name string) tea.Cmd {
	logger.Logger.Printf("Starting download for function: %s", name)
	return func() tea.Msg {
		// Create downloads base directory if it doesn't exist
		baseDir := "downloads"
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			logger.Logger.Printf("Error creating downloads directory: %v", err)
			return functionCodeDownloadedMsg{err: fmt.Errorf("failed to create downloads directory: %w", err)}
		}

		// Create function-specific download directory
		downloadPath := filepath.Join(baseDir, name)

		// Check if directory already exists and warn user
		if _, err := os.Stat(downloadPath); err == nil {
			logger.Logger.Printf("Download directory already exists, will overwrite: %s", downloadPath)
		}

		err := m.provider.DownloadFunctionCode(context.Background(), name, downloadPath)
		if err != nil {
			logger.Logger.Printf("Error downloading function code: %v", err)
			return functionCodeDownloadedMsg{err: fmt.Errorf("download failed: %w", err)}
		}

		// Get absolute path for display
		absPath, _ := filepath.Abs(downloadPath)

		logger.Logger.Printf("Function code downloaded successfully to: %s", absPath)
		return functionCodeDownloadedMsg{path: absPath}
	}
}

func (m Model) loadCodeFiles(functionName string) tea.Cmd {
	logger.Logger.Printf("Loading code files for function: %s", functionName)
	return func() tea.Msg {
		downloadPath := filepath.Join("downloads", functionName)

		// Check if download directory exists
		if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
			return codeFilesLoadedMsg{err: fmt.Errorf("code not downloaded yet. Press 'd' first to download the code")}
		}

		content, err := m.readCodeFiles(downloadPath)
		if err != nil {
			logger.Logger.Printf("Error reading code files: %v", err)
			return codeFilesLoadedMsg{err: fmt.Errorf("failed to read code files: %w", err)}
		}

		logger.Logger.Printf("Code files loaded successfully")
		return codeFilesLoadedMsg{content: content}
	}
}

func (m Model) readCodeFiles(dirPath string) (string, error) {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("📁 Code Files for %s\n", filepath.Base(dirPath)))
	content.WriteString("═══════════════════════════════════════\n\n")

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-readable files
		if info.IsDir() {
			return nil
		}

		// Only show common code file extensions
		ext := strings.ToLower(filepath.Ext(path))
		if !isCodeFile(ext) {
			return nil
		}

		relPath, _ := filepath.Rel(dirPath, path)
		content.WriteString(fmt.Sprintf("📄 %s\n", relPath))
		content.WriteString("─────────────────────────────────────\n")

		// Read file content (limit size for display)
		fileContent, err := os.ReadFile(path)
		if err != nil {
			content.WriteString(fmt.Sprintf("Error reading file: %v\n\n", err))
			return nil
		}

		// Limit file size for display (100KB max)
		if len(fileContent) > 100*1024 {
			content.WriteString(fmt.Sprintf("File too large (%d bytes). Showing first 100KB...\n\n", len(fileContent)))
			fileContent = fileContent[:100*1024]
		}

		content.WriteString(string(fileContent))
		content.WriteString("\n\n")

		return nil
	})

	if err != nil {
		return "", err
	}

	if content.Len() == 0 {
		content.WriteString("No code files found in the downloaded directory.\n")
		content.WriteString("The download may contain only configuration files or archives.")
	}

	return content.String(), nil
}

func isCodeFile(ext string) bool {
	codeExtensions := map[string]bool{
		".js":    true,
		".py":    true,
		".go":    true,
		".java":  true,
		".php":   true,
		".rb":    true,
		".cs":    true,
		".cpp":   true,
		".c":     true,
		".h":     true,
		".hpp":   true,
		".ts":    true,
		".jsx":   true,
		".tsx":   true,
		".html":  true,
		".css":   true,
		".scss":  true,
		".less":  true,
		".json":  true,
		".yaml":  true,
		".yml":   true,
		".xml":   true,
		".md":    true,
		".txt":   true,
		".sh":    true,
		".bat":   true,
		".ps1":   true,
		".sql":   true,
		".r":     true,
		".scala": true,
		".kt":    true,
		".swift": true,
		".dart":  true,
		".rs":    true,
		".lua":   true,
		".pl":    true,
		".pm":    true,
	}
	return codeExtensions[ext]
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

	// Initialize textarea for code editing
	ta := textarea.New()
	ta.Placeholder = "Enter code here..."
	ta.ShowLineNumbers = true
	ta.SetWidth(80)
	ta.SetHeight(20)

	return Model{
		table:       t,
		viewport:    vp,
		textInput:   ti,
		textarea:    ta,
		provider:    prov,
		currentView: ListView,
		environment: environment,
		inputMode:   NormalMode,
		editMode:    false,
		loading:     true,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	if m.err != nil {
		return tea.Quit
	}
	return tea.Batch(
		m.fetchFunctions(),
		m.fetchAccountID(),
		tea.EnterAltScreen,
	)
}

// fetchFunctions loads functions from the cloud provider

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case accountIDLoadedMsg:
		if msg.err == nil {
			m.accountID = msg.accountID
		}
		return m, nil

	case functionsLoadedMsg:
		return m.handleFunctionsLoaded(msg)

	case functionLogsLoadedMsg:
		if msg.err != nil {
			m.viewport.SetContent(fmt.Sprintf("Error: %v", msg.err))
		} else {
			m.viewport.SetContent(strings.Join(msg.logs, "\n"))
		}
		return m, nil

	case functionMetricsLoadedMsg:
		if msg.err != nil {
			m.viewport.SetContent(fmt.Sprintf("Error loading metrics: %v", msg.err))
		} else {
			// Import charts and render metrics overview
			content := renderMetricsContent(msg.metrics, m.width)
			m.viewport.SetContent(content)
		}
		return m, nil

	case logStreamStartedMsg:
		// Start streaming logs for the function
		m.streamingLogs = true
		m.realTimeLogs = []string{fmt.Sprintf("🔴 Streaming logs for %s (real-time) - Press 's' to stop", msg.functionName)}
		m.logStreamErr = nil

		// Create a cancellable context for the stream
		ctx, cancel := context.WithCancel(context.Background())
		m.streamCancel = cancel

		// Set initial content and start streaming
		m.viewport.SetContent(strings.Join(m.realTimeLogs, "\n"))
		return m, m.streamLogs(ctx, msg.functionName)

	case newLogEntryMsg:
		if m.streamingLogs {
			// Format the log entry
			timestamp := msg.entry.Timestamp.Format("2006-01-02 15:04:05")
			logLine := fmt.Sprintf("[%s] %s: %s", timestamp, msg.entry.Severity, msg.entry.Message)

			// Add to real-time logs buffer (keep last 1000 entries)
			m.realTimeLogs = append(m.realTimeLogs, logLine)
			if len(m.realTimeLogs) > 1000 {
				m.realTimeLogs = m.realTimeLogs[1:]
			}

			// Update viewport content
			m.viewport.SetContent(strings.Join(m.realTimeLogs, "\n"))

			// Continue streaming
			ctx, cancel := context.WithCancel(context.Background())
			if m.streamCancel != nil {
				m.streamCancel()
			}
			m.streamCancel = cancel

			return m, m.streamLogs(ctx, m.selectedFunc.Name)
		}
		return m, nil

	case logStreamErrorMsg:
		if m.streamingLogs {
			m.logStreamErr = msg.err
			m.streamingLogs = false
			if m.streamCancel != nil {
				m.streamCancel()
				m.streamCancel = nil
			}

			// Add error message to logs
			errorLine := fmt.Sprintf("❌ Stream error: %v", msg.err)
			m.realTimeLogs = append(m.realTimeLogs, errorLine)
			m.viewport.SetContent(strings.Join(m.realTimeLogs, "\n"))
		}
		return m, nil

	case functionCodeLoadedMsg:
		if msg.err != nil {
			m.viewport.SetContent(fmt.Sprintf("Error: %v", msg.err))
		} else {
			m.viewport.SetContent(msg.code)
		}
		return m, nil

	case downloadingMsg:
		logger.Logger.Printf("Received downloadingMsg for function: %s", msg.functionName)
		m.viewport.SetContent(fmt.Sprintf("Downloading code for %s...\n\nThis may take a few moments.", msg.functionName))
		return m, nil

	case functionCodeDownloadedMsg:
		logger.Logger.Printf("Received functionCodeDownloadedMsg - success: %t", msg.err == nil)
		if msg.err != nil {
			logger.Logger.Printf("Download error: %v", msg.err)
			m.viewport.SetContent(fmt.Sprintf("Download failed: %v\n\nPress 'esc' to go back.", msg.err))
		} else {
			logger.Logger.Printf("Download successful to path: %s", msg.path)
			content := fmt.Sprintf("✅ Code downloaded successfully!\n\nLocation: %s\n\n", msg.path)
			content += "The function code has been downloaded to your local machine.\n"
			content += "You can now explore the source files in the specified directory.\n\n"
			content += "Press 'esc' to go back to the function list."
			m.viewport.SetContent(content)
		}
		return m, nil

	case loadingCodeFilesMsg:
		m.viewport.SetContent(fmt.Sprintf("Loading code files for %s...\n\nReading downloaded files...", msg.functionName))
		return m, nil

	case codeFilesLoadedMsg:
		if msg.err != nil {
			m.viewport.SetContent(fmt.Sprintf("Error loading code files: %v\n\nPress 'esc' to go back.", msg.err))
		} else {
			m.viewport.SetContent(msg.content)
		}
		return m, nil

	case editSavedMsg:
		if msg.success {
			// Show a temporary save confirmation
			currentContent := m.viewport.View()
			saveConfirmation := "✅ Changes saved!\n\n" + currentContent
			m.viewport.SetContent(saveConfirmation)
		} else if msg.err != nil {
			errorMsg := fmt.Sprintf("❌ Save failed: %v\n\nPress 'esc' to go back.", msg.err)
			m.viewport.SetContent(errorMsg)
		}
		return m, nil

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

	// Update textarea size for edit mode
	m.textarea.SetWidth(msg.Width - 4)
	m.textarea.SetHeight(msg.Height - 10)

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
	logger.Logger.Printf("Key pressed: %s", msg.String())
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
		// Clean up streaming when leaving LogsView
		if m.currentView == LogsView && m.streamingLogs {
			if m.streamCancel != nil {
				m.streamCancel()
				m.streamCancel = nil
			}
			m.streamingLogs = false
		}

		if m.currentView == CodeDisplayView {
			// Go back to CodeView from CodeDisplayView
			m.currentView = CodeView
		} else if m.currentView != ListView {
			m.currentView = ListView
		} else if m.filterActive {
			// Clear active filter when in list view
			m.filterActive = false
			m.activeFilter = ""
			m.functions = m.allFunctions
			m.updateTable()
		}
		return m, nil

	case "l":
		if m.currentView == ListView && len(m.functions) > 0 {
			selectedIdx := m.table.Cursor()
			if selectedIdx < len(m.functions) {
				m.selectedFunc = &m.functions[selectedIdx]
				m.currentView = LogsView
				m.viewport.SetContent("Loading logs...")
				return m, m.fetchFunctionLogs(m.selectedFunc.Name)
			}
		} else if m.currentView == LogsView && m.selectedFunc != nil {
			// In LogsView, 'l' refreshes static logs (stops streaming if active)
			if m.streamingLogs && m.streamCancel != nil {
				m.streamCancel()
				m.streamingLogs = false
				m.streamCancel = nil
			}
			m.viewport.SetContent("Loading logs...")
			return m, m.fetchFunctionLogs(m.selectedFunc.Name)
		}
		return m, nil

	case "s":
		if m.currentView == LogsView && m.selectedFunc != nil {
			if m.streamingLogs {
				// Stop streaming
				if m.streamCancel != nil {
					m.streamCancel()
					m.streamCancel = nil
				}
				m.streamingLogs = false

				// Add stopped message to logs
				stoppedLine := "⏹️  Log streaming stopped"
				m.realTimeLogs = append(m.realTimeLogs, stoppedLine)
				m.viewport.SetContent(strings.Join(m.realTimeLogs, "\n"))
			} else {
				// Start streaming
				return m, m.startLogStreaming(m.selectedFunc.Name)
			}
		}
		return m, nil

	case "c":
		if m.currentView == ListView && len(m.functions) > 0 {
			selectedIdx := m.table.Cursor()
			if selectedIdx < len(m.functions) {
				m.selectedFunc = &m.functions[selectedIdx]
				m.currentView = CodeView
				m.viewport.SetContent("Loading code...")
				return m, m.fetchFunctionCode(m.selectedFunc.Name)
			}
		}
		return m, nil

	case "m":
		logger.Logger.Printf("'m' key pressed for metrics in view: %s", m.currentView.String())
		if m.currentView == ListView && len(m.functions) > 0 {
			selectedIdx := m.table.Cursor()
			logger.Logger.Printf("Selected function index: %d, total functions: %d", selectedIdx, len(m.functions))
			if selectedIdx < len(m.functions) {
				m.selectedFunc = &m.functions[selectedIdx]
				m.currentView = MetricsView
				logger.Logger.Printf("Switching to MetricsView for function: %s", m.selectedFunc.Name)
				m.viewport.SetContent("Loading metrics...")
				return m, m.fetchFunctionMetrics(m.selectedFunc.Name)
			}
		} else if m.currentView == MetricsView && m.selectedFunc != nil {
			// Refresh metrics when in MetricsView
			logger.Logger.Printf("Refreshing metrics for function: %s", m.selectedFunc.Name)
			m.viewport.SetContent("Refreshing metrics...")
			return m, m.fetchFunctionMetrics(m.selectedFunc.Name)
		}
		return m, nil

	case "w":
		logger.Logger.Printf("'w' key pressed for download in view: %s", m.currentView.String())
		if m.currentView == ListView && len(m.functions) > 0 {
			selectedIdx := m.table.Cursor()
			logger.Logger.Printf("Selected function index: %d, total functions: %d", selectedIdx, len(m.functions))
			if selectedIdx < len(m.functions) {
				selectedFunc := &m.functions[selectedIdx]
				logger.Logger.Printf("Starting download for function: %s", selectedFunc.Name)
				m.viewport.SetContent(fmt.Sprintf("Downloading code for %s...", selectedFunc.Name))
				return m, tea.Batch(
					func() tea.Msg { return downloadingMsg{functionName: selectedFunc.Name} },
					m.downloadFunctionCode(selectedFunc.Name),
				)
			} else {
				logger.Logger.Printf("Invalid function index: %d", selectedIdx)
			}
		} else {
			logger.Logger.Printf("Download not available - currentView: %s, functions count: %d", m.currentView.String(), len(m.functions))
		}
		return m, nil

	case "v":
		if m.currentView == CodeView && m.selectedFunc != nil {
			m.currentView = CodeDisplayView
			m.viewport.SetContent(fmt.Sprintf("Loading code files for %s...", m.selectedFunc.Name))
			return m, tea.Batch(
				func() tea.Msg { return loadingCodeFilesMsg{functionName: m.selectedFunc.Name} },
				m.loadCodeFiles(m.selectedFunc.Name),
			)
		}
		return m, nil

	case "e":
		if m.currentView == CodeView && m.selectedFunc != nil {
			if !m.editMode {
				// Enter edit mode
				m.editMode = true
				m.originalContent = m.viewport.View()
				m.textarea.SetValue(m.viewport.View())
				m.textarea.SetWidth(m.width - 4)
				m.textarea.SetHeight(m.height - 10)
				m.textarea.Focus()
			} else {
				// Exit edit mode without saving
				m.editMode = false
				m.viewport.SetContent(m.originalContent)
				m.textarea.Blur()
			}
		}
		return m, nil

	case "ctrl+s":
		if m.currentView == CodeView && m.editMode && m.selectedFunc != nil {
			// Save the edited content
			editedContent := m.textarea.Value()
			m.viewport.SetContent(editedContent)
			m.editMode = false
			m.textarea.Blur()

			// TODO: Here you could add actual save functionality to persist changes
			// For now, we just update the display
			return m, func() tea.Msg {
				return editSavedMsg{success: true}
			}
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
	} else if m.currentView == CodeView && m.editMode {
		// Update textarea when in edit mode
		m.textarea, cmd = m.textarea.Update(msg)
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

// renderMetricsContent renders the metrics overview using charts
func renderMetricsContent(metrics *provider.FunctionMetrics, width int) string {
	if metrics == nil {
		return "No metrics data available"
	}

	// Add some debug info
	debug := fmt.Sprintf("DEBUG: Function: %s\n", metrics.FunctionName)
	debug += fmt.Sprintf("Invocations data points: %d\n", len(metrics.Invocations.DataPoints))
	debug += fmt.Sprintf("Duration data points: %d\n", len(metrics.Duration.DataPoints))
	debug += fmt.Sprintf("Width: %d\n\n", width)

	chartContent := charts.RenderMetricsOverview(metrics, width)
	return debug + chartContent
}
