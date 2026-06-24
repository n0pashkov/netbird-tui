package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
)

func buildStatesTable(states []*proto.State, width, height int) table.Model {
	available := width - 18
	if available < 40 {
		available = 40
	}
	tableHeight := height - 18
	if tableHeight < 3 {
		tableHeight = 3
	}

	columns := []table.Column{
		{Title: "State Name", Width: available},
	}

	rows := make([]table.Row, 0, len(states))
	for _, s := range states {
		rows = append(rows, table.Row{s.Name})
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

func renderDiagnostics(m *Model) string {
	switch m.diagMode {
	case diagModeTrace:
		return renderTracePacket(m)
	case diagModeStates:
		return renderStates(m)
	case diagModeOutput:
		return renderDebugOutput(m)
	}
	return renderDiagnosticsOverview(m)
}

func renderDiagnosticsOverview(m *Model) string {
	var sb strings.Builder

	summaryStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	sb.WriteString(summaryStyle.Render("Diagnostics") + "\n\n")

	// Log Level section
	sb.WriteString(styleSectionHeader.Render("Log Level") + "\n")
	if m.logLevelKnown {
		levelStr := logLevelName(m.logLevel)
		levelStyle := logLevelStyle(m.logLevel)
		sb.WriteString("  Current: " + levelStyle.Render(levelStr) + "\n")
		sb.WriteString(styleNeutral.Render("  l:increase  L:decrease") + "\n")
	} else {
		sb.WriteString(styleNeutral.Render("  Unknown (daemon not responding)") + "\n")
	}
	sb.WriteString("\n")

	// States section
	sb.WriteString(styleSectionHeader.Render("Internal States") + "\n")
	if len(m.states) == 0 {
		sb.WriteString(styleNeutral.Render("  No states loaded") + "\n")
	} else {
		sb.WriteString(styleNeutral.Render(fmt.Sprintf("  %d state(s) registered", len(m.states))) + "\n")
		for _, s := range m.states {
			sb.WriteString("  " + styleNeutral.Render("•") + " " + styleValue.Render(s.Name) + "\n")
		}
	}
	sb.WriteString(styleNeutral.Render("  Press s to manage states") + "\n\n")

	// Packet Trace section
	sb.WriteString(styleSectionHeader.Render("Packet Tracer") + "\n")
	sb.WriteString(styleNeutral.Render("  Trace a packet through firewall rules to see if it would be allowed") + "\n")
	sb.WriteString(styleNeutral.Render("  Press t to open tracer") + "\n\n")

	// Debug Bundle section
	sb.WriteString(styleSectionHeader.Render("Debug Bundle") + "\n")
	sb.WriteString(styleNeutral.Render("  b:create bundle  B:create anonymized bundle  f:debug for 1m") + "\n\n")

	// Debug CLI section
	sb.WriteString(styleSectionHeader.Render("Debug Commands") + "\n")
	sb.WriteString(styleNeutral.Render("  c:dump config  a:capture packets for 10s") + "\n")
	sb.WriteString(styleNeutral.Render("  p:persistence on  P:persistence off") + "\n\n")

	// Connection test info
	sb.WriteString(styleSectionHeader.Render("Connection Info") + "\n")
	if m.status != nil && m.status.FullStatus != nil {
		fs := m.status.FullStatus
		lbl := lipgloss.NewStyle().Foreground(colorGray).Width(20)
		val := styleValue

		if fs.ManagementState != nil {
			connStr := "Connected"
			connStyle := styleOnline
			if !fs.ManagementState.Connected {
				connStr = "Disconnected"
				connStyle = styleOffline
			}
			sb.WriteString("  " + lbl.Render("Management:") + connStyle.Render(connStr))
			if fs.ManagementState.Error != "" {
				sb.WriteString("  " + styleError.Render(fs.ManagementState.Error))
			}
			sb.WriteString("\n")
		}
		if fs.SignalState != nil {
			connStr := "Connected"
			connStyle := styleOnline
			if !fs.SignalState.Connected {
				connStr = "Disconnected"
				connStyle = styleOffline
			}
			sb.WriteString("  " + lbl.Render("Signal:") + connStyle.Render(connStr))
			if fs.SignalState.Error != "" {
				sb.WriteString("  " + styleError.Render(fs.SignalState.Error))
			}
			sb.WriteString("\n")
		}
		if fs.LocalPeerState != nil {
			sb.WriteString("  " + lbl.Render("Interface:") + val.Render(func() string {
				if fs.LocalPeerState.KernelInterface {
					return "kernel (wt0)"
				}
				return "userspace"
			}()) + "\n")
		}

		// Relays
		if len(fs.Relays) > 0 {
			avail := 0
			for _, r := range fs.Relays {
				if r.Available {
					avail++
				}
			}
			sb.WriteString("  " + lbl.Render("Relays:") + val.Render(fmt.Sprintf("%d/%d available", avail, len(fs.Relays))) + "\n")
		}
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

func renderDebugOutput(m *Model) string {
	var sb strings.Builder

	title := m.debugTitle
	if title == "" {
		title = "Debug Output"
	}
	sb.WriteString(styleSectionHeader.Render(title) + "\n")
	sb.WriteString(styleNeutral.Render("Esc:back to Diagnostics") + "\n\n")

	if m.debugErr != "" {
		sb.WriteString(styleError.Render("Error: "+m.debugErr) + "\n\n")
	}
	if strings.TrimSpace(m.debugOutput) == "" {
		sb.WriteString(styleNeutral.Render("No output"))
	} else {
		sb.WriteString(styleValue.Render(fitContentWidth(m.debugOutput, m.width-8)))
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

func renderTracePacket(m *Model) string {
	var sb strings.Builder

	summaryStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	sb.WriteString(summaryStyle.Render("Packet Tracer") + "  ")
	sb.WriteString(styleNeutral.Render("Trace packet through firewall rules") + "\n\n")

	// Form
	sb.WriteString(styleSectionHeader.Render("Packet Parameters") + "\n")

	inputs := []struct {
		label string
		field int
		view  string
		help  string
	}{
		{"Source IP:", 0, m.traceSrcIP.View(), "e.g. 100.64.0.1"},
		{"Destination IP:", 1, m.traceDstIP.View(), "e.g. 100.64.0.2"},
		{"Protocol:", 2, m.traceProto.View(), "tcp / udp / icmp"},
		{"Source Port:", 3, m.traceSrcPort.View(), "0-65535"},
		{"Dest Port:", 4, m.traceDstPort.View(), "0-65535"},
		{"Direction:", 5, m.traceDir.View(), "in / out"},
	}

	lbl := lipgloss.NewStyle().Width(18)
	for _, inp := range inputs {
		focused := m.traceFocused == inp.field && m.traceEditing
		lblStr := ""
		if focused {
			lblStr = lipgloss.NewStyle().Foreground(colorBlue).Bold(true).Render("✎ " + inp.label)
		} else if m.traceEditing {
			lblStr = styleNeutral.Render(inp.label)
		} else {
			lblStr = lbl.Foreground(colorGray).Render(inp.label)
		}
		sb.WriteString("  " + lblStr + " " + inp.view + "  " + styleNeutral.Render(inp.help) + "\n")
	}
	sb.WriteString("\n")

	if m.traceEditing {
		sb.WriteString(styleNeutral.Render("Tab:Next  ctrl+s:Trace  Esc:Cancel") + "\n\n")
	} else {
		sb.WriteString(styleNeutral.Render("Enter/e:Edit  Esc:Back to Diagnostics") + "\n\n")
	}

	// Results
	if m.traceErr != "" {
		sb.WriteString(styleError.Render("Error: "+m.traceErr) + "\n")
	} else if m.traceResult != nil {
		sb.WriteString(styleSectionHeader.Render("Trace Result") + "\n")

		finalStyle := styleOnline
		finalStr := "ALLOWED"
		if !m.traceResult.FinalDisposition {
			finalStyle = styleOffline
			finalStr = "BLOCKED"
		}
		sb.WriteString("  Final: " + finalStyle.Bold(true).Render(finalStr) + "\n\n")

		if len(m.traceResult.Stages) > 0 {
			sb.WriteString(styleNeutral.Render("Stages:") + "\n")
			for _, stage := range m.traceResult.Stages {
				icon := "✓"
				stStyle := styleOnline
				if !stage.Allowed {
					icon = "✗"
					stStyle = styleOffline
				}
				sb.WriteString("  " + stStyle.Render(icon) + "  ")
				sb.WriteString(styleValue.Bold(true).Render(stage.Name) + "  ")
				sb.WriteString(styleNeutral.Render(stage.Message) + "\n")
				if stage.ForwardingDetails != nil && *stage.ForwardingDetails != "" {
					sb.WriteString("     " + styleNeutral.Render("→ "+*stage.ForwardingDetails) + "\n")
				}
			}
		}
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

func renderStates(m *Model) string {
	var sb strings.Builder

	summaryStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	sb.WriteString(summaryStyle.Render("Internal Daemon States") + "\n")
	sb.WriteString(styleNeutral.Render(fmt.Sprintf("%d state(s)  •  c:clean  x:delete  r:refresh  Esc:back", len(m.states))) + "\n\n")

	if len(m.states) == 0 {
		sb.WriteString(styleNeutral.Render("No states registered") + "\n")
	} else {
		sb.WriteString(m.statesTable.View() + "\n")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

func logLevelName(l proto.LogLevel) string {
	switch l {
	case proto.LogLevel_PANIC:
		return "PANIC"
	case proto.LogLevel_FATAL:
		return "FATAL"
	case proto.LogLevel_ERROR:
		return "ERROR"
	case proto.LogLevel_WARN:
		return "WARN"
	case proto.LogLevel_INFO:
		return "INFO"
	case proto.LogLevel_DEBUG:
		return "DEBUG"
	case proto.LogLevel_TRACE:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

func logLevelStyle(l proto.LogLevel) lipgloss.Style {
	switch l {
	case proto.LogLevel_PANIC, proto.LogLevel_FATAL:
		return lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	case proto.LogLevel_ERROR:
		return styleError
	case proto.LogLevel_WARN:
		return lipgloss.NewStyle().Foreground(colorYellow)
	case proto.LogLevel_INFO:
		return styleOnline
	case proto.LogLevel_DEBUG, proto.LogLevel_TRACE:
		return lipgloss.NewStyle().Foreground(colorBlue)
	default:
		return styleNeutral
	}
}
