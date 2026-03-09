package ui

import (
	"fmt"
	"strings"

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

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

// ensure proto import is used
var _ *proto.FullStatus
