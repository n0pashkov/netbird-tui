package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
	"netbird-tui/internal/client"
)

type tab int

const (
	tabStatus tab = iota
	tabPeers
	tabRoutes
	tabFwdRules
	tabSettings
)

type tickMsg time.Time

type statusMsg struct {
	status *proto.StatusResponse
	err    error
}
type networksMsg struct {
	networks []*proto.Network
	err      error
}
type fwdRulesMsg struct {
	rules []*proto.ForwardingRule
	err   error
}
type upDownMsg struct {
	err error
}
type logoutMsg struct {
	err error
}
type debugBundleMsg struct {
	path string
	err  error
}
type loginMsg struct {
	err error
}
type toggleRouteMsg struct {
	err error
}
type configMsg struct {
	cfg *proto.GetConfigResponse
	err error
}

type Model struct {
	client          *client.Client
	activeTab       tab
	status          *proto.StatusResponse
	networks        []*proto.Network
	fwdRules        []*proto.ForwardingRule
	config          *proto.GetConfigResponse
	peersTable      table.Model
	routesTable     table.Model
	fwdTable        table.Model
	spinner         spinner.Model
	loading         bool
	err             error
	width           int
	height          int
	confirm         string // "up"/"down"/"logout"/"debug" pending confirmation
	lastAction      string
	setupKeyInput   textinput.Model
	mgmtURLInput    textinput.Model
	settingsFocused int  // 0=setupKey, 1=mgmtURL
	settingsEditing bool // false=browse, true=editing active field
	settingsMsg     string
	peerDetail      bool // true=showing peer detail view
}

func New(c *client.Client) *Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorBlue)

	ski := textinput.New()
	ski.Placeholder = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	ski.EchoMode = textinput.EchoPassword
	ski.CharLimit = 36
	// No Focus() here — browse mode by default

	mui := textinput.New()
	mui.Placeholder = "https://api.netbird.io"
	mui.CharLimit = 256

	return &Model{
		client:        c,
		spinner:       sp,
		loading:       true,
		setupKeyInput: ski,
		mgmtURLInput:  mui,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchStatus(),
		m.fetchNetworks(),
		m.fetchFwdRules(),
		m.fetchConfig(),
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) fetchStatus() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		resp, err := m.client.Status(ctx)
		return statusMsg{status: resp, err: err}
	}
}

func (m *Model) fetchNetworks() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		nets, err := m.client.ListNetworks(ctx)
		return networksMsg{networks: nets, err: err}
	}
}

func (m *Model) fetchFwdRules() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rules, err := m.client.ForwardingRules(ctx)
		return fwdRulesMsg{rules: rules, err: err}
	}
}

func (m *Model) fetchConfig() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cfg, err := m.client.GetConfig(ctx)
		return configMsg{cfg: cfg, err: err}
	}
}

func (m *Model) doUp() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := m.client.Up(ctx)
		return upDownMsg{err: err}
	}
}

func (m *Model) doDown() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := m.client.Down(ctx)
		return upDownMsg{err: err}
	}
}

func (m *Model) doLogout() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := m.client.Logout(ctx)
		return logoutMsg{err: err}
	}
}

func (m *Model) doDebugBundle() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		path, err := m.client.DebugBundle(ctx)
		return debugBundleMsg{path: path, err: err}
	}
}

func (m *Model) doLogin() tea.Cmd {
	setupKey := m.setupKeyInput.Value()
	mgmtURL := m.mgmtURLInput.Value()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := m.client.Login(ctx, setupKey, mgmtURL)
		return loginMsg{err: err}
	}
}

func (m *Model) doToggleRoute() tea.Cmd {
	row := m.routesTable.SelectedRow()
	if row == nil || len(m.networks) == 0 {
		return nil
	}
	// Find network by ID (first column)
	networkID := row[0]
	var selected bool
	for _, n := range m.networks {
		if n.ID == networkID {
			selected = n.Selected
			break
		}
	}
	ids := []string{networkID}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var err error
		if selected {
			err = m.client.DeselectNetworks(ctx, ids)
		} else {
			err = m.client.SelectNetworks(ctx, ids)
		}
		return toggleRouteMsg{err: err}
	}
}

