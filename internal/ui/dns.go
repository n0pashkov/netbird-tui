package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
)

func renderDNS(m *Model) string {
	if m.status == nil || m.status.FullStatus == nil {
		return styleNeutral.Padding(1, 2).Render("No DNS data available")
	}

	fs := m.status.FullStatus
	if len(fs.DnsServers) == 0 {
		return styleNeutral.Padding(1, 2).Render("No DNS servers configured")
	}

	var sb strings.Builder

	// Summary header
	total := len(fs.DnsServers)
	active := 0
	errCount := 0
	for _, ns := range fs.DnsServers {
		if ns.Enabled {
			active++
		}
		if ns.Error != "" {
			errCount++
		}
	}

	summaryStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	sb.WriteString(summaryStyle.Render("DNS Nameserver Groups") + "\n")
	sb.WriteString(styleNeutral.Render(fmt.Sprintf("  %d groups  •  %d active  •  %d errors", total, active, errCount)) + "\n\n")

	if m.config != nil && m.config.DisableDns {
		sb.WriteString(lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Render("  ⚠ DNS is globally disabled in config") + "\n\n")
	}

	// Render each nameserver group
	for i, ns := range fs.DnsServers {
		selected := i == m.dnsSelected

		// Group border style
		groupStyle := styleBox.Width(m.width - 8)
		if selected {
			groupStyle = groupStyle.BorderForeground(colorBlue)
		}

		var groupSB strings.Builder

		// Status indicator
		indicator := "○"
		indStyle := styleNeutral
		if ns.Error != "" {
			indicator = "✗"
			indStyle = styleError
		} else if ns.Enabled {
			indicator = "●"
			indStyle = styleOnline
		}

		// Group header
		groupHeader := indStyle.Render(indicator) + "  "
		if len(ns.Domains) > 0 {
			groupHeader += styleValue.Bold(true).Render(strings.Join(ns.Domains, ", "))
		} else {
			groupHeader += styleValue.Bold(true).Render("(match-all)")
		}
		if ns.Enabled {
			groupHeader += "  " + styleOnline.Render("enabled")
		} else {
			groupHeader += "  " + styleNeutral.Render("disabled")
		}
		groupSB.WriteString(groupHeader + "\n")

		// Servers
		if len(ns.Servers) > 0 {
			groupSB.WriteString(styleNeutral.Render("  Servers: ") + styleValue.Render(strings.Join(ns.Servers, "  ")) + "\n")
		}

		// Domains detail
		if len(ns.Domains) > 0 {
			groupSB.WriteString(styleNeutral.Render("  Domains: "))
			for j, d := range ns.Domains {
				if j > 0 {
					groupSB.WriteString(", ")
				}
				groupSB.WriteString(styleValue.Render(d))
			}
			groupSB.WriteString("\n")
		}

		// Error
		if ns.Error != "" {
			groupSB.WriteString(styleError.Render("  Error: "+ns.Error) + "\n")
		}

		sb.WriteString(groupStyle.Render(groupSB.String()) + "\n")
	}

	// Config info
	sb.WriteString("\n")
	sb.WriteString(styleSectionHeader.Render("DNS Configuration") + "\n")

	if m.config != nil {
		lbl := styleLabel

		sb.WriteString(lbl.Render("DNS disabled:") + boolVal(styleValue, m.config.DisableDns) + "\n")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

// boolVal renders a bool as colored yes/no.
func boolVal(val lipgloss.Style, b bool) string {
	if b {
		return styleOnline.Render("Yes")
	}
	return styleOffline.Render("No")
}

// renderNSGroupState renders a single NSGroupState summary line.
func renderNSGroupState(ns *proto.NSGroupState) string {
	indicator := "○"
	indStyle := styleNeutral
	if ns.Error != "" {
		indicator = "✗"
		indStyle = styleError
	} else if ns.Enabled {
		indicator = "●"
		indStyle = styleOnline
	}

	domains := "(match-all)"
	if len(ns.Domains) > 0 {
		domains = strings.Join(ns.Domains, ", ")
	}

	servers := strings.Join(ns.Servers, " ")

	return indStyle.Render(indicator) + "  " + styleValue.Render(domains) + "  " + styleNeutral.Render(servers)
}
