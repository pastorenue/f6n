package ui

// ViewType represents different views in the TUI
type ViewType int

const (
	// ListView shows the table of Lambda functions
	ListView ViewType = iota
	// DetailView shows detailed information about a selected function
	DetailView
	// LogsView shows logs for a selected function
	LogsView
	// CodeView shows the code for a selected function
	CodeView
	// CodeDisplayView shows downloaded code files in a scrollable pane
	CodeDisplayView
	// MetricsView shows metrics and charts for a selected function
	MetricsView
)

// String returns the string representation of the view type
func (v ViewType) String() string {
	switch v {
	case ListView:
		return "list"
	case DetailView:
		return "detail"
	case LogsView:
		return "logs"
	case CodeView:
		return "code"
	case CodeDisplayView:
		return "code-display"
	case MetricsView:
		return "metrics"
	default:
		return "unknown"
	}
}