func (m *Model) rebuildTables() {
	if m.status != nil && m.status.FullStatus != nil {
		m.peersTable = buildPeersTable(m.status.FullStatus.Peers, m.width, m.height)
	}
	if m.networks != nil {
		m.routesTable = buildRoutesTable(m.networks, m.width, m.height)
	}
	if m.fwdRules != nil {
		m.fwdTable = buildFwdTable(m.fwdRules, m.width, m.height)
	}
}

// isConnected returns true if management is currently connected.
func (m *Model) isConnected() bool {
	if m.status == nil || m.status.FullStatus == nil {
		return false
	}
	ms := m.status.FullStatus.ManagementState
	return ms != nil && ms.Connected
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.rebuildTables()

	case tea.KeyMsg:
		// Settings tab — browse/edit mode handling
		if m.activeTab == tabSettings && m.confirm == "" {
			if m.settingsEditing {
				// edit mode
				switch msg.String() {
				case "esc":
					m.settingsEditing = false
					m.setupKeyInput.Blur()
					m.mgmtURLInput.Blur()
					return m, nil
				case "ctrl+s":
					m.settingsMsg = ""
					m.loading = true
					return m, m.doLogin()
				case "ctrl+c":
					return m, tea.Quit
				default:
					var skiCmd, muiCmd tea.Cmd
					m.setupKeyInput, skiCmd = m.setupKeyInput.Update(msg)
					m.mgmtURLInput, muiCmd = m.mgmtURLInput.Update(msg)
					return m, tea.Batch(skiCmd, muiCmd)
				}
			} else {
				// browse mode
				switch msg.String() {
				case "up", "k":
					if m.settingsFocused > 0 {
						m.settingsFocused--
					}
					return m, nil
				case "down", "j":
					if m.settingsFocused < 1 {
						m.settingsFocused++
					}
					return m, nil
				case "enter":
					m.settingsEditing = true
					if m.settingsFocused == 0 {
						m.setupKeyInput.Focus()
					} else {
						m.mgmtURLInput.Focus()
					}
					return m, nil
				case "ctrl+s":
					m.settingsMsg = ""
					m.loading = true
					return m, m.doLogin()
				// fall through to global handler for everything else
				default:
					// handle global keys (tabs, quit, etc.) below
				}
			}
		}

		// Peer detail view
		if m.activeTab == tabPeers && m.peerDetail && m.confirm == "" {
			switch msg.String() {
			case "esc", "enter":
				m.peerDetail = false
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}

		// Handle confirmation prompt
		if m.confirm != "" {
			switch msg.String() {
			case "y", "Y":
				action := m.confirm
				m.confirm = ""
				m.loading = true
				switch action {
				case "up":
					return m, m.doUp()
				case "down":
					return m, m.doDown()
				case "logout":
					return m, m.doLogout()
				case "debug":
					return m, m.doDebugBundle()
				}
			default:
				m.confirm = ""
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "left":
			if m.activeTab > tabStatus {
				m.activeTab--
				m.settingsEditing = false
				m.setupKeyInput.Blur()
				m.mgmtURLInput.Blur()
			}
		case "right":
			if m.activeTab < tabSettings {
				m.activeTab++
				m.settingsEditing = false
				m.setupKeyInput.Blur()
				m.mgmtURLInput.Blur()
			}
		case "1":
			m.activeTab = tabStatus
		case "2":
			m.activeTab = tabPeers
		case "3":
			m.activeTab = tabRoutes
		case "4":
			m.activeTab = tabFwdRules
		case "5":
			m.activeTab = tabSettings
			m.settingsMsg = ""
			m.settingsEditing = false
			m.setupKeyInput.Blur()
			m.mgmtURLInput.Blur()
		case "r":
			m.loading = true
			cmds = append(cmds, m.fetchStatus(), m.fetchNetworks(), m.fetchFwdRules())
		case "c":
			if m.isConnected() {
				m.confirm = "down"
			} else {
				m.confirm = "up"
			}
			return m, nil
		case "u":
			m.confirm = "up"
			return m, nil
		case "d":
			m.confirm = "down"
			return m, nil
		case "L":
			m.confirm = "logout"
			return m, nil
		case "b":
			m.confirm = "debug"
			return m, nil
		case "up", "k":
			if m.activeTab == tabPeers {
				m.peersTable, _ = m.peersTable.Update(msg)
			} else if m.activeTab == tabRoutes {
				m.routesTable, _ = m.routesTable.Update(msg)
			} else if m.activeTab == tabFwdRules {
				m.fwdTable, _ = m.fwdTable.Update(msg)
			}
		case "down", "j":
			if m.activeTab == tabPeers {
				m.peersTable, _ = m.peersTable.Update(msg)
			} else if m.activeTab == tabRoutes {
				m.routesTable, _ = m.routesTable.Update(msg)
			} else if m.activeTab == tabFwdRules {
				m.fwdTable, _ = m.fwdTable.Update(msg)
			}
		case "enter":
			if m.activeTab == tabRoutes {
				cmd := m.doToggleRoute()
				if cmd != nil {
					return m, cmd
				}
			} else if m.activeTab == tabPeers {
				m.peerDetail = true
				return m, nil
			}
		}

	case tickMsg:
		cmds = append(cmds, m.fetchStatus(), m.fetchNetworks(), m.fetchFwdRules(), tickCmd())

	case statusMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
			m.status = msg.status
			if m.status != nil && m.status.FullStatus != nil {
				m.peersTable = buildPeersTable(m.status.FullStatus.Peers, m.width, m.height)
			}
		}

	case networksMsg:
		if msg.err == nil {
			m.networks = msg.networks
			m.routesTable = buildRoutesTable(m.networks, m.width, m.height)
		}

	case fwdRulesMsg:
		if msg.err == nil {
			m.fwdRules = msg.rules
			m.fwdTable = buildFwdTable(m.fwdRules, m.width, m.height)
		}

	case configMsg:
		if msg.err == nil && msg.cfg != nil {
			m.config = msg.cfg
			if m.mgmtURLInput.Value() == "" {
				m.mgmtURLInput.SetValue(msg.cfg.ManagementUrl)
			}
		}

	case upDownMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.lastAction = "Done"
			cmds = append(cmds, m.fetchStatus())
		}

	case logoutMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.lastAction = "Logged out"
			cmds = append(cmds, m.fetchStatus())
		}

	case debugBundleMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.lastAction = "Debug bundle: " + msg.path
		}

	case loginMsg:
		m.loading = false
		if msg.err != nil {
			m.settingsMsg = "Error: " + msg.err.Error()
		} else {
			m.settingsMsg = "Login successful"
			m.setupKeyInput.SetValue("")
			cmds = append(cmds, m.fetchStatus())
		}

	case toggleRouteMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			cmds = append(cmds, m.fetchNetworks())
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var sections []string

	sections = append(sections, m.renderHeader())
	sections = append(sections, m.renderTabBar())

	contentHeight := m.height - 8
	if contentHeight < 5 {
		contentHeight = 5
	}

	content := m.renderContent()
	contentStyle := lipgloss.NewStyle().
		Height(contentHeight).
		Width(m.width - 2)
	sections = append(sections, contentStyle.Render(content))

	if m.confirm != "" {
		sections = append(sections, m.renderConfirm())
	} else {
		sections = append(sections, m.renderFooter())
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) renderHeader() string {
	title := styleTitle.Render("NetBird TUI")

	status := ""
	if m.loading {
		status = m.spinner.View() + " Connecting..."
	} else if m.err != nil {
		status = styleOffline.Render("● Error: " + m.err.Error())
	} else if m.status != nil && m.status.FullStatus != nil {
		fs := m.status.FullStatus
		if fs.LocalPeerState != nil {
			ip := fs.LocalPeerState.IP
			fqdn := fs.LocalPeerState.Fqdn
			mgmt := fs.ManagementState
			if mgmt != nil && mgmt.Connected {
				status = styleOnline.Render("● Connected") + styleNeutral.Render("  "+ip+"  "+fqdn)
			} else {
				status = styleOffline.Render("○ Disconnected") + styleNeutral.Render("  "+ip)
			}
		}
	}

	headerContent := lipgloss.JoinHorizontal(lipgloss.Center,
		title,
		styleNeutral.Render("  │  "),
		status,
	)

	return styleHeader.Width(m.width - 2).Render(headerContent)
}

