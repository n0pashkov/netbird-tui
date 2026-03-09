package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
)

func buildPeersTable(peers []*proto.PeerState, width, height int) table.Model {
	available := width - 6
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
