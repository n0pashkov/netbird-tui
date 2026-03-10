package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
)

func buildPeersTable(peers []*proto.PeerState, width, height int) table.Model {
	// width - 2 (contentStyle) - 4 (Padding(0,2)) - 12 (6 cols × Padding(0,1) cell style)
	available := width - 18
	if available < 60 {
		available = 60
	}
	tableHeight := height - 12
	if tableHeight < 3 {
		tableHeight = 3
	}

	// Distribute: FQDN=30%, IP=15%, Status=12%, Latency=11%, Relayed=11%, Rx/Tx=21%
	fqdn := available * 30 / 100
	ip := available * 15 / 100
	status := available * 12 / 100
	latency := available * 11 / 100
	relayed := available * 11 / 100
	rxtx := available - fqdn - ip - status - latency - relayed

	if fqdn < 12 {
		fqdn = 12
	}
	if ip < 10 {
		ip = 10
	}
	if status < 8 {
		status = 8
	}
	if latency < 7 {
		latency = 7
	}
	if relayed < 7 {
		relayed = 7
	}
	if rxtx < 8 {
		rxtx = 8
	}

	columns := []table.Column{
		{Title: "FQDN", Width: fqdn},
		{Title: "IP", Width: ip},
		{Title: "Status", Width: status},
		{Title: "Latency", Width: latency},
		{Title: "Relayed", Width: relayed},
		{Title: "Rx/Tx", Width: rxtx},
	}

	rows := make([]table.Row, 0, len(peers))
	for _, p := range peers {
		status := "○ Offline"
		if p.ConnStatus == "Connected" {
			status = "● Online"
		}

		latency := "---"
		if p.Latency != nil && p.ConnStatus == "Connected" {
			d := p.Latency.AsDuration()
			if d > 0 {
				latency = fmt.Sprintf("%dms", d.Milliseconds())
			}
		}

		relayed := "No"
		if p.Relayed {
			relayed = "Yes"
		}

		rx := formatBytes(p.BytesRx)
		tx := formatBytes(p.BytesTx)
		rxtx := fmt.Sprintf("%s/%s", rx, tx)

		rows = append(rows, table.Row{
			p.Fqdn,
			p.IP,
			status,
			latency,
			relayed,
			rxtx,
		})
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

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func peersHeader(peers []*proto.PeerState) string {
	total := len(peers)
	online := 0
	for _, p := range peers {
		if p.ConnStatus == "Connected" {
			online++
		}
	}
	return styleNeutral.Render(fmt.Sprintf("PEERS  (%d online / %d total)  — refreshed %s",
		online, total, time.Now().Format("15:04:05")))
}

func renderPeerDetail(peer *proto.PeerState, width int) string {
	var sb strings.Builder

	lbl := lipgloss.NewStyle().Foreground(colorGray).Width(28)
	val := styleValue

	// Header
	connStr := "○ Offline"
	connStyle := styleOffline
	if peer.ConnStatus == "Connected" {
		connStr = "● Online"
		connStyle = styleOnline
	}
	sb.WriteString(styleTitle.Render("Peer Detail") + "  " + connStyle.Render(connStr) + "\n\n")

	// Identity
	sb.WriteString(lbl.Render("FQDN:") + val.Render(peer.Fqdn) + "\n")
	sb.WriteString(lbl.Render("IP:") + val.Render(peer.IP) + "\n")
	if peer.PubKey != "" {
		key := peer.PubKey
		if len(key) > 32 {
			key = key[:32] + "…"
		}
		sb.WriteString(lbl.Render("Public Key:") + val.Render(key) + "\n")
	}
	sb.WriteString("\n")

	// Connection timing
	if peer.ConnStatusUpdate != nil {
		t := peer.ConnStatusUpdate.AsTime()
		sb.WriteString(lbl.Render("Status changed:") + val.Render(relativeTime(t)+" ("+t.Local().Format("15:04:05")+")") + "\n")
	}
	if peer.LastWireguardHandshake != nil {
		t := peer.LastWireguardHandshake.AsTime()
		if !t.IsZero() {
			sb.WriteString(lbl.Render("Last WG Handshake:") + val.Render(relativeTime(t)+" ("+t.Local().Format("15:04:05")+")") + "\n")
		}
	}
	sb.WriteString("\n")

	// Connection type
	connType := "P2P"
	if peer.Relayed {
		connType = "Relayed"
		if peer.RelayAddress != "" {
			connType += " (" + peer.RelayAddress + ")"
		}
	}
	sb.WriteString(lbl.Render("Connection:") + val.Render(connType) + "\n")

	// ICE candidates
	if peer.LocalIceCandidateType != "" || peer.RemoteIceCandidateType != "" {
		local := peer.LocalIceCandidateType
		if peer.LocalIceCandidateEndpoint != "" {
			local += " / " + peer.LocalIceCandidateEndpoint
		}
		remote := peer.RemoteIceCandidateType
		if peer.RemoteIceCandidateEndpoint != "" {
			remote += " / " + peer.RemoteIceCandidateEndpoint
		}
		sb.WriteString(lbl.Render("ICE Local:") + val.Render(local) + "\n")
		sb.WriteString(lbl.Render("ICE Remote:") + val.Render(remote) + "\n")
	}
	sb.WriteString("\n")

	// Stats
	if peer.Latency != nil {
		d := peer.Latency.AsDuration()
		if d > 0 {
			sb.WriteString(lbl.Render("Latency:") + val.Render(fmt.Sprintf("%dms", d.Milliseconds())) + "\n")
		}
	}
	sb.WriteString(lbl.Render("Rx / Tx:") + val.Render(formatBytes(peer.BytesRx)+" / "+formatBytes(peer.BytesTx)) + "\n")
	sb.WriteString("\n")

	// Rosenpass
	rosenpass := "No"
	if peer.RosenpassEnabled {
		rosenpass = "Yes"
	}
	sb.WriteString(lbl.Render("Rosenpass:") + val.Render(rosenpass) + "\n")

	// SSH host key
	sshKey := "absent"
	if len(peer.SshHostKey) > 0 {
		sshKey = "present"
	}
	sb.WriteString(lbl.Render("SSH Host Key:") + val.Render(sshKey) + "\n")

	// Networks
	if len(peer.Networks) > 0 {
		sb.WriteString("\n")
		sb.WriteString(lbl.Render("Networks:") + "\n")
		for _, n := range peer.Networks {
			sb.WriteString("  " + styleNeutral.Render("•") + " " + val.Render(n) + "\n")
		}
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}