func (m *Model) renderTabBar() string {
	tabs := []struct {
		id    tab
		label string
	}{
		{tabStatus, "[1] Status"},
		{tabPeers, "[2] Peers"},
		{tabRoutes, "[3] Routes"},
		{tabFwdRules, "[4] Forwarding"},
		{tabSettings, "[5] Settings"},
	}

	var parts []string
	for _, t := range tabs {
		if m.activeTab == t.id {
			parts = append(parts, styleActiveTab.Render(t.label))
		} else {
			parts = append(parts, styleInactiveTab.Render(t.label))
		}
	}

	return styleNeutral.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		BorderBottom(true).
		Width(m.width - 2).
		Render(strings.Join(parts, "  "))
}

func (m *Model) renderContent() string {
	switch m.activeTab {
	case tabStatus:
		return renderStatus(m)
	case tabPeers:
		if m.status == nil || m.status.FullStatus == nil {
			return styleNeutral.Padding(1, 2).Render("No peer data available")
		}
		if m.peerDetail {
			row := m.peersTable.SelectedRow()
			if row != nil {
				// Find peer by FQDN (first column) or IP (second column)
				for _, p := range m.status.FullStatus.Peers {
					if p.Fqdn == row[0] || p.IP == row[1] {
						return renderPeerDetail(p, m.width)
					}
				}
			}
			m.peerDetail = false
		}
		header := peersHeader(m.status.FullStatus.Peers)
		return lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Padding(0, 2).Render(header),
			lipgloss.NewStyle().Padding(0, 2).Render(m.peersTable.View()),
		)
	case tabRoutes:
		if len(m.networks) == 0 {
			return styleNeutral.Padding(1, 2).Render("No routes available")
		}
		return lipgloss.NewStyle().Padding(0, 2).Render(m.routesTable.View())
	case tabFwdRules:
		if len(m.fwdRules) == 0 {
			return styleNeutral.Padding(1, 2).Render("No forwarding rules")
		}
		return lipgloss.NewStyle().Padding(0, 2).Render(m.fwdTable.View())
	case tabSettings:
		return renderSettings(m)
	}
	return ""
}

