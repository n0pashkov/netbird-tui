package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/netbirdio/netbird/client/proto"
)

type mapPeerRow struct {
	name    string
	ip      string
	path    string
	latency string
	network string
	style   lipgloss.Style
	rank    int
}

func renderNetworkMap(m *Model, maxWidth int) string {
	if m.status == nil || m.status.FullStatus == nil || m.status.FullStatus.LocalPeerState == nil {
		return emptyState("Network Map", "No local peer data available")
	}

	fs := m.status.FullStatus
	width := maxWidth
	if width <= 0 {
		width = m.width - 4
	}
	if width < 50 {
		width = 50
	}

	var sb strings.Builder
	metrics := mapMetrics(fs.Peers, m.networks)
	sb.WriteString(styleSectionHeader.Render("Network Map") + "  ")
	sb.WriteString(styleNeutral.Render(fmt.Sprintf("%d peer(s), %d route(s)", len(fs.Peers), len(m.networks))) + "\n\n")

	sb.WriteString(renderMapSummary(fs, m.networks, metrics, width))
	sb.WriteString("\n")
	sb.WriteString(renderPeerRows(fs.Peers, width))

	if relaySummary := renderRelaySummary(fs.Relays, width); relaySummary != "" {
		sb.WriteString("\n" + relaySummary)
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(fitContentWidth(sb.String(), width))
}

func mapMetrics(peers []*proto.PeerState, networks []*proto.Network) struct {
	online   int
	relayed  int
	offline  int
	selected int
} {
	var out struct {
		online   int
		relayed  int
		offline  int
		selected int
	}
	for _, p := range peers {
		if p.ConnStatus == "Connected" {
			out.online++
		} else {
			out.offline++
		}
		if p.Relayed {
			out.relayed++
		}
	}
	for _, n := range networks {
		if n.Selected {
			out.selected++
		}
	}
	return out
}

func renderMapSummary(fs *proto.FullStatus, networks []*proto.Network, metrics struct {
	online   int
	relayed  int
	offline  int
	selected int
}, width int) string {
	localName := firstNonEmpty(fs.LocalPeerState.Fqdn, "local peer")
	localIP := firstNonEmpty(fs.LocalPeerState.IP, "-")
	localW := clamp(width/3, 24, 42)
	pathW := clamp(width/4, 20, 34)
	destW := width - localW - pathW - 8
	if destW < 18 {
		destW = 18
	}

	local := []string{
		styleNeutral.Render("Local peer"),
		styleActiveTab.Render(ansi.Truncate(localName, localW, "")),
		styleNeutral.Render(ansi.Truncate(localIP, localW, "")),
	}
	paths := []string{
		styleNeutral.Render("Paths"),
		styleOnline.Render(fmt.Sprintf("p2p %d", metrics.online-metrics.relayed)),
		styleWarning.Render(fmt.Sprintf("relay %d", metrics.relayed)),
		styleNeutral.Render(fmt.Sprintf("offline %d", metrics.offline)),
	}
	routes := selectedRouteLabels(networks)
	if len(routes) == 0 {
		routes = []string{styleNeutral.Render("no selected routes")}
	}
	dest := append([]string{styleNeutral.Render("Destinations")}, routes...)

	rows := maxInt(len(local), len(paths), len(dest))
	var sb strings.Builder
	for i := 0; i < rows; i++ {
		connectorA := "   "
		connectorB := "   "
		if i == 1 {
			connectorA = styleOnline.Render("──▶")
			connectorB = styleOnline.Render("──▶")
		}
		sb.WriteString(padRight(lineAtPlain(local, i), localW))
		sb.WriteString(" " + connectorA + " ")
		sb.WriteString(padRight(lineAtPlain(paths, i), pathW))
		sb.WriteString(" " + connectorB + " ")
		sb.WriteString(ansi.Truncate(lineAtPlain(dest, i), destW, ""))
		sb.WriteString("\n")
	}
	return sb.String()
}

func renderPeerRows(peers []*proto.PeerState, width int) string {
	rows := mapPeerRows(peers)
	if len(rows) == 0 {
		return styleNeutral.Render("No peers to show") + "\n"
	}

	statusW := 8
	ipW := 15
	pathW := 10
	latencyW := 8
	nameW := width - statusW - ipW - pathW - latencyW - 8
	if nameW < 18 {
		nameW = 18
	}

	var sb strings.Builder
	sb.WriteString(styleNeutral.Render(padRight("State", statusW)))
	sb.WriteString("  " + styleNeutral.Render(padRight("Peer", nameW)))
	sb.WriteString("  " + styleNeutral.Render(padRight("IP", ipW)))
	sb.WriteString("  " + styleNeutral.Render(padRight("Path", pathW)))
	sb.WriteString("  " + styleNeutral.Render(padRight("Latency", latencyW)))
	sb.WriteString("\n")
	sb.WriteString(styleNeutral.Render(strings.Repeat("─", minInt(width, statusW+nameW+ipW+pathW+latencyW+8))) + "\n")

	for _, row := range rows {
		state := row.style.Render(padRight(row.pathState(), statusW))
		name := row.style.Render(padRight(row.name, nameW))
		ip := styleNeutral.Render(padRight(row.ip, ipW))
		path := row.style.Render(padRight(row.path, pathW))
		latency := styleNeutral.Render(padRight(row.latency, latencyW))
		sb.WriteString(state + "  " + name + "  " + ip + "  " + path + "  " + latency + "\n")
		if row.network != "" {
			sb.WriteString(styleNeutral.Render(padRight("", statusW+2)) + styleNeutral.Render(ansi.Truncate("↳ "+row.network, width-statusW-2, "")) + "\n")
		}
	}
	return sb.String()
}

func mapPeerRows(peers []*proto.PeerState) []mapPeerRow {
	rows := make([]mapPeerRow, 0, len(peers))
	for _, p := range peers {
		row := mapPeerRow{
			name:    firstNonEmpty(p.Fqdn, p.IP),
			ip:      firstNonEmpty(p.IP, "-"),
			path:    "offline",
			latency: "-",
			network: strings.Join(p.Networks, ", "),
			style:   styleNeutral,
			rank:    2,
		}
		if p.ConnStatus == "Connected" {
			row.path = "p2p"
			row.style = styleOnline
			row.rank = 0
			if p.Relayed {
				row.path = "relay"
				row.style = styleWarning
				row.rank = 1
			}
			if p.Latency != nil {
				d := p.Latency.AsDuration()
				if d > 0 {
					row.latency = fmt.Sprintf("%dms", d.Milliseconds())
				}
			}
		}
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].rank != rows[j].rank {
			return rows[i].rank < rows[j].rank
		}
		return rows[i].name < rows[j].name
	})
	return rows
}

