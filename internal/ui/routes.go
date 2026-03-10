package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
)

func buildRoutesTable(networks []*proto.Network, width, height int) table.Model {
	// width - 2 (contentStyle) - 4 (Padding(0,2)) - 8 (4 cols × Padding(0,1) cell style)
	available := width - 14
	if available < 50 {
		available = 50
	}
	tableHeight := height - 12
	if tableHeight < 3 {
		tableHeight = 3
	}

	// Distribute: NetworkID=25%, CIDR=30%, Selected=12%, Domains=33%
	netID := available * 25 / 100
	cidr := available * 30 / 100
	selected := available * 12 / 100
	domains := available - netID - cidr - selected

	if netID < 10 {
		netID = 10
	}
	if cidr < 12 {
		cidr = 12
	}
	if selected < 8 {
		selected = 8
	}
	if domains < 10 {
		domains = 10
	}

	columns := []table.Column{
		{Title: "Network ID", Width: netID},
		{Title: "CIDR / Domains", Width: cidr},
		{Title: "Selected", Width: selected},
		{Title: "Domains", Width: domains},
	}

	rows := make([]table.Row, 0, len(networks))
	for _, n := range networks {
		sel := "No"
		if n.Selected {
			sel = "Yes"
		}
		domains := strings.Join(n.Domains, ", ")
		rows = append(rows, table.Row{
			n.ID,
			n.Range,
			sel,
			domains,
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
