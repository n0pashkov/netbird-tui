package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderServices(m *Model) string {
	var sb strings.Builder

	summaryStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	sb.WriteString(summaryStyle.Render("Services") + "  " + styleNeutral.Render("Forwarding rules and local service status") + "\n\n")

	// Active Forwarding Rules section
	sb.WriteString(styleSectionHeader.Render("Forwarding Rules") + "\n")

	if len(m.fwdRules) == 0 {
		sb.WriteString(styleNeutral.Render("  No forwarding rules active") + "\n")
	} else {
		sb.WriteString(styleNeutral.Render(fmt.Sprintf("  %d rule(s) active", len(m.fwdRules))) + "\n\n")
		sb.WriteString(lipgloss.NewStyle().Padding(0, 0).Render(m.fwdTable.View()))
	}

	// Network services info
	sb.WriteString("\n")
	sb.WriteString(styleSectionHeader.Render("Network Status") + "\n")
	if m.status != nil && m.status.FullStatus != nil {
		fs := m.status.FullStatus
		lbl := lipgloss.NewStyle().Foreground(colorGray).Width(22)

		if fs.LocalPeerState != nil {
			sb.WriteString("  " + lbl.Render("Local IP:") + styleValue.Render(fs.LocalPeerState.IP) + "\n")
			sb.WriteString("  " + lbl.Render("FQDN:") + styleValue.Render(fs.LocalPeerState.Fqdn) + "\n")
		}

		// Peers with open connections
		connected := 0
		for _, p := range fs.Peers {
			if p.ConnStatus == "Connected" {
				connected++
			}
		}
		sb.WriteString("  " + lbl.Render("Connected peers:") + styleValue.Render(fmt.Sprintf("%d / %d", connected, len(fs.Peers))) + "\n")

		// SSH server
		if fs.SshServerState != nil && fs.SshServerState.Enabled {
			sessions := len(fs.SshServerState.Sessions)
			sb.WriteString("  " + lbl.Render("SSH server:") + styleOnline.Render("enabled"))
			if sessions > 0 {
				sb.WriteString(styleNeutral.Render(fmt.Sprintf("  (%d session(s))", sessions)))
				for _, sess := range fs.SshServerState.Sessions {
					sb.WriteString("\n    " + styleNeutral.Render("•") + " " + styleValue.Render(sess.Username+"@"+sess.RemoteAddress))
				}
			}
			sb.WriteString("\n")
		}
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}