func (r mapPeerRow) pathState() string {
	switch r.path {
	case "p2p":
		return "online"
	case "relay":
		return "relayed"
	default:
		return "offline"
	}
}

func selectedRouteLabels(networks []*proto.Network) []string {
	var labels []string
	for _, n := range networks {
		if !n.Selected {
			continue
		}
		label := firstNonEmpty(n.ID, n.Range)
		if n.Range != "" && n.Range != label {
			label += " " + n.Range
		}
		labels = append(labels, label)
	}
	return labels
}

func renderRelaySummary(relays []*proto.RelayState, width int) string {
	if len(relays) == 0 {
		return ""
	}
	parts := make([]string, 0, len(relays))
	for _, relay := range relays {
		marker := styleNeutral.Render("○")
		if relay.Available {
			marker = styleOnline.Render("●")
		}
		part := marker + " " + relay.URI
		if relay.Error != "" {
			part += " " + styleError.Render(relay.Error)
		}
		parts = append(parts, part)
	}
	return styleSectionHeader.Render("Relays") + "  " + ansi.Truncate(strings.Join(parts, "  "), width-8, "") + "\n"
}

func emptyState(title, message string) string {
	return lipgloss.NewStyle().Padding(1, 2).Render(styleSectionHeader.Render(title) + "\n\n" + styleNeutral.Render(message))
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return "-"
}

func padRight(s string, width int) string {
	if width <= 0 {
		return ""
	}
	s = ansi.Truncate(s, width, "")
	for ansi.StringWidth(s) < width {
		s += " "
	}
	return s
}

func lineAtPlain(lines []string, i int) string {
	if i < len(lines) {
		return lines[i]
	}
	return ""
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func maxInt(values ...int) int {
	max := 0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
