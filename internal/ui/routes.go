package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
)

func buildRoutesTable(networks []*proto.Network, width, height int) table.Model {
	available := width - 14
	if available < 50 {
		available = 50
	}
	tableHeight := height - 14
	if tableHeight < 3 {
		tableHeight = 3
	}

	// Distribute: NetworkID=22%, CIDR=20%, Selected=10%, Domains=48%
	netID := available * 22 / 100
	cidr := available * 20 / 100
	selected := available * 10 / 100
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
	if domains < 12 {
		domains = 12
	}

	columns := []table.Column{
		{Title: "Network ID", Width: netID},
		{Title: "CIDR", Width: cidr},
		{Title: "Selected", Width: selected},
		{Title: "Domains", Width: domains},
	}

	rows := make([]table.Row, 0, len(networks))
	for _, n := range networks {
		sel := "○ No"
		if n.Selected {
			sel = "● Yes"
		}
		domainsStr := strings.Join(n.Domains, ", ")
		if domainsStr == "" {
			domainsStr = "—"
		}
		rows = append(rows, table.Row{
			n.ID,
			n.Range,
			sel,
			domainsStr,
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

// renderRoutesDetail renders detailed view of a selected network.
func renderRoutesDetail(n *proto.Network, width int) string {
	var sb strings.Builder

	lbl := lipgloss.NewStyle().Foreground(colorGray).Width(20)
	val := styleValue

	sb.WriteString(styleTitle.Render("Route Detail") + "\n\n")
	sb.WriteString(lbl.Render("Network ID:") + val.Render(n.ID) + "\n")
	sb.WriteString(lbl.Render("CIDR Range:") + val.Render(n.Range) + "\n")
	sb.WriteString(lbl.Render("Selected:"))
	if n.Selected {
		sb.WriteString(styleOnline.Render("● Yes"))
	} else {
		sb.WriteString(styleOffline.Render("○ No"))
	}
	sb.WriteString("\n")

	if len(n.Domains) > 0 {
		sb.WriteString(lbl.Render("Domains:") + "\n")
		for _, d := range n.Domains {
			sb.WriteString("  " + styleNeutral.Render("•") + " " + val.Render(d) + "\n")
		}
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

// routesHeader renders the routes tab header with summary.
func routesHeader(networks []*proto.Network) string {
	total := len(networks)
	selected := 0
	for _, n := range networks {
		if n.Selected {
			selected++
		}
	}
	return styleNeutral.Render(fmt.Sprintf("ROUTES  (%d selected / %d total)  — Enter to toggle", selected, total))
}
