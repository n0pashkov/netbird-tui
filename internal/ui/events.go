package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
)

func buildEventsTable(events []*proto.SystemEvent, filter proto.SystemEvent_Severity, width, height int) table.Model {
	available := width - 18
	if available < 60 {
		available = 60
	}
	tableHeight := height - 14
	if tableHeight < 3 {
		tableHeight = 3
	}

	timeW := 8
	sevW := 10
	catW := 14
	msgW := available - timeW - sevW - catW
	if msgW < 20 {
		msgW = 20
	}

	columns := []table.Column{
		{Title: "Time", Width: timeW},
		{Title: "Severity", Width: sevW},
		{Title: "Category", Width: catW},
		{Title: "Message", Width: msgW},
	}

	rows := make([]table.Row, 0, len(events))
	for _, ev := range events {
		if filter >= 0 && ev.Severity != filter {
			continue
		}

		ts := ""
		if ev.Timestamp != nil {
			ts = ev.Timestamp.AsTime().Local().Format("15:04:05")
		}

		sev := severityLabel(ev.Severity)
		cat := categoryLabel(ev.Category)

		msg := ev.UserMessage
		if msg == "" {
			msg = ev.Message
		}
		if len(msg) > msgW-2 {
			msg = msg[:msgW-5] + "…"
		}

		rows = append(rows, table.Row{ts, sev, cat, msg})
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

func renderEvents(m *Model) string {
	// Detail view
	if m.eventsDetail {
		return renderEventDetail(m)
	}

	var sb strings.Builder

	// Header
	total := len(m.events)
	shown := total
	filterLabel := "All"
	if m.eventsFilter >= 0 {
		filterLabel = severityLabel(m.eventsFilter)
		shown = 0
		for _, ev := range m.events {
			if ev.Severity == m.eventsFilter {
				shown++
			}
		}
	}

	summaryStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	sb.WriteString(summaryStyle.Render("System Events") + "  ")
	sb.WriteString(styleNeutral.Render(fmt.Sprintf("%d total  •  showing: %s (%d)", total, filterLabel, shown)))
	sb.WriteString("  " + styleNeutral.Render("[f] to cycle filter") + "\n\n")

	if len(m.events) == 0 {
		sb.WriteString(styleNeutral.Render("No events recorded"))
		return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
	}

	sb.WriteString(lipgloss.NewStyle().Padding(0, 0).Render(m.eventsTable.View()))

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

func renderEventDetail(m *Model) string {
	row := m.eventsTable.SelectedRow()
	if row == nil {
		return styleNeutral.Padding(1, 2).Render("No event selected")
	}

	// Find the matching event
	idx := m.eventsTable.Cursor()
	filteredIdx := 0
	var ev *proto.SystemEvent
	for _, e := range m.events {
		if m.eventsFilter >= 0 && e.Severity != m.eventsFilter {
			continue
		}
		if filteredIdx == idx {
			ev = e
			break
		}
		filteredIdx++
	}

	if ev == nil {
		return styleNeutral.Padding(1, 2).Render("Event not found")
	}

	var sb strings.Builder

	lbl := lipgloss.NewStyle().Foreground(colorGray).Width(18)
	val := styleValue

	// Header with severity color
	sevStyle := severityStyle(ev.Severity)
	sb.WriteString(styleTitle.Render("Event Detail") + "  " + sevStyle.Render(severityLabel(ev.Severity)) + "\n\n")

	if ev.Timestamp != nil {
		t := ev.Timestamp.AsTime().Local()
		sb.WriteString(lbl.Render("Time:") + val.Render(t.Format("2006-01-02 15:04:05")) + "\n")
	}
	sb.WriteString(lbl.Render("ID:") + val.Render(ev.Id) + "\n")
	sb.WriteString(lbl.Render("Severity:") + sevStyle.Render(severityLabel(ev.Severity)) + "\n")
	sb.WriteString(lbl.Render("Category:") + val.Render(categoryLabel(ev.Category)) + "\n\n")

	if ev.UserMessage != "" {
		sb.WriteString(lbl.Render("User Message:") + "\n")
		sb.WriteString(wordWrap(ev.UserMessage, m.width-10) + "\n\n")
	}

	if ev.Message != "" {
		sb.WriteString(lbl.Render("Technical:") + "\n")
		sb.WriteString(styleNeutral.Render(wordWrap(ev.Message, m.width-10)) + "\n\n")
	}

	// Metadata
	if len(ev.Metadata) > 0 {
		sb.WriteString(styleSectionHeader.Render("Metadata") + "\n")
		for k, v := range ev.Metadata {
			sb.WriteString("  " + styleNeutral.Render(k+": ") + val.Render(v) + "\n")
		}
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

func severityLabel(s proto.SystemEvent_Severity) string {
	switch s {
	case proto.SystemEvent_INFO:
		return "INFO"
	case proto.SystemEvent_WARNING:
		return "WARN"
	case proto.SystemEvent_ERROR:
		return "ERROR"
	case proto.SystemEvent_CRITICAL:
		return "CRIT"
	default:
		return "UNKN"
	}
}

func severityStyle(s proto.SystemEvent_Severity) lipgloss.Style {
	switch s {
	case proto.SystemEvent_INFO:
		return lipgloss.NewStyle().Foreground(colorBlue)
	case proto.SystemEvent_WARNING:
		return lipgloss.NewStyle().Foreground(colorYellow)
	case proto.SystemEvent_ERROR:
		return styleError
	case proto.SystemEvent_CRITICAL:
		return lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	default:
		return styleNeutral
	}
}

func categoryLabel(c proto.SystemEvent_Category) string {
	switch c {
	case proto.SystemEvent_NETWORK:
		return "network"
	case proto.SystemEvent_DNS:
		return "dns"
	case proto.SystemEvent_AUTHENTICATION:
		return "auth"
	case proto.SystemEvent_CONNECTIVITY:
		return "connectivity"
	case proto.SystemEvent_SYSTEM:
		return "system"
	default:
		return "general"
	}
}

func wordWrap(text string, width int) string {
	if width <= 0 {
		width = 80
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	line := ""
	for _, w := range words {
		if line == "" {
			line = w
		} else if len(line)+1+len(w) <= width {
			line += " " + w
		} else {
			lines = append(lines, line)
			line = w
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
