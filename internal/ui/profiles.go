package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
)

func buildProfilesTable(profiles []*proto.Profile, activeProfile string, width, height int) table.Model {
	available := width - 18
	if available < 40 {
		available = 40
	}
	tableHeight := height - 18
	if tableHeight < 3 {
		tableHeight = 3
	}

	nameW := available * 50 / 100
	activeW := 10
	if nameW < 20 {
		nameW = 20
	}

	columns := []table.Column{
		{Title: "Name", Width: nameW},
		{Title: "Active", Width: activeW},
	}

	rows := make([]table.Row, 0, len(profiles))
	for _, p := range profiles {
		active := "No"
		if p.IsActive || p.Name == activeProfile {
			active = "● Yes"
		}
		rows = append(rows, table.Row{p.Name, active})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		BorderBottom(true).
		Bold(true).
		Foreground(colorBlue)
	s.Selected = s.Selected.
		Foreground(colorWhite).
		Background(lipgloss.Color("#1e3a5f")).
		Bold(false)
	t.SetStyles(s)

	return t
}

func renderProfiles(m *Model) string {
	var sb strings.Builder

	summaryStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	sb.WriteString(summaryStyle.Render("Profiles") + "\n")

	if m.activeProfile != "" {
		sb.WriteString(styleNeutral.Render("Active: ") + styleOnline.Render("● "+m.activeProfile) + "\n")
	}
	sb.WriteString(styleNeutral.Render(fmt.Sprintf("%d profile(s) configured", len(m.profiles))) + "\n\n")

	if m.features != nil && m.features.DisableProfiles {
		sb.WriteString(lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Render("  ⚠ Profile management is disabled by server policy") + "\n\n")
	}

	// Add profile form
	if m.profilesEditing {
		sb.WriteString(styleSectionHeader.Render("Add Profile") + "\n")

		lbl0 := profileFieldLabel(0, m.profilesFocused, m.profilesEditing, "Profile Name:")
		sb.WriteString(lbl0 + "\n")
		sb.WriteString("  " + m.profileNameInput.View() + "\n\n")

		sb.WriteString(styleNeutral.Render("ctrl+s:Add  Esc:Cancel  Tab:Next") + "\n\n")
	}

	// Profile message
	if m.profilesMsg != "" {
		if strings.HasPrefix(m.profilesMsg, "Error") {
			sb.WriteString(styleError.Render(m.profilesMsg) + "\n\n")
		} else {
			sb.WriteString(styleOnline.Render(m.profilesMsg) + "\n\n")
		}
	}

	// Profiles table
	if len(m.profiles) == 0 {
		sb.WriteString(styleNeutral.Render("No profiles configured") + "\n")
		sb.WriteString(styleNeutral.Render("Press n to add a profile") + "\n")
	} else {
		sb.WriteString(m.profilesTable.View() + "\n\n")
		sb.WriteString(styleNeutral.Render("Enter:Switch  n:New  x:Remove  ↑↓:Navigate"))
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

func profileFieldLabel(fieldIdx, focused int, editing bool, label string) string {
	if fieldIdx != focused {
		return styleNeutral.Render(label)
	}
	if editing {
		return lipgloss.NewStyle().Foreground(colorBlue).Bold(true).Render("✎ " + label)
	}
	return lipgloss.NewStyle().Foreground(colorBlue).Bold(true).Render("▶ " + label)
}
