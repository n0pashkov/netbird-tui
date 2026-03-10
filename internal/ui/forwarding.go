package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
)

func buildFwdTable(rules []*proto.ForwardingRule, width, height int) table.Model {
	// width - 2 (contentStyle) - 4 (Padding(0,2)) - 8 (4 cols × Padding(0,1) cell style)
	available := width - 14
	if available < 50 {
		available = 50
	}
	tableHeight := height - 12
	if tableHeight < 3 {
		tableHeight = 3
	}

	// Distribute: Protocol=15%, DestPort=20%, TranslatedAddr=35%, TranslatedPort=30%
	proto := available * 15 / 100
	destPort := available * 20 / 100
	transAddr := available * 35 / 100
	transPort := available - proto - destPort - transAddr

	if proto < 8 {
		proto = 8
	}
	if destPort < 10 {
		destPort = 10
	}
	if transAddr < 14 {
		transAddr = 14
	}
	if transPort < 10 {
		transPort = 10
	}

	columns := []table.Column{
		{Title: "Protocol", Width: proto},
		{Title: "Dest Port", Width: destPort},
		{Title: "Translated Address", Width: transAddr},
		{Title: "Translated Port", Width: transPort},
	}

	rows := make([]table.Row, 0, len(rules))
	for _, r := range rules {
		destPortStr := formatPortInfo(r.DestinationPort)
		transPortStr := formatPortInfo(r.TranslatedPort)
		addr := r.TranslatedAddress
		if addr == "" {
			addr = r.TranslatedHostname
		}
		rows = append(rows, table.Row{
			r.Protocol,
			destPortStr,
			addr,
			transPortStr,
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

func formatPortInfo(p *proto.PortInfo) string {
	if p == nil {
		return "---"
	}
	switch v := p.PortSelection.(type) {
	case *proto.PortInfo_Port:
		return fmt.Sprintf("%d", v.Port)
	case *proto.PortInfo_Range_:
		return fmt.Sprintf("%d-%d", v.Range.Start, v.Range.End)
	}
	return "---"
}
