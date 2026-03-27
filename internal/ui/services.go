package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderServices(m *Model) string {
	var sb strings.Builder

	summaryStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	sb.WriteString(summaryStyle.Render("Services & Port Forwarding") + "\n\n")

	// Expose Service form
	if m.svcEditing {
		sb.WriteString(styleSectionHeader.Render("Expose Local Service") + "\n")

		fields := []struct {
			label string
			idx   int
			view  string
			help  string
		}{
			{"Port:", 0, m.svcPortInput.View(), "local port to expose"},
			{"Protocol:", 1, m.svcProtoInput.View(), "tcp / udp"},
			{"Groups:", 2, m.svcGroupInput.View(), "comma-separated NetBird groups"},
			{"Domain prefix:", 3, m.svcDomainInput.View(), "optional custom domain prefix"},
		}

		for _, f := range fields {
			lbl := ""
			if m.svcFocused == f.idx {
				lbl = lipgloss.NewStyle().Foreground(colorBlue).Bold(true).Render("✎ " + f.label)
			} else {
				lbl = lipgloss.NewStyle().Foreground(colorGray).Width(14).Render(f.label)
			}
			sb.WriteString("  " + lbl + "  " + f.view + "\n")
			sb.WriteString("  " + lipgloss.NewStyle().Width(16).Render("") + styleNeutral.Render(f.help) + "\n\n")
		}

		sb.WriteString(styleNeutral.Render("ctrl+s:Expose  Tab:Next  Esc:Cancel") + "\n\n")
	} else {
		sb.WriteString(styleNeutral.Render("n:Expose a local service") + "\n\n")
	}

	if m.svcMsg != "" {
		if strings.HasPrefix(m.svcMsg, "Error") {
			sb.WriteString(styleError.Render(m.svcMsg) + "\n\n")
		} else {
			sb.WriteString(styleOnline.Render(m.svcMsg) + "\n\n")
		}
	}

	// Active Forwarding Rules section
	sb.WriteString(styleSectionHeader.Render("Active Forwarding Rules") + "\n")

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
