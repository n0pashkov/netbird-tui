package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorGreen  = lipgloss.Color("#22c55e")
	colorRed    = lipgloss.Color("#ef4444")
	colorYellow = lipgloss.Color("#eab308")
	colorBlue   = lipgloss.Color("#3b82f6")
	colorGray   = lipgloss.Color("#6b7280")
	colorWhite  = lipgloss.Color("#f9fafb")
	colorBorder = lipgloss.Color("#374151")

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlue).
			Padding(0, 1)

	styleHeader = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			BorderBottom(true).
			Padding(0, 1)

	styleFooter = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			BorderTop(true).
			Padding(0, 1).
			Foreground(colorGray)

	styleActiveTab = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlue).
			Padding(0, 1)

	styleInactiveTab = lipgloss.NewStyle().
				Foreground(colorGray).
				Padding(0, 1)

	styleOnline  = lipgloss.NewStyle().Foreground(colorGreen)
	styleOffline = lipgloss.NewStyle().Foreground(colorRed)
	styleNeutral = lipgloss.NewStyle().Foreground(colorGray)

	styleBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	styleLabel = lipgloss.NewStyle().
			Foreground(colorGray).
			Width(22)

	styleValue = lipgloss.NewStyle().
			Foreground(colorWhite)

	styleError = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)
)

// Ensure styleBox is used (referenced in status.go if needed)
var _ = styleBox
