package ui

import (
	"fmt"
	"os/user"
	"runtime"
	"strings"

	"f6n/internal/provider"
	"f6n/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

// renderView renders the main view
func renderView(m Model) string {
	if m.err != nil {
		errView := fmt.Sprintf("\n  %s %v\n\n  Press q to quit.\n", 
			styles.ErrorStyle.Render("Error:"), m.err)
		return renderASCII(m.width) + "\n" + errView
	}

	if m.loading {
		return renderASCII(m.width) + "\n\n  Loading Lambda functions...\n\n"
	}

	var content string

	// ASCII Art Header
	ascii := renderASCII(m.width)

	// Info rows
	info := renderInfo(m)

	// Shortcuts
	shortcuts := renderShortcuts()

	// Filter/Command input (show when in input mode or when filter is active)
	var inputBox string
	if m.inputMode == FilterMode || m.inputMode == CommandMode {
		inputBox = m.textInput.View() + "\n"
	} else if m.filterActive {
		// Show active filter indicator
		filterIndicator := styles.CommandKeyStyle.Render("Filter active:") + " " + 
			styles.InfoValueStyle.Render(m.activeFilter) + " " +
			styles.HelpStyle.Render("(press Esc to clear)")
		inputBox = filterIndicator + "\n"
	}

	// Main content
	if len(m.functions) == 0 {
		content = "\n  No Lambda functions found in this region.\n\n  " +
			styles.HelpStyle.Render("Press 'r' to refresh or 'q' to quit")
	} else if m.currentView == ListView {
		content = inputBox + m.table.View()
	} else {
		content = m.viewport.View()
	}

	// Help text
	var help string
	if m.currentView == ListView {
		help = styles.HelpStyle.Render("Use keyboard shortcuts above to navigate")
	} else {
		help = styles.HelpStyle.Render("↑/↓: scroll • esc: back • q: quit")
	}

	// Combine ASCII art, info and shortcuts side by side
	headerLayout := lipgloss.JoinHorizontal(
		lipgloss.Top,
		"    ", // Spacer
		info,
		"    ", // Spacer
		shortcuts,
	)

	logoLayout := lipgloss.NewStyle().
		MarginRight(4).
		Render(ascii)

	// Combine all elements
	view := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", logoLayout, headerLayout, content, help)
	
	// Apply top padding using lipgloss style
	paddedView := lipgloss.NewStyle().
		PaddingTop(2).
		Render(view)
	
	return paddedView
}

// renderASCII renders the ASCII art logo centered
func renderASCII(width int) string {
	art := `  _____  ________       
_/ ____\/  _____/ ____  
\   __\/   __  \ /    \ 
 |  |  \  |__\  \   |  \
 |__|   \_____  /___|  /
              \/     \/ `
	
	// Apply yellow color first
	styledArt := styles.ASCIIStyle.Render(art)
	
	// Center-align the ASCII art with the terminal width
	if width > 0 {
		return lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Render(styledArt)
	}
	
	return styledArt
}

// renderInfo renders the info section in a single column
func renderInfo(m Model) string {
	providerName := string(m.provider.GetProviderName())
	region := m.provider.GetRegion()
	accountID := m.accountID

	accountKey := "Account"
	if providerName == "gcp" {
		accountKey = "Project"
	}

	info := []struct {
		key   string
		value string
	}{
		{"Provider", strings.ToUpper(providerName)},
		{accountKey, accountID},
		{"Region", region},
		{"Environment", m.environment},
		{"Functions", fmt.Sprintf("%d", len(m.functions))},
		{"CPU", getCPUInfo()},
		{"MEM", getMemInfo()},
		{"OS", getOSInfo()},
		{"User", getUserInfo()},
	}

	// Build info in single column
	var lines []string
	for _, item := range info {
		// Pink for key, teal for value
		line := styles.CommandKeyStyle.Render(item.key+":") + " " + styles.InfoValueStyle.Render(item.value)
		lines = append(lines, line)
	}

	if providerName == "gcp" {
		lines = append(lines, styles.HelpStyle.Render("\n(Cloud Functions, 1st Gen)"))
	}

	return strings.Join(lines, "\n")
}

// getCPUInfo returns CPU architecture information
func getCPUInfo() string {
	return runtime.GOARCH
}

// getMemInfo returns available memory information
func getMemInfo() string {
	// For simplicity, return number of CPUs as a proxy
	// In production, you'd want to use a proper memory library
	return fmt.Sprintf("%d cores", runtime.NumCPU())
}

// getOSInfo returns operating system information
func getOSInfo() string {
	return runtime.GOOS
}

// getUserInfo returns current user information
func getUserInfo() string {
	currentUser, err := user.Current()
	if err != nil {
		return "unknown"
	}
	return currentUser.Username
}

// renderShortcuts renders the keyboard shortcuts bar in a single column
func renderShortcuts() string {
	shortcuts := []struct {
		key   string
		value string
	}{
		{"<enter>", "view details"},
		{"<\\>", "filter"},
		{"<:>", "command"},
		{"<l>", "logs"},
		{"<a>", "api gateway"},
		{"<c>", "view code"},
		{"<r>", "refresh"},
		{"<q>", "quit"},
	}

	// Build shortcuts in single column
	var lines []string
	for _, s := range shortcuts {
		// Pink for key, grey for value
		line := styles.CommandKeyStyle.Render(s.key) + ": " + styles.CommandValueStyle.Render(s.value)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// formatFunctionDetails formats detailed function information for display
func formatFunctionDetails(fn *provider.FunctionInfo) string {
	if fn == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(styles.SelectedStyle.Render("━━━ Function Details ━━━") + "\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Name: "))
	b.WriteString(fn.Name + "\n\n")

	if fn.ARN != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("ARN/Resource: "))
		b.WriteString(fn.ARN + "\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Runtime: "))
	b.WriteString(fn.Runtime + "\n\n")

	if fn.Handler != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Handler: "))
		b.WriteString(fn.Handler + "\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Memory: "))
	b.WriteString(fmt.Sprintf("%d MB\n\n", fn.Memory))

	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Timeout: "))
	b.WriteString(fmt.Sprintf("%d seconds\n\n", fn.Timeout))

	if fn.Region != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Region/Location: "))
		b.WriteString(fn.Region + "\n\n")
	}

	if fn.Description != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Description: "))
		b.WriteString(fn.Description + "\n\n")
	}

	if fn.Role != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Role: "))
		b.WriteString(fn.Role + "\n\n")
	}

	if fn.LastModified != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Last Modified: "))
		b.WriteString(fn.LastModified + "\n\n")
	}

	if len(fn.Environment) > 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("Environment Variables:\n"))
		for k, v := range fn.Environment {
			b.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	return b.String()
}
