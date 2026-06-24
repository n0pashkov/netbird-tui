package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderServices(m *Model) string {
	var sb strings.Builder

	summaryStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	sb.WriteString(summaryStyle.Render("Services") + "  " + styleNeutral.Render("Forwarding rules, reverse proxy, and local service status") + "\n\n")

	sb.WriteString(styleSectionHeader.Render("Expose Service") + "\n")
	if m.exposeSession != nil {
		sb.WriteString("  " + styleOnline.Render("● Active") + styleNeutral.Render(fmt.Sprintf("  %s/%d", m.exposeSession.protocol, m.exposeSession.port)) + "\n")
		if m.exposeSession.name != "" {
			sb.WriteString("  " + styleNeutral.Render("Name: ") + styleValue.Render(m.exposeSession.name) + "\n")
		}
		if m.exposeSession.serviceURL != "" {
			sb.WriteString("  " + styleNeutral.Render("URL:  ") + styleValue.Render(m.exposeSession.serviceURL) + "\n")
		}
		if m.exposeSession.domain != "" {
			sb.WriteString("  " + styleNeutral.Render("Host: ") + styleValue.Render(m.exposeSession.domain) + "\n")
		}
		sb.WriteString(styleNeutral.Render("  x:stop expose") + "\n")
	} else {
		sb.WriteString(styleNeutral.Render("  No active expose session") + "\n")
	}
	sb.WriteString(styleNeutral.Render("  n:new expose") + "\n\n")

	if m.exposeEditing {
		sb.WriteString(styleSectionHeader.Render("New Expose") + "\n")
		fields := []struct {
			label string
			index int
			view  string
			help  string
		}{
			{"Port:", 0, m.exposePortInput.View(), "1-65535"},
			{"Protocol:", 1, m.exposeProtocolInput.View(), "http / https / tcp / udp / tls"},
			{"External Port:", 2, m.exposeExternalInput.View(), "tcp/udp/tls only"},
			{"Custom Domain:", 3, m.exposeDomainInput.View(), "optional"},
			{"Name Prefix:", 4, m.exposeNameInput.View(), "optional"},
			{"Password:", 5, m.exposePasswordInput.View(), "http/https only"},
			{"PIN:", 6, m.exposePinInput.View(), "6 digits, http/https only"},
			{"User Groups:", 7, m.exposeGroupsInput.View(), "comma-separated, http/https only"},
		}
		for _, field := range fields {
			sb.WriteString("  " + exposeFieldLabel(field.index, m.exposeFocused, field.label) + " " + field.view + "  " + styleNeutral.Render(field.help) + "\n")
		}
		sb.WriteString(styleNeutral.Render("  ctrl+s:Start  Tab:Next  Esc:Cancel") + "\n\n")
	}

	if m.exposeMsg != "" {
		if strings.HasPrefix(m.exposeMsg, "Error") {
			sb.WriteString(styleError.Render(m.exposeMsg) + "\n\n")
		} else {
			sb.WriteString(styleOnline.Render(m.exposeMsg) + "\n\n")
		}
	}

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

func exposeFieldLabel(index, focused int, label string) string {
	if index == focused {
		return lipgloss.NewStyle().Foreground(colorBlue).Bold(true).Render("✎ " + label)
	}
	return styleNeutral.Render(label)
}
