package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
)

func renderStatus(m *Model) string {
	if m.err != nil {
		return styleError.Render("Error: " + m.err.Error())
	}
	if m.status == nil {
		return styleNeutral.Render("Loading...")
	}

	fs := m.status.FullStatus
	if fs == nil {
		return styleNeutral.Render("No status available")
	}

	var sb strings.Builder

	// Management state
	sb.WriteString(styleLabel.Render("Management:"))
	if fs.ManagementState != nil && fs.ManagementState.Connected {
		sb.WriteString(styleOnline.Render("● Connected"))
		if fs.ManagementState.URL != "" {
			sb.WriteString(styleNeutral.Render("  " + fs.ManagementState.URL))
		}
	} else {
		sb.WriteString(styleOffline.Render("○ Disconnected"))
		if fs.ManagementState != nil && fs.ManagementState.Error != "" {
			sb.WriteString(styleError.Render("  " + fs.ManagementState.Error))
		}
	}
	sb.WriteString("\n")

	// Signal state
	sb.WriteString(styleLabel.Render("Signal:"))
	if fs.SignalState != nil && fs.SignalState.Connected {
		sb.WriteString(styleOnline.Render("● Connected"))
		if fs.SignalState.URL != "" {
			sb.WriteString(styleNeutral.Render("  " + fs.SignalState.URL))
		}
	} else {
		sb.WriteString(styleOffline.Render("○ Disconnected"))
	}
	sb.WriteString("\n\n")

	// Local peer info
	if fs.LocalPeerState != nil {
		lp := fs.LocalPeerState
		sb.WriteString(styleLabel.Render("IP:"))
		sb.WriteString(styleValue.Render(lp.IP))
		sb.WriteString("\n")

		sb.WriteString(styleLabel.Render("FQDN:"))
		sb.WriteString(styleValue.Render(lp.Fqdn))
		sb.WriteString("\n")

		sb.WriteString(styleLabel.Render("Kernel Interface:"))
		if lp.KernelInterface {
			sb.WriteString(styleOnline.Render("Yes"))
		} else {
			sb.WriteString(styleOffline.Render("No"))
		}
		sb.WriteString("\n")

		// Rosenpass
		sb.WriteString(styleLabel.Render("Rosenpass:"))
		if lp.RosenpassEnabled {
			sb.WriteString(styleOnline.Render("● Enabled"))
		} else {
			sb.WriteString(styleNeutral.Render("○ Disabled"))
		}
		sb.WriteString("\n\n")
	}

	// Peers summary
	total := len(fs.Peers)
	online := 0
	for _, p := range fs.Peers {
		if p.ConnStatus == "Connected" {
			online++
		}
	}
	sb.WriteString(styleLabel.Render("Peers:"))
	sb.WriteString(styleValue.Render(fmt.Sprintf("%d online / %d total", online, total)))
	sb.WriteString("\n")

	// Relay states
	if len(fs.Relays) > 0 {
		sb.WriteString("\n")
		sb.WriteString(styleLabel.Render("Relays:") + "\n")
		for _, r := range fs.Relays {
			if r.Available {
				sb.WriteString("  " + styleOnline.Render("● ") + styleValue.Render(r.URI) + "\n")
			} else {
				sb.WriteString("  " + styleOffline.Render("○ ") + styleValue.Render(r.URI) + "\n")
			}
		}
	}

	// DNS servers
	if len(fs.DnsServers) > 0 {
		sb.WriteString("\n")
		sb.WriteString(styleLabel.Render("DNS Servers:") + "\n")
		for _, ns := range fs.DnsServers {
			for _, srv := range ns.Servers {
				domains := ""
				if len(ns.Domains) > 0 {
					domains = " (" + strings.Join(ns.Domains, ", ") + ")"
				}
				if ns.Error != "" {
					sb.WriteString("  " + styleOffline.Render("○ ") + styleValue.Render(srv) + styleError.Render(" [error: "+ns.Error+"]") + "\n")
				} else if ns.Enabled {
					sb.WriteString("  " + styleOnline.Render("● ") + styleValue.Render(srv) + styleNeutral.Render(domains) + "\n")
				} else {
					sb.WriteString("  " + styleNeutral.Render("○ ") + styleValue.Render(srv) + styleNeutral.Render(domains) + "\n")
				}
			}
		}
	}

	// SSH server state
	if fs.SshServerState != nil {
		sb.WriteString("\n")
		sb.WriteString(styleLabel.Render("SSH Server:"))
		if fs.SshServerState.Enabled {
			sessions := len(fs.SshServerState.Sessions)
			sessionStr := ""
			if sessions > 0 {
				sessionStr = fmt.Sprintf("  (%d session", sessions)
				if sessions != 1 {
					sessionStr += "s"
				}
				sessionStr += ")"
			}
			sb.WriteString(styleOnline.Render("● Enabled") + styleNeutral.Render(sessionStr))
		} else {
			sb.WriteString(styleNeutral.Render("○ Disabled"))
		}
		sb.WriteString("\n")
	}

	// System events (last 3)
	if len(fs.Events) > 0 {
		sb.WriteString("\n")
		sb.WriteString(styleLabel.Render("Events:") + "\n")
		events := fs.Events
		start := len(events) - 3
		if start < 0 {
			start = 0
		}
		for _, ev := range events[start:] {
			icon := "ℹ"
			evStyle := styleNeutral
			switch ev.Severity {
			case proto.SystemEvent_WARNING:
				icon = "⚠"
				evStyle = lipgloss.NewStyle().Foreground(colorYellow)
			case proto.SystemEvent_ERROR, proto.SystemEvent_CRITICAL:
				icon = "✗"
				evStyle = styleError
			}
			ts := ""
			if ev.Timestamp != nil {
				ts = "[" + ev.Timestamp.AsTime().Local().Format("15:04") + "] "
			}
			msg := ev.UserMessage
			if msg == "" {
				msg = ev.Message
			}
			sb.WriteString("  " + evStyle.Render(icon+" "+ts+msg) + "\n")
		}
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

// relativeTime returns a human-readable relative time string.
func relativeTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
