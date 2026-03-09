package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderSettings(m *Model) string {
	var sb strings.Builder

	sb.WriteString(styleTitle.Render("Settings") + "\n\n")

	// Setup Key field
	skiLabel := "Setup Key:"
	if m.settingsFocused == 0 {
		sb.WriteString(styleActiveTab.Render(skiLabel) + "\n")
	} else {
		sb.WriteString(styleLabel.Render(skiLabel) + "\n")
	}
	sb.WriteString("  " + m.setupKeyInput.View() + "\n\n")

	// Management URL field
	muiLabel := "Management URL:"
	if m.settingsFocused == 1 {
		sb.WriteString(styleActiveTab.Render(muiLabel) + "\n")
	} else {
		sb.WriteString(styleLabel.Render(muiLabel) + "\n")
	}
	sb.WriteString("  " + m.mgmtURLInput.View() + "\n\n")

	// Status message
	if m.settingsMsg != "" {
		if strings.HasPrefix(m.settingsMsg, "Error") {
			sb.WriteString(styleError.Render(m.settingsMsg) + "\n\n")
		} else {
			sb.WriteString(styleOnline.Render(m.settingsMsg) + "\n\n")
		}
	}

	sb.WriteString(styleNeutral.Render("Tab/Enter: next field  •  ctrl+s: submit  •  esc: cancel"))

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}
