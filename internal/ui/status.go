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

	// ── Connection section ───────────────────────────────────────────────────
	sb.WriteString(styleSectionHeader.Render("Connection") + "\n")

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

	sb.WriteString(styleLabel.Render("Signal:"))
	if fs.SignalState != nil && fs.SignalState.Connected {
		sb.WriteString(styleOnline.Render("● Connected"))
		if fs.SignalState.URL != "" {
			sb.WriteString(styleNeutral.Render("  " + fs.SignalState.URL))
		}
	} else {
		sb.WriteString(styleOffline.Render("○ Disconnected"))
		if fs.SignalState != nil && fs.SignalState.Error != "" {
			sb.WriteString(styleError.Render("  " + fs.SignalState.Error))
		}
	}
	sb.WriteString("\n\n")

	// ── Local peer section ───────────────────────────────────────────────────
	if fs.LocalPeerState != nil {
		lp := fs.LocalPeerState
		sb.WriteString(styleSectionHeader.Render("Local Peer") + "\n")

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

		sb.WriteString(styleLabel.Render("Rosenpass:"))
		if lp.RosenpassEnabled {
			perm := ""
			if lp.RosenpassPermissive {
				perm = " (permissive)"
			}
			sb.WriteString(styleOnline.Render("● Enabled") + styleNeutral.Render(perm))
		} else {
			sb.WriteString(styleNeutral.Render("○ Disabled"))
		}
		sb.WriteString("\n\n")
	}

	// ── Network summary ──────────────────────────────────────────────────────
	sb.WriteString(styleSectionHeader.Render("Network") + "\n")

	total := len(fs.Peers)
	online := 0
	relayed := 0
	for _, p := range fs.Peers {
		if p.ConnStatus == "Connected" {
			online++
			if p.Relayed {
				relayed++
			}
		}
	}
	sb.WriteString(styleLabel.Render("Peers:"))
	sb.WriteString(styleValue.Render(fmt.Sprintf("%d online / %d total", online, total)))
	if relayed > 0 {
		sb.WriteString(styleNeutral.Render(fmt.Sprintf("  (%d relayed)", relayed)))
	}
	sb.WriteString("\n")

	// Routes summary
	if len(m.networks) > 0 {
		selected := 0
		for _, n := range m.networks {
			if n.Selected {
				selected++
			}
		}
		sb.WriteString(styleLabel.Render("Routes:"))
		sb.WriteString(styleValue.Render(fmt.Sprintf("%d selected / %d total", selected, len(m.networks))))
		sb.WriteString("\n")
	}

	// Relay states
	if len(fs.Relays) > 0 {
		avail := 0
		for _, r := range fs.Relays {
			if r.Available {
				avail++
			}
		}
		sb.WriteString(styleLabel.Render("Relays:"))
		sb.WriteString(styleValue.Render(fmt.Sprintf("%d/%d available", avail, len(fs.Relays))))
		sb.WriteString("\n")
		for _, r := range fs.Relays {
			if r.Available {
				sb.WriteString("  " + styleOnline.Render("● ") + styleNeutral.Render(r.URI) + "\n")
			} else {
				sb.WriteString("  " + styleOffline.Render("○ ") + styleNeutral.Render(r.URI))
				if r.Error != "" {
					sb.WriteString(styleError.Render(" [" + r.Error + "]"))
				}
				sb.WriteString("\n")
			}
		}
	}
	sb.WriteString("\n")

	// ── DNS summary ──────────────────────────────────────────────────────────
	if len(fs.DnsServers) > 0 {
		sb.WriteString(styleSectionHeader.Render("DNS") + "\n")
		errDNS := 0
		activeDNS := 0
		for _, ns := range fs.DnsServers {
			if ns.Enabled {
				activeDNS++
			}
			if ns.Error != "" {
				errDNS++
			}
		}
		sb.WriteString(styleLabel.Render("Nameservers:"))
		sb.WriteString(styleValue.Render(fmt.Sprintf("%d groups  •  %d active", len(fs.DnsServers), activeDNS)))
		if errDNS > 0 {
			sb.WriteString(styleError.Render(fmt.Sprintf("  •  %d error(s)", errDNS)))
		}
		sb.WriteString("\n")

		for _, ns := range fs.DnsServers {
			for _, srv := range ns.Servers {
				domains := ""
				if len(ns.Domains) > 0 {
					d := strings.Join(ns.Domains, ", ")
					if len(d) > 40 {
						d = d[:40] + "…"
					}
					domains = " (" + d + ")"
				} else {
					domains = " (match-all)"
				}
				if ns.Error != "" {
					sb.WriteString("  " + styleOffline.Render("○ ") + styleValue.Render(srv) + styleError.Render(" [err]") + "\n")
				} else if ns.Enabled {
					sb.WriteString("  " + styleOnline.Render("● ") + styleValue.Render(srv) + styleNeutral.Render(domains) + "\n")
				} else {
					sb.WriteString("  " + styleNeutral.Render("○ ") + styleNeutral.Render(srv) + styleNeutral.Render(domains) + "\n")
				}
			}
		}
		sb.WriteString("\n")
	}

	// ── SSH server ───────────────────────────────────────────────────────────
	if fs.SshServerState != nil {
		sb.WriteString(styleSectionHeader.Render("SSH Server") + "\n")
		sb.WriteString(styleLabel.Render("Status:"))
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
			for _, sess := range fs.SshServerState.Sessions {
				sb.WriteString("\n  " + styleNeutral.Render("•") + " " + styleValue.Render(sess.Username+"@"+sess.RemoteAddress))
			}
		} else {
			sb.WriteString(styleNeutral.Render("○ Disabled"))
		}
		sb.WriteString("\n\n")
	}

	// ── System events (last 5) ───────────────────────────────────────────────
	allEvents := fs.Events
	// Use fresh events from GetEvents if available
	if len(m.events) > 0 {
		allEvents = m.events
	}

	if len(allEvents) > 0 {
		sb.WriteString(styleSectionHeader.Render(fmt.Sprintf("Recent Events  (%d total, Events screen for full list)", len(allEvents))) + "\n")
		start := len(allEvents) - 5
		if start < 0 {
			start = 0
		}
		for _, ev := range allEvents[start:] {
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
				ts = "[" + ev.Timestamp.AsTime().Local().Format("15:04:05") + "] "
			}
			msg := ev.UserMessage
			if msg == "" {
				msg = ev.Message
			}
			if len(msg) > 70 {
				msg = msg[:70] + "…"
			}
			sb.WriteString("  " + evStyle.Render(icon+" "+ts+msg) + "\n")
		}
	}

	// ── Features (from management) ───────────────────────────────────────────
	if m.features != nil {
		sb.WriteString("\n")
		sb.WriteString(styleSectionHeader.Render("Server Features") + "\n")
		sb.WriteString(styleLabel.Render("Profiles:"))
		if m.features.DisableProfiles {
			sb.WriteString(styleNeutral.Render("Disabled by server"))
		} else {
			sb.WriteString(styleOnline.Render("Enabled"))
		}
		sb.WriteString("\n")
		sb.WriteString(styleLabel.Render("Update settings:"))
		if m.features.DisableUpdateSettings {
			sb.WriteString(styleNeutral.Render("Disabled by server"))
		} else {
			sb.WriteString(styleOnline.Render("Enabled"))
		}
		sb.WriteString("\n")
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
