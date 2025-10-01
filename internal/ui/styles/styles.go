package styles

import "github.com/charmbracelet/lipgloss"

// Color palette
const (
	ColorPrimary    = "#07646bff"
	ColorBackground = "#1a1a1a"
	ColorGray       = "#3a3a3a"
	ColorWhite      = "#FFFFFF"
	ColorDimmed     = "#808080"     // Grey for command values
	ColorYellow     = "#FFD700"     // Yellow for ASCII art
	ColorPink       = "#FF69B4"     // Pink for command keys
)

// Styles for various UI components
var (
	ASCIIStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorYellow)).
			Bold(true)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary))

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWhite)).
			Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDimmed))

	// New styles for command shortcuts
	CommandKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPink)).
			Bold(true)

	CommandValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDimmed))

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary)).
			Bold(true)

	InfoLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary)).
			Bold(true)

	InfoValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00CED1")).  // Teal color
			Bold(false)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	ViewportStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorPrimary)).
			Padding(1, 2)
)
