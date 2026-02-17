package charts

import (
	"fmt"
	"strings"

	"f6n/internal/provider"

	"github.com/charmbracelet/lipgloss"
)

// ChartStyle defines the styling for charts
var ChartStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(1, 2)

// RenderSparkline creates a simple ASCII sparkline
func RenderSparkline(data []provider.MetricDataPoint, width int) string {
	if len(data) == 0 {
		return strings.Repeat("_", width)
	}

	// Find min and max values
	min, max := data[0].Value, data[0].Value
	for _, point := range data {
		if point.Value < min {
			min = point.Value
		}
		if point.Value > max {
			max = point.Value
		}
	}

	// Characters for different heights (from low to high)
	chars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	var result strings.Builder

	// Calculate how many data points per character
	pointsPerChar := float64(len(data)) / float64(width)

	for i := 0; i < width; i++ {
		// Get the data point index for this character position
		dataIndex := int(float64(i) * pointsPerChar)
		if dataIndex >= len(data) {
			dataIndex = len(data) - 1
		}

		value := data[dataIndex].Value

		// Normalize value to character index
		var charIndex int
		if max == min {
			charIndex = 0
		} else {
			normalized := (value - min) / (max - min)
			charIndex = int(normalized * float64(len(chars)-1))
		}

		result.WriteString(chars[charIndex])
	}

	return result.String()
}

// RenderBarChart creates a horizontal bar chart
func RenderBarChart(data []provider.MetricDataPoint, width, height int) string {
	if len(data) == 0 {
		return "No data available"
	}

	// Find max value for scaling
	max := data[0].Value
	for _, point := range data {
		if point.Value > max {
			max = point.Value
		}
	}

	var lines []string

	// Take the last 'height' data points
	startIndex := 0
	if len(data) > height {
		startIndex = len(data) - height
	}

	for i := startIndex; i < len(data) && len(lines) < height; i++ {
		point := data[i]

		// Calculate bar length
		barLength := int((point.Value / max) * float64(width-20)) // Reserve space for labels
		if barLength < 0 {
			barLength = 0
		}

		// Format timestamp
		timeStr := point.Timestamp.Format("15:04")

		// Create bar
		bar := strings.Repeat("█", barLength)
		if barLength < width-20 {
			bar += strings.Repeat(" ", width-20-barLength)
		}

		// Format value
		valueStr := fmt.Sprintf("%.1f", point.Value)

		line := fmt.Sprintf("%s │%s│ %s", timeStr, bar, valueStr)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// RenderTimeSeriesChart creates a time series line chart
func RenderTimeSeriesChart(data []provider.MetricDataPoint, width, height int, title string) string {
	if len(data) == 0 {
		return ChartStyle.Render(fmt.Sprintf("%s\n\nNo data available", title))
	}

	// Normalize dimensions and derive usable space for bars/rows
	usableWidth := width
	if usableWidth < 20 {
		usableWidth = 20
	}
	usableHeight := height
	if usableHeight < 3 {
		usableHeight = 3
	}

	// Find min and max values
	min, max := data[0].Value, data[0].Value
	for _, point := range data {
		if point.Value < min {
			min = point.Value
		}
		if point.Value > max {
			max = point.Value
		}
	}

	// Create a simple bar chart instead of complex line chart for better terminal compatibility
	var lines []string

	// Add title
	lines = append(lines, title)
	lines = append(lines, "")

	// Show recent values as bars
	recentCount := 8
	if usableHeight-2 < recentCount {
		recentCount = usableHeight - 2
	}
	if recentCount < 1 {
		recentCount = 1
	}

	if len(data) < recentCount {
		recentCount = len(data)
	}

	startIdx := len(data) - recentCount
	for i := startIdx; i < len(data); i++ {
		point := data[i]

		// Calculate bar length (max 30 chars)
		maxBar := usableWidth - 10
		if maxBar < 1 {
			maxBar = 1
		}

		barLength := maxBar
		if max > min {
			barLength = int((point.Value - min) / (max - min) * float64(maxBar))
		}
		if barLength < 1 {
			barLength = 1
		}

		bar := strings.Repeat("█", barLength)
		timeStr := point.Timestamp.Format("15:04")
		valueStr := fmt.Sprintf("%.1f", point.Value)

		line := fmt.Sprintf("%s │%s %s", timeStr, bar, valueStr)
		lines = append(lines, line)
	}

	// Add range info
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Range: %.1f - %.1f", min, max))

	content := strings.Join(lines, "\n")
	return ChartStyle.Render(content)
}

// RenderMetricsOverview creates a comprehensive metrics dashboard
func RenderMetricsOverview(metrics *provider.FunctionMetrics, width int) string {
	if metrics == nil {
		return ChartStyle.Render("No metrics data available")
	}

	var sections []string

	// Header
	header := fmt.Sprintf("📊 Metrics for %s", metrics.FunctionName)
	timeRange := fmt.Sprintf("Time Range: %s - %s",
		metrics.TimeRange.Start.Format("15:04"),
		metrics.TimeRange.End.Format("15:04"))
	sections = append(sections, header, timeRange, "")

	// Invocations chart
	if len(metrics.Invocations.DataPoints) > 0 {
		invocationsChart := RenderTimeSeriesChart(
			metrics.Invocations.DataPoints,
			width-8, 8,
			fmt.Sprintf("🔥 %s (%s)", metrics.Invocations.MetricName, metrics.Invocations.Unit))
		sections = append(sections, invocationsChart, "")
	}

	// Duration chart
	if len(metrics.Duration.DataPoints) > 0 {
		durationChart := RenderTimeSeriesChart(
			metrics.Duration.DataPoints,
			width-8, 8,
			fmt.Sprintf("⏱️  %s (%s)", metrics.Duration.MetricName, metrics.Duration.Unit))
		sections = append(sections, durationChart, "")
	}

	// Memory chart
	if len(metrics.Memory.DataPoints) > 0 {
		memoryChart := RenderTimeSeriesChart(
			metrics.Memory.DataPoints,
			width-8, 6,
			fmt.Sprintf("💾 %s (%s)", metrics.Memory.MetricName, metrics.Memory.Unit))
		sections = append(sections, memoryChart, "")
	}

	// Summary statistics
	if len(metrics.Invocations.DataPoints) > 0 {
		totalInvocations := 0.0
		for _, point := range metrics.Invocations.DataPoints {
			totalInvocations += point.Value
		}

		avgDuration := 0.0
		if len(metrics.Duration.DataPoints) > 0 {
			totalDuration := 0.0
			for _, point := range metrics.Duration.DataPoints {
				totalDuration += point.Value
			}
			avgDuration = totalDuration / float64(len(metrics.Duration.DataPoints))
		}

		summary := fmt.Sprintf(`📈 Summary Statistics:
• Total Invocations: %.0f
• Average Duration: %.2f ms
• Data Points: %d`,
			totalInvocations,
			avgDuration,
			len(metrics.Invocations.DataPoints))

		sections = append(sections, ChartStyle.Render(summary))
	}

	return strings.Join(sections, "\n")
}