func (m *Model) renderFooter() string {
	var help string
	switch m.activeTab {
	case tabRoutes:
		help = "Enter:Toggle  c:Toggle  u:Up  d:Down  L:Logout  b:Debug  r:Refresh  ←→/1-5:Tabs  ↑↓:Nav  q:Quit"
	case tabSettings:
		if m.settingsEditing {
			help = "Esc:Stop editing  ctrl+s:Submit"
		} else {
			help = "↑↓:Select  Enter:Edit  ctrl+s:Submit  ←→/1-5:Tabs  q:Quit"
		}
	case tabPeers:
		if m.peerDetail {
			help = "Esc/Enter:Back to list"
		} else {
			help = "Enter:Detail  c:Toggle  u:Up  d:Down  L:Logout  b:Debug  r:Refresh  ←→/1-5:Tabs  ↑↓:Nav  q:Quit"
		}
	default:
		help = "c:Toggle  u:Up  d:Down  L:Logout  b:Debug  r:Refresh  ←→/1-5:Tabs  ↑↓:Nav  q:Quit"
	}
	if m.lastAction != "" {
		help = fmt.Sprintf("Last: %s  |  %s", m.lastAction, help)
	}
	return styleFooter.Width(m.width - 2).Render(help)
}

func (m *Model) renderConfirm() string {
	action := m.confirm
	var actionStr string
	switch action {
	case "up":
		actionStr = "netbird up"
	case "down":
		actionStr = "netbird down"
	case "logout":
		actionStr = "logout (disconnect and delete peer)"
	case "debug":
		actionStr = "create debug bundle"
	default:
		actionStr = action
	}
	msg := fmt.Sprintf("Run '%s'? Press y to confirm, any other key to cancel.", actionStr)
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorYellow).
		Padding(0, 2).
		Foreground(colorYellow).
		Width(m.width - 2).
		Render(msg)
}
