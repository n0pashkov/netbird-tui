package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderSettings(m *Model) string {
	var sb strings.Builder

	sb.WriteString(styleTitle.Render("Settings") + "\n\n")

	// Current config info (read-only)
	if m.config != nil {
		sb.WriteString(styleLabel.Render("Current Config:") + "\n")
		sb.WriteString("  " + styleNeutral.Render("Management URL: ") + styleValue.Render(m.config.ManagementUrl) + "\n")
		if m.config.ConfigFile != "" {
			sb.WriteString("  " + styleNeutral.Render("Config file:    ") + styleValue.Render(m.config.ConfigFile) + "\n")
		}
		if m.config.LogFile != "" {
			sb.WriteString("  " + styleNeutral.Render("Log file:       ") + styleValue.Render(m.config.LogFile) + "\n")
		}
		sb.WriteString("\n")
	}

	// Setup Key field
	skiLabel := labelForField(0, m.settingsFocused, m.settingsEditing, "Setup Key:")
	sb.WriteString(skiLabel + "\n")
	sb.WriteString("  " + m.setupKeyInput.View() + "\n\n")

	// Management URL field
	muiLabel := labelForField(1, m.settingsFocused, m.settingsEditing, "Management URL:")
	sb.WriteString(muiLabel + "\n")
	sb.WriteString("  " + m.mgmtURLInput.View() + "\n\n")

	// Status message
	if m.settingsMsg != "" {
		if strings.HasPrefix(m.settingsMsg, "Error") {
			sb.WriteString(styleError.Render(m.settingsMsg) + "\n\n")
		} else {
			sb.WriteString(styleOnline.Render(m.settingsMsg) + "\n\n")
		}
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

// labelForField returns the styled label for a settings field.
// fieldIdx: index of this field (0 or 1)
// focused: currently focused field index
// editing: whether we're in edit mode
// label: raw label text
func labelForField(fieldIdx, focused int, editing bool, label string) string {
	if fieldIdx != focused {
		return styleNeutral.Render(label)
	}
	if editing {
		// active edit indicator
		return lipgloss.NewStyle().Foreground(colorBlue).Bold(true).Render("✎ " + label)
	}
	// browse focus indicator
	return lipgloss.NewStyle().Foreground(colorBlue).Bold(true).Render("▶ " + label)
}
