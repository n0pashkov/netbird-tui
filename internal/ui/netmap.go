package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
)

// renderNetworkMap generates an ASCII network topology visualization.
// It's embedded in the Status tab as a compact overview.
func renderNetworkMap(m *Model, maxWidth int) string {
	if m.status == nil || m.status.FullStatus == nil {
		return ""
	}

	fs := m.status.FullStatus
	if fs.LocalPeerState == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(styleSectionHeader.Render("Network Map") + "\n\n")

	// Center node (local peer)
	localIP := fs.LocalPeerState.IP
	localFQDN := fs.LocalPeerState.Fqdn
	if len(localFQDN) > 20 {
		localFQDN = localFQDN[:20] + "…"
	}
	localLabel := fmt.Sprintf("[ %s ]", localFQDN)
	localIPLabel := fmt.Sprintf("  %s  ", localIP)

	centerStyle := lipgloss.NewStyle().Foreground(colorBlue).Bold(true)
	sb.WriteString(centerStyle.Render(localLabel) + "\n")
	sb.WriteString(styleNeutral.Render(localIPLabel) + "\n")

	if len(fs.Peers) == 0 {
		sb.WriteString(styleNeutral.Render("  (no peers)") + "\n")
		return sb.String()
	}

	// Sort peers: connected first
	connected := make([]*proto.PeerState, 0)
	disconnected := make([]*proto.PeerState, 0)
	for _, p := range fs.Peers {
		if p.ConnStatus == "Connected" {
			connected = append(connected, p)
		} else {
			disconnected = append(disconnected, p)
		}
	}

	// Show connections
	maxPeers := 8
	shown := 0

	renderPeerLine := func(p *proto.PeerState) {
		if shown >= maxPeers {
			return
		}
		shown++

		connector := "├──"
		if shown == len(connected)+len(disconnected) || shown == maxPeers {
			connector = "└──"
		}

		fqdn := p.Fqdn
		if len(fqdn) > 25 {
			fqdn = fqdn[:25] + "…"
		}

		connType := ""
		if p.Relayed {
			connType = styleNeutral.Render(" ~relay")
		} else if p.ConnStatus == "Connected" {
			connType = styleOnline.Render(" p2p")
		}

		latency := ""
		if p.Latency != nil && p.ConnStatus == "Connected" {
			d := p.Latency.AsDuration()
			if d > 0 {
				latency = styleNeutral.Render(fmt.Sprintf(" %dms", d.Milliseconds()))
			}
		}

		if p.ConnStatus == "Connected" {
			sb.WriteString(styleOnline.Render(connector) + " " + styleValue.Render(fqdn) + connType + latency + "\n")
		} else {
			sb.WriteString(styleNeutral.Render(connector) + " " + styleNeutral.Render(fqdn) + styleOffline.Render(" offline") + "\n")
		}
	}

	for _, p := range connected {
		renderPeerLine(p)
	}
	for _, p := range disconnected {
		renderPeerLine(p)
	}

	remaining := len(fs.Peers) - shown
	if remaining > 0 {
		sb.WriteString(styleNeutral.Render(fmt.Sprintf("    … and %d more peer(s)", remaining)) + "\n")
	}

	return sb.String()
}
