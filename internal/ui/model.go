package ui

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/netbirdio/netbird/client/proto"
	"netbird-tui/internal/client"
)

type tab int

const (
	tabStatus      tab = iota // 1
	tabPeers                  // 2
	tabRoutes                 // 3
	tabDNS                    // 4
	tabEvents                 // 5
	tabProfiles               // 6
	tabDiagnostics            // 7
	tabServices               // 8
	tabFwdRules               // 9
	tabSettings               // 0
	tabNetMap
)

const tabCount = 11

type tabGroup int

const (
	groupMonitor tabGroup = iota
	groupNetwork
	groupManage
	groupTools
)

type tabDef struct {
	id    tab
	group tabGroup
	name  string
	key   string
}

var tabGroups = []struct {
	id   tabGroup
	name string
	key  string
	tabs []tab
}{
	{groupMonitor, "Monitor", "1", []tab{tabStatus, tabPeers, tabEvents, tabNetMap}},
	{groupNetwork, "Network", "2", []tab{tabRoutes, tabDNS, tabFwdRules}},
	{groupManage, "Manage", "3", []tab{tabProfiles, tabSettings}},
	{groupTools, "Tools", "4", []tab{tabDiagnostics, tabServices}},
}

var tabs = []tabDef{
	{tabStatus, groupMonitor, "Status", "1"},
	{tabPeers, groupMonitor, "Peers", "2"},
	{tabEvents, groupMonitor, "Events", "3"},
	{tabNetMap, groupMonitor, "Map", "m"},
	{tabRoutes, groupNetwork, "Routes", "4"},
	{tabDNS, groupNetwork, "DNS", "5"},
	{tabFwdRules, groupNetwork, "Forwarding", "6"},
	{tabProfiles, groupManage, "Profiles", "7"},
	{tabSettings, groupManage, "Settings", "8"},
	{tabDiagnostics, groupTools, "Diagnostics", "9"},
	{tabServices, groupTools, "Services", "0"},
}

type peerFilterMode int

const (
	peerFilterAll peerFilterMode = iota
	peerFilterOnline
	peerFilterOffline
	peerFilterRelayed
)

// ─── Message types ─────────────────────────────────────────────────────────────

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
type upDownMsg struct{ err error }
type logoutMsg struct{ err error }
type debugBundleMsg struct {
	path string
	err  error
}
type loginMsg struct{ err error }
type toggleRouteMsg struct{ err error }
type configMsg struct {
	cfg *proto.GetConfigResponse
	err error
}

// New message types
type eventsMsg struct {
	events []*proto.SystemEvent
	err    error
}
type profilesMsg struct {
	profiles      []*proto.Profile
	activeProfile string
	err           error
}
type switchProfileMsg struct{ err error }
type addProfileMsg struct{ err error }
type removeProfileMsg struct{ err error }
type logLevelMsg struct {
	level proto.LogLevel
	err   error
}
type setLogLevelMsg struct{ err error }
type statesMsg struct {
	states []*proto.State
	err    error
}
type cleanStateMsg struct{ err error }
type deleteStateMsg struct{ err error }
type tracePacketMsg struct {
	resp *proto.TracePacketResponse
	err  error
}
type featuresMsg struct {
	features *proto.GetFeaturesResponse
	err      error
}
type sshHostKeyMsg struct {
	key string
	err error
}
type syncPersistenceMsg struct{ err error }

// ─── Diagnostics sub-modes ────────────────────────────────────────────────────

type diagMode int

const (
	diagModeOverview diagMode = iota
	diagModeTrace
	diagModeStates
)

// ─── Model ────────────────────────────────────────────────────────────────────

type Model struct {
	client    *client.Client
	activeTab tab
	width     int
	height    int

	// Spinner / loading
	spinner     spinner.Model
	loading     bool
	loadingWhat string

	// Errors / messages
	err        error
	lastAction string

	// Confirmation
	confirm     string // pending confirmation action
	quickSwitch bool
	helpOverlay bool

	// ── Status tab ──
	status   *proto.StatusResponse
	features *proto.GetFeaturesResponse

	// ── Peers tab ──
	peersTable     table.Model
	peerDetail     bool
	peerSSHKey     string
	peerSSHKeyErr  string
	peersSearch    textinput.Model
	peersSearching bool
	peersFilter    peerFilterMode

	// ── Routes tab ──
	networks    []*proto.Network
	routesTable table.Model

	// ── DNS tab ──
	// Data comes from status.FullStatus.DnsServers
	dnsSelected int

	// ── Events tab ──
	events          []*proto.SystemEvent
	eventsTable     table.Model
	eventsFilter    proto.SystemEvent_Severity // -1 = all (since 0 == INFO)
	eventsSearch    textinput.Model
	eventsSearching bool
	eventsDetail    bool
	eventsScroll    int

	// ── Profiles tab ──
	profiles         []*proto.Profile
	activeProfile    string
	profilesTable    table.Model
	profilesEditing  bool
	profileNameInput textinput.Model
	profileMgmtInput textinput.Model
	profilesFocused  int // 0=name, 1=mgmtURL
	profilesMsg      string

	// ── Diagnostics tab ──
	diagMode      diagMode
	logLevel      proto.LogLevel
	logLevelKnown bool
	states        []*proto.State
	statesTable   table.Model
	statesFocused int
	// Trace packet form
	traceSrcIP   textinput.Model
	traceDstIP   textinput.Model
	traceProto   textinput.Model
	traceSrcPort textinput.Model
	traceDstPort textinput.Model
	traceDir     textinput.Model
	traceFocused int
	traceResult  *proto.TracePacketResponse
	traceErr     string
	traceEditing bool

	// ── Services tab ──
	fwdRules []*proto.ForwardingRule
	fwdTable table.Model

	// ── Settings tab ──
	config          *proto.GetConfigResponse
	setupKeyInput   textinput.Model
	mgmtURLInput    textinput.Model
	settingsFocused int
	settingsEditing bool
	settingsMsg     string
	// Extended settings
	settingsPage int // 0=login, 1=config flags
}

func New(c *client.Client) *Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorBlue)

	// Peers search
	ps := textinput.New()
	ps.Placeholder = "search by FQDN or IP…"
	ps.CharLimit = 64

	es := textinput.New()
	es.Placeholder = "search events…"
	es.CharLimit = 96

	ski := textinput.New()
	ski.Placeholder = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	ski.EchoMode = textinput.EchoPassword
	ski.CharLimit = 36

	mui := textinput.New()
	mui.Placeholder = "https://api.netbird.io"
	mui.CharLimit = 256

	// Profile inputs
	pni := textinput.New()
	pni.Placeholder = "profile-name"
	pni.CharLimit = 64

	pmi := textinput.New()
	pmi.Placeholder = "https://api.netbird.io"
	pmi.CharLimit = 256

	// Trace inputs
	tSrc := textinput.New()
	tSrc.Placeholder = "100.64.0.1"
	tSrc.CharLimit = 45

	tDst := textinput.New()
	tDst.Placeholder = "100.64.0.2"
	tDst.CharLimit = 45

	tProto := textinput.New()
	tProto.Placeholder = "tcp"
	tProto.CharLimit = 10

	tSPort := textinput.New()
	tSPort.Placeholder = "0"
	tSPort.CharLimit = 5

	tDPort := textinput.New()
	tDPort.Placeholder = "80"
	tDPort.CharLimit = 5

	tDir := textinput.New()
	tDir.Placeholder = "in"
	tDir.CharLimit = 10

	return &Model{
		client:           c,
		spinner:          sp,
		loading:          true,
		loadingWhat:      "Connecting",
		setupKeyInput:    ski,
		mgmtURLInput:     mui,
		profileNameInput: pni,
		profileMgmtInput: pmi,
		traceSrcIP:       tSrc,
		traceDstIP:       tDst,
		traceProto:       tProto,
		traceSrcPort:     tSPort,
		traceDstPort:     tDPort,
		traceDir:         tDir,
		eventsFilter:     -1,
		eventsSearch:     es,
		peersSearch:      ps,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchStatus(),
		m.fetchNetworks(),
		m.fetchFwdRules(),
		m.fetchConfig(),
		m.fetchEvents(),
		m.fetchProfiles(),
		m.fetchLogLevel(),
		m.fetchStates(),
		m.fetchFeatures(),
		tickCmd(),
	)
}

// ─── Ticker ───────────────────────────────────────────────────────────────────

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ─── Fetch commands ───────────────────────────────────────────────────────────

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

func (m *Model) fetchEvents() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		events, err := m.client.GetEvents(ctx)
		return eventsMsg{events: events, err: err}
	}
}

func (m *Model) fetchProfiles() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		profiles, err := m.client.ListProfiles(ctx)
		if err != nil {
			return profilesMsg{err: err}
		}
		active, _ := m.client.GetActiveProfile(ctx)
		return profilesMsg{profiles: profiles, activeProfile: active}
	}
}

func (m *Model) fetchLogLevel() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		level, err := m.client.GetLogLevel(ctx)
		return logLevelMsg{level: level, err: err}
	}
}

func (m *Model) fetchStates() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		states, err := m.client.ListStates(ctx)
		return statesMsg{states: states, err: err}
	}
}

func (m *Model) fetchFeatures() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		feat, err := m.client.GetFeatures(ctx)
		return featuresMsg{features: feat, err: err}
	}
}

// ─── Action commands ──────────────────────────────────────────────────────────

func (m *Model) doUp() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return upDownMsg{err: m.client.Up(ctx)}
	}
}

func (m *Model) doDown() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return upDownMsg{err: m.client.Down(ctx)}
	}
}

func (m *Model) doLogout() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return logoutMsg{err: m.client.Logout(ctx)}
	}
}

func (m *Model) doDebugBundle(anonymize, sysInfo bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		path, err := m.client.DebugBundle(ctx, anonymize, sysInfo)
		return debugBundleMsg{path: path, err: err}
	}
}

func (m *Model) doLogin() tea.Cmd {
	setupKey := m.setupKeyInput.Value()
	mgmtURL := m.mgmtURLInput.Value()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return loginMsg{err: m.client.Login(ctx, setupKey, mgmtURL)}
	}
}

func (m *Model) doToggleRoute() tea.Cmd {
	row := m.routesTable.SelectedRow()
	if row == nil || len(m.networks) == 0 {
		return nil
	}
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

func (m *Model) doSwitchProfile(name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return switchProfileMsg{err: m.client.SwitchProfile(ctx, name)}
	}
}

func (m *Model) doAddProfile() tea.Cmd {
	name := m.profileNameInput.Value()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return addProfileMsg{err: m.client.AddProfile(ctx, name)}
	}
}

func (m *Model) doRemoveProfile(name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return removeProfileMsg{err: m.client.RemoveProfile(ctx, name)}
	}
}

func (m *Model) doSetLogLevel(level proto.LogLevel) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return setLogLevelMsg{err: m.client.SetLogLevel(ctx, level)}
	}
}

func (m *Model) doCleanState(name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return cleanStateMsg{err: m.client.CleanState(ctx, name)}
	}
}

func (m *Model) doDeleteState(name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return deleteStateMsg{err: m.client.DeleteState(ctx, name)}
	}
}

func (m *Model) doTracePacket() tea.Cmd {
	if err := validateTraceInput(
		m.traceSrcIP.Value(),
		m.traceDstIP.Value(),
		m.traceProto.Value(),
		m.traceSrcPort.Value(),
		m.traceDstPort.Value(),
		m.traceDir.Value(),
	); err != nil {
		m.loading = false
		m.traceErr = err.Error()
		return nil
	}
	srcIP := m.traceSrcIP.Value()
	dstIP := m.traceDstIP.Value()
	proto_ := m.traceProto.Value()
	dir := m.traceDir.Value()

	srcPort := uint32(0)
	fmt.Sscanf(m.traceSrcPort.Value(), "%d", &srcPort)
	dstPort := uint32(0)
	fmt.Sscanf(m.traceDstPort.Value(), "%d", &dstPort)

	if proto_ == "" {
		proto_ = "tcp"
	}
	if dir == "" {
		dir = "in"
	}

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := m.client.TracePacket(ctx, &proto.TracePacketRequest{
			SourceIp:        srcIP,
			DestinationIp:   dstIP,
			Protocol:        proto_,
			SourcePort:      srcPort,
			DestinationPort: dstPort,
			Direction:       dir,
		})
		return tracePacketMsg{resp: resp, err: err}
	}
}

func validateTraceInput(srcIP, dstIP, protoName, srcPort, dstPort, dir string) error {
	if net.ParseIP(strings.TrimSpace(srcIP)) == nil {
		return fmt.Errorf("source IP is invalid")
	}
	if net.ParseIP(strings.TrimSpace(dstIP)) == nil {
		return fmt.Errorf("destination IP is invalid")
	}
	protoName = strings.ToLower(strings.TrimSpace(protoName))
	if protoName == "" {
		protoName = "tcp"
	}
	if protoName != "tcp" && protoName != "udp" && protoName != "icmp" {
		return fmt.Errorf("protocol must be tcp, udp, or icmp")
	}
	dir = strings.ToLower(strings.TrimSpace(dir))
	if dir == "" {
		dir = "in"
	}
	if dir != "in" && dir != "out" {
		return fmt.Errorf("direction must be in or out")
	}
	for _, item := range []struct {
		name  string
		value string
	}{
		{"source port", srcPort},
		{"destination port", dstPort},
	} {
		if strings.TrimSpace(item.value) == "" {
			continue
		}
		p, err := strconv.Atoi(item.value)
		if err != nil || p < 0 || p > 65535 {
			return fmt.Errorf("%s must be 0-65535", item.name)
		}
	}
	return nil
}

func (m *Model) doGetSSHHostKey(peerAddr string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		key, err := m.client.GetPeerSSHHostKey(ctx, peerAddr)
		return sshHostKeyMsg{key: key, err: err}
	}
}

// ─── Table builders ───────────────────────────────────────────────────────────

func (m *Model) rebuildTables() {
	if m.status != nil && m.status.FullStatus != nil {
		m.rebuildPeersTableKeepingSelection()
	}
	if m.networks != nil {
		m.routesTable = buildRoutesTable(m.networks, m.width, m.height)
	}
	if m.fwdRules != nil {
		m.fwdTable = buildFwdTable(m.fwdRules, m.width, m.height)
	}
	if m.events != nil {
		m.eventsTable = buildEventsTable(m.events, m.eventsFilter, m.eventsSearch.Value(), m.width, m.height)
	}
	if m.profiles != nil {
		m.profilesTable = buildProfilesTable(m.profiles, m.activeProfile, m.width, m.height)
	}
	if m.states != nil {
		m.statesTable = buildStatesTable(m.states, m.width, m.height)
	}
}

func (m *Model) filteredPeers() []*proto.PeerState {
	if m.status == nil || m.status.FullStatus == nil {
		return nil
	}
	return filterPeers(m.status.FullStatus.Peers, m.peersSearch.Value(), m.peersFilter)
}

func (m *Model) selectedPeer() *proto.PeerState {
	peers := m.filteredPeers()
	idx := m.peersTable.Cursor()
	if idx < 0 || idx >= len(peers) {
		return nil
	}
	return peers[idx]
}

func (m *Model) rebuildPeersTableKeepingSelection() {
	selectedKey := ""
	if peer := m.selectedPeer(); peer != nil {
		selectedKey = peer.Fqdn + "\x00" + peer.IP
	}
	peers := m.filteredPeers()
	m.peersTable = buildPeersTable(peers, m.width, m.height)
	if selectedKey == "" {
		return
	}
	for i, p := range peers {
		if p.Fqdn+"\x00"+p.IP == selectedKey {
			m.peersTable.SetCursor(i)
			return
		}
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

// ─── Update ───────────────────────────────────────────────────────────────────

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.rebuildTables()
		cmds = append(cmds, tea.ClearScreen)

	case tea.KeyMsg:
		// Global quit always works
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Route keys to specific tab handlers
		cmd := m.handleTabKey(msg)
		if cmd != nil {
			return m, cmd
		}

	case tickMsg:
		cmds = append(cmds, m.fetchStatus(), m.fetchNetworks(), m.fetchFwdRules(), m.fetchEvents(), tickCmd())

	case statusMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.status = msg.status
			if m.status != nil && m.status.FullStatus != nil {
				m.rebuildPeersTableKeepingSelection()
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

	case eventsMsg:
		if msg.err == nil {
			m.events = msg.events
			m.eventsTable = buildEventsTable(m.events, m.eventsFilter, m.eventsSearch.Value(), m.width, m.height)
		}

	case profilesMsg:
		if msg.err == nil {
			m.profiles = msg.profiles
			m.activeProfile = msg.activeProfile
			m.profilesTable = buildProfilesTable(m.profiles, m.activeProfile, m.width, m.height)
		}

	case logLevelMsg:
		if msg.err == nil {
			m.logLevel = msg.level
			m.logLevelKnown = true
		}

	case statesMsg:
		if msg.err == nil {
			m.states = msg.states
			m.statesTable = buildStatesTable(m.states, m.width, m.height)
		}

	case featuresMsg:
		if msg.err == nil {
			m.features = msg.features
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

	case switchProfileMsg:
		m.loading = false
		if msg.err != nil {
			m.profilesMsg = "Error: " + msg.err.Error()
		} else {
			m.profilesMsg = "Profile switched"
			cmds = append(cmds, m.fetchProfiles(), m.fetchStatus())
		}

	case addProfileMsg:
		m.loading = false
		if msg.err != nil {
			m.profilesMsg = "Error: " + msg.err.Error()
		} else {
			m.profilesMsg = "Profile added"
			m.profileNameInput.SetValue("")
			m.profileMgmtInput.SetValue("")
			m.profilesEditing = false
			cmds = append(cmds, m.fetchProfiles())
		}

	case removeProfileMsg:
		m.loading = false
		if msg.err != nil {
			m.profilesMsg = "Error: " + msg.err.Error()
		} else {
			m.profilesMsg = "Profile removed"
			cmds = append(cmds, m.fetchProfiles())
		}

	case setLogLevelMsg:
		m.loading = false
		if msg.err != nil {
			m.lastAction = "Error setting log level: " + msg.err.Error()
		} else {
			m.lastAction = "Log level updated"
			cmds = append(cmds, m.fetchLogLevel())
		}

	case cleanStateMsg:
		m.loading = false
		if msg.err != nil {
			m.lastAction = "Error cleaning state: " + msg.err.Error()
		} else {
			m.lastAction = "State cleaned"
			cmds = append(cmds, m.fetchStates())
		}

	case deleteStateMsg:
		m.loading = false
		if msg.err != nil {
			m.lastAction = "Error deleting state: " + msg.err.Error()
		} else {
			m.lastAction = "State deleted"
			cmds = append(cmds, m.fetchStates())
		}

	case tracePacketMsg:
		m.loading = false
		m.traceResult = msg.resp
		if msg.err != nil {
			m.traceErr = msg.err.Error()
		} else {
			m.traceErr = ""
		}

	case sshHostKeyMsg:
		if msg.err != nil {
			m.peerSSHKeyErr = msg.err.Error()
		} else {
			m.peerSSHKey = msg.key
			m.peerSSHKeyErr = ""
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleTabKey dispatches keyboard input to the correct handler per tab/mode.
func (m *Model) handleTabKey(msg tea.KeyMsg) tea.Cmd {
	// Confirmation overlay takes priority
	if m.confirm != "" {
		return m.handleConfirm(msg)
	}
	if m.helpOverlay {
		return m.handleHelpOverlay(msg)
	}
	if m.quickSwitch {
		return m.handleQuickSwitch(msg)
	}

	// Tab-specific handlers
	switch m.activeTab {
	case tabSettings:
		return m.handleSettingsKey(msg)
	case tabPeers:
		return m.handlePeersKey(msg)
	case tabProfiles:
		return m.handleProfilesKey(msg)
	case tabDiagnostics:
		return m.handleDiagnosticsKey(msg)
	case tabEvents:
		return m.handleEventsKey(msg)
	case tabServices:
		return m.handleServicesKey(msg)
	}

	// Global key handling
	return m.handleGlobalKey(msg)
}

func (m *Model) handleConfirm(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		action := m.confirm
		m.confirm = ""
		m.loading = true
		switch action {
		case "up":
			return m.doUp()
		case "down":
			return m.doDown()
		case "logout":
			return m.doLogout()
		case "debug":
			return m.doDebugBundle(false, true)
		case "debug-anon":
			return m.doDebugBundle(true, true)
		case "clean-state":
			if m.statesFocused < len(m.states) {
				return m.doCleanState(m.states[m.statesFocused].Name)
			}
		case "delete-state":
			if m.statesFocused < len(m.states) {
				return m.doDeleteState(m.states[m.statesFocused].Name)
			}
		case "remove-profile":
			row := m.profilesTable.SelectedRow()
			if row != nil {
				return m.doRemoveProfile(row[0])
			}
		case "switch-profile":
			row := m.profilesTable.SelectedRow()
			if row != nil {
				return m.doSwitchProfile(row[0])
			}
		}
	default:
		m.confirm = ""
	}
	return nil
}

func (m *Model) handleHelpOverlay(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		return tea.Quit
	default:
		m.helpOverlay = false
	}
	return nil
}

func (m *Model) handleQuickSwitch(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		return tea.Quit
	case "esc", "g", "q":
		m.quickSwitch = false
	case "1", "2", "3", "4":
		m.quickSwitch = false
		m.setActiveGroup(tabGroups[int(msg.String()[0]-'1')].id)
	case "5", "6", "7", "8", "9", "0":
		for _, td := range tabs {
			if td.key == msg.String() {
				m.quickSwitch = false
				m.setActiveTab(td.id)
				break
			}
		}
	case "m":
		m.quickSwitch = false
		m.setActiveTab(tabNetMap)
	}
	return nil
}

func (m *Model) activeGroup() tabGroup {
	for _, td := range tabs {
		if td.id == m.activeTab {
			return td.group
		}
	}
	return groupMonitor
}

func (m *Model) activeGroupIndex() int {
	active := m.activeGroup()
	for i, g := range tabGroups {
		if g.id == active {
			return i
		}
	}
	return 0
}

func (m *Model) switchGroup(delta int) {
	idx := m.activeGroupIndex()
	idx = (idx + delta + len(tabGroups)) % len(tabGroups)
	m.setActiveGroup(tabGroups[idx].id)
}

func (m *Model) setActiveGroup(group tabGroup) {
	for _, g := range tabGroups {
		if g.id == group && len(g.tabs) > 0 {
			m.setActiveTab(g.tabs[0])
			return
		}
	}
}

func (m *Model) switchTabInGroup(delta int) {
	for _, g := range tabGroups {
		if g.id != m.activeGroup() {
			continue
		}
		idx := 0
		for i, t := range g.tabs {
			if t == m.activeTab {
				idx = i
				break
			}
		}
		idx = (idx + delta + len(g.tabs)) % len(g.tabs)
		m.setActiveTab(g.tabs[idx])
		return
	}
}

func tabName(t tab) string {
	for _, td := range tabs {
		if td.id == t {
			return td.name
		}
	}
	return ""
}

func (m *Model) handleGlobalKey(msg tea.KeyMsg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg.String() {
	case "q":
		return tea.Quit
	case "tab":
		m.switchGroup(1)
	case "shift+tab":
		m.switchGroup(-1)
	case "left":
		m.switchTabInGroup(-1)
	case "right":
		m.switchTabInGroup(1)
	case "1":
		m.setActiveGroup(groupMonitor)
	case "2":
		m.setActiveGroup(groupNetwork)
	case "3":
		m.setActiveGroup(groupManage)
	case "4":
		m.setActiveGroup(groupTools)
	case "g":
		m.quickSwitch = true
	case "?":
		m.helpOverlay = true
	case "r":
		m.loading = true
		m.err = nil
		cmds = append(cmds, m.fetchStatus(), m.fetchNetworks(), m.fetchFwdRules(), m.fetchEvents(), m.fetchProfiles(), m.fetchStates())
	case "c":
		if m.isConnected() {
			m.confirm = "down"
		} else {
			m.confirm = "up"
		}
	case "u":
		m.confirm = "up"
	case "d":
		m.confirm = "down"
	case "L":
		m.confirm = "logout"
	case "b":
		m.confirm = "debug"
	case "up", "k":
		m.navigateUp()
	case "down", "j":
		m.navigateDown()
	case "enter":
		return m.handleEnter()
	}

	return tea.Batch(cmds...)
}

func (m *Model) navigateUp() {
	switch m.activeTab {
	case tabPeers:
		m.peersTable, _ = m.peersTable.Update(tea.KeyMsg{Type: tea.KeyUp})
	case tabRoutes:
		m.routesTable, _ = m.routesTable.Update(tea.KeyMsg{Type: tea.KeyUp})
	case tabFwdRules:
		m.fwdTable, _ = m.fwdTable.Update(tea.KeyMsg{Type: tea.KeyUp})
	case tabEvents:
		if m.eventsDetail {
			if m.eventsScroll > 0 {
				m.eventsScroll--
			}
		} else {
			m.eventsTable, _ = m.eventsTable.Update(tea.KeyMsg{Type: tea.KeyUp})
		}
	case tabProfiles:
		m.profilesTable, _ = m.profilesTable.Update(tea.KeyMsg{Type: tea.KeyUp})
	case tabDiagnostics:
		if m.diagMode == diagModeStates {
			m.statesTable, _ = m.statesTable.Update(tea.KeyMsg{Type: tea.KeyUp})
		}
	case tabDNS:
		if m.dnsSelected > 0 {
			m.dnsSelected--
		}
	}
}

func (m *Model) navigateDown() {
	switch m.activeTab {
	case tabPeers:
		m.peersTable, _ = m.peersTable.Update(tea.KeyMsg{Type: tea.KeyDown})
	case tabRoutes:
		m.routesTable, _ = m.routesTable.Update(tea.KeyMsg{Type: tea.KeyDown})
	case tabFwdRules:
		m.fwdTable, _ = m.fwdTable.Update(tea.KeyMsg{Type: tea.KeyDown})
	case tabEvents:
		if m.eventsDetail {
			m.eventsScroll++
		} else {
			m.eventsTable, _ = m.eventsTable.Update(tea.KeyMsg{Type: tea.KeyDown})
		}
	case tabProfiles:
		m.profilesTable, _ = m.profilesTable.Update(tea.KeyMsg{Type: tea.KeyDown})
	case tabDiagnostics:
		if m.diagMode == diagModeStates {
			m.statesTable, _ = m.statesTable.Update(tea.KeyMsg{Type: tea.KeyDown})
		}
	case tabDNS:
		if m.status != nil && m.status.FullStatus != nil {
			if m.dnsSelected < len(m.status.FullStatus.DnsServers)-1 {
				m.dnsSelected++
			}
		}
	}
}

func (m *Model) handleEnter() tea.Cmd {
	switch m.activeTab {
	case tabRoutes:
		return m.doToggleRoute()
	case tabPeers:
		if !m.peerDetail {
			m.peerDetail = true
			m.peerSSHKey = ""
			m.peerSSHKeyErr = ""
			// Fetch SSH host key for this peer
			if peer := m.selectedPeer(); peer != nil {
				return m.doGetSSHHostKey(peer.IP)
			}
		}
	case tabEvents:
		if !m.eventsDetail {
			m.eventsDetail = true
			m.eventsScroll = 0
		}
	case tabNetMap:
		m.setActiveTab(tabPeers)
	}
	return nil
}

func (m *Model) setActiveTab(t tab) {
	m.activeTab = t
	m.quickSwitch = false
	m.helpOverlay = false
	m.clearTabInputFocus()
	m.settingsEditing = false
}

func (m *Model) clearTabInputFocus() {
	m.peersSearch.Blur()
	m.peersSearching = false
	m.eventsSearch.Blur()
	m.eventsSearching = false
	m.setupKeyInput.Blur()
	m.mgmtURLInput.Blur()
	m.profileNameInput.Blur()
	m.profileMgmtInput.Blur()
	m.traceSrcIP.Blur()
	m.traceDstIP.Blur()
	m.traceProto.Blur()
	m.traceSrcPort.Blur()
	m.traceDstPort.Blur()
	m.traceDir.Blur()
	m.settingsEditing = false
	m.profilesEditing = false
	m.traceEditing = false
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := fitContentWidth(m.renderHeader(), m.width)
	tabBar := fitContentWidth(m.renderTabBar(), m.width)
	footer := fitContentWidth(m.renderFooter(), m.width)
	if m.confirm != "" {
		footer = fitContentWidth(m.renderConfirm(), m.width)
	} else if m.quickSwitch {
		footer = fitContentWidth(m.renderQuickSwitch(), m.width)
	} else if m.helpOverlay {
		footer = fitContentWidth(m.renderHelpOverlay(), m.width)
	}

	contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(tabBar) - lipgloss.Height(footer)
	if contentHeight < 5 {
		contentHeight = 5
	}

	content := fitContentBox(m.renderContent(), m.width, contentHeight)
	contentStyle := lipgloss.NewStyle().
		Height(contentHeight).
		Width(m.width - 2)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabBar,
		contentStyle.Render(content),
		footer,
	)
}

func fitContentBox(content string, width, height int) string {
	if height <= 0 {
		return ""
	}
	lines := strings.Split(fitContentWidth(content, width), "\n")
	if len(lines) > height {
		lines = lines[:height]
		lines[height-1] = ansi.Truncate(styleNeutral.Render("..."), width, "")
		return strings.Join(lines, "\n")
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func fitContentWidth(content string, width int) string {
	if width <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "")
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderHeader() string {
	title := styleTitle.Render("NetBird TUI")

	status := ""
	if m.loading {
		status = m.spinner.View() + " " + m.loadingWhat + "..."
	} else if m.err != nil {
		errMsg := m.err.Error()
		if len(errMsg) > 60 {
			errMsg = errMsg[:60] + "…"
		}
		status = styleOffline.Render("● Error: " + errMsg)
	} else if m.status != nil && m.status.FullStatus != nil {
		fs := m.status.FullStatus
		if fs.LocalPeerState != nil {
			ip := fs.LocalPeerState.IP
			fqdn := fs.LocalPeerState.Fqdn
			mgmt := fs.ManagementState
			if mgmt != nil && mgmt.Connected {
				statusText := "  " + ip
				if m.width >= 96 {
					statusText += "  " + fqdn
				}
				status = styleOnline.Render("● Connected") + styleNeutral.Render(statusText)
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
	activeGroup := m.activeGroup()
	var groupParts []string
	for _, g := range tabGroups {
		lbl := fmt.Sprintf("[%s] %s", g.key, g.name)
		if g.id == activeGroup {
			groupParts = append(groupParts, styleActiveTab.Render(lbl))
		} else {
			groupParts = append(groupParts, styleInactiveTab.Render(lbl))
		}
	}

	var screenParts []string
	for _, g := range tabGroups {
		if g.id != activeGroup {
			continue
		}
		for _, t := range g.tabs {
			lbl := tabName(t)
			if m.activeTab == t {
				screenParts = append(screenParts, styleActiveTab.Render("● "+lbl))
			} else {
				screenParts = append(screenParts, styleInactiveTab.Render("○ "+lbl))
			}
		}
	}

	return styleNeutral.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		BorderBottom(true).
		Width(m.width - 2).
		Render(strings.Join(groupParts, "  ") + "\n" + strings.Join(screenParts, "  "))
}

func (m *Model) renderContent() string {
	switch m.activeTab {
	case tabStatus:
		return renderStatus(m)
	case tabPeers:
		return m.renderPeers()
	case tabRoutes:
		return m.renderRoutes()
	case tabDNS:
		return renderDNS(m)
	case tabEvents:
		return renderEvents(m)
	case tabNetMap:
		return renderNetworkMap(m, m.width-4)
	case tabProfiles:
		return renderProfiles(m)
	case tabDiagnostics:
		return renderDiagnostics(m)
	case tabServices:
		return renderServices(m)
	case tabFwdRules:
		return m.renderFwdRules()
	case tabSettings:
		return renderSettings(m)
	}
	return ""
}

func (m *Model) renderPeers() string {
	if m.status == nil || m.status.FullStatus == nil {
		return styleNeutral.Padding(1, 2).Render("No peer data available")
	}
	if m.peerDetail {
		if peer := m.selectedPeer(); peer != nil {
			return renderPeerDetail(peer, m.width, m.peerSSHKey, m.peerSSHKeyErr)
		}
		m.peerDetail = false
	}

	peers := m.status.FullStatus.Peers
	searchQuery := m.peersSearch.Value()
	filtered := m.filteredPeers()

	header := peersHeader(peers)
	searchBar := ""
	if m.peersSearching {
		searchBar = "  " + styleNeutral.Render("Search: ") + m.peersSearch.View() + "  " + styleNeutral.Render("Enter:apply  Esc:clear")
	} else {
		searchBar = styleNeutral.Render(fmt.Sprintf("  Filter: %s  •  / search", peerFilterLabel(m.peersFilter)))
		if searchQuery != "" || m.peersFilter != peerFilterAll {
			searchBar = styleNeutral.Render(fmt.Sprintf("  Showing %d/%d  •  filter: %s", len(filtered), len(peers), peerFilterLabel(m.peersFilter)))
			if searchQuery != "" {
				searchBar += styleNeutral.Render(fmt.Sprintf("  •  search: %q", searchQuery))
			}
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(0, 2).Render(header),
		lipgloss.NewStyle().Padding(0, 2).Render(searchBar),
		lipgloss.NewStyle().Padding(0, 2).Render(m.peersTable.View()),
	)
}

func (m *Model) renderRoutes() string {
	if len(m.networks) == 0 {
		return styleNeutral.Padding(1, 2).Render("No routes available")
	}
	header := routesHeader(m.networks)
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(0, 2).Render(header),
		lipgloss.NewStyle().Padding(0, 2).Render(m.routesTable.View()),
	)
}

func (m *Model) renderFwdRules() string {
	if len(m.fwdRules) == 0 {
		return styleNeutral.Padding(1, 2).Render("No forwarding rules")
	}
	return lipgloss.NewStyle().Padding(0, 2).Render(m.fwdTable.View())
}

func (m *Model) renderFooter() string {
	var help string
	nav := "Tab/S-Tab:Groups  ←→:Screens  g:Switch  ?:Help  q:Quit"
	switch m.activeTab {
	case tabRoutes:
		help = "Enter:Toggle  ↑↓:Nav  r:Refresh  " + nav
	case tabSettings:
		if m.settingsEditing {
			help = "Esc:Cancel  Tab:Next field  ctrl+s:Login"
		} else {
			help = "↑↓:Select  Enter:Edit  ctrl+s:Login  PgUp/PgDn:Page  " + nav
		}
	case tabPeers:
		if m.peerDetail {
			help = "Esc/Enter:Back  " + nav
		} else {
			help = "Enter:Detail  ↑↓:Nav  /:Search  f:Filter  x:Clear search  r:Refresh  " + nav
		}
	case tabEvents:
		if m.eventsDetail {
			help = "Esc:Back  ↑↓:Scroll  " + nav
		} else {
			help = "Enter:Detail  ↑↓:Nav  /:Search  f:Severity  x:Clear search  r:Refresh  " + nav
		}
	case tabNetMap:
		help = "r:Refresh  Enter:Open Peers  " + nav
	case tabProfiles:
		if m.profilesEditing {
			help = "Esc:Cancel  Tab:Next  ctrl+s:Add Profile"
		} else {
			help = "Enter:Switch  n:New  x:Remove  ↑↓:Nav  " + nav
		}
	case tabDiagnostics:
		switch m.diagMode {
		case diagModeTrace:
			if m.traceEditing {
				help = "Esc:Cancel  Tab:Next  ctrl+s:Trace"
			} else {
				help = "Enter/e:Edit  Esc:Back  " + nav
			}
		case diagModeStates:
			help = "c:Clean  x:Delete  ↑↓:Nav  Esc:Back  " + nav
		default:
			help = "t:Trace  s:States  l/L:Log level  b/B:Debug bundle  " + nav
		}
	case tabServices:
		help = "r:Refresh forwarding rules  " + nav
	default:
		help = "c:Connect  u:Up  d:Down  L:Logout  r:Refresh  " + nav
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
	case "debug-anon":
		actionStr = "create anonymized debug bundle"
	case "clean-state":
		if m.statesFocused < len(m.states) {
			actionStr = "clean state: " + m.states[m.statesFocused].Name
		}
	case "delete-state":
		if m.statesFocused < len(m.states) {
			actionStr = "delete state: " + m.states[m.statesFocused].Name
		}
	case "remove-profile":
		row := m.profilesTable.SelectedRow()
		if row != nil {
			actionStr = "remove profile: " + row[0]
		}
	case "switch-profile":
		row := m.profilesTable.SelectedRow()
		if row != nil {
			actionStr = "switch to profile: " + row[0]
		}
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

func (m *Model) renderQuickSwitch() string {
	var sb strings.Builder
	sb.WriteString(styleWarning.Render("Quick switch") + styleNeutral.Render("  1-4 groups, 5-0 screens, Esc close") + "\n")
	for _, g := range tabGroups {
		label := fmt.Sprintf("[%s] %s", g.key, g.name)
		if g.id == m.activeGroup() {
			sb.WriteString(styleActiveTab.Render(label))
		} else {
			sb.WriteString(styleInactiveTab.Render(label))
		}
		sb.WriteString("  ")
		for _, t := range g.tabs {
			key := ""
			for _, td := range tabs {
				if td.id == t {
					key = td.key
					break
				}
			}
			item := fmt.Sprintf("[%s] %s", key, tabName(t))
			if t == m.activeTab {
				sb.WriteString(styleActiveTab.Render(item))
			} else {
				sb.WriteString(styleNeutral.Render(item))
			}
			sb.WriteString("  ")
		}
		sb.WriteString("\n")
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorBlue).
		Padding(0, 2).
		Width(m.width - 2).
		Render(sb.String())
}

func (m *Model) renderHelpOverlay() string {
	lines := []string{
		styleWarning.Render("Help: ") + styleValue.Render(tabName(m.activeTab)),
		"Tab / Shift+Tab: switch Monitor, Network, Manage, Tools",
		"Left / Right: switch screens inside the active group",
		"1-4: jump to a group  •  g: quick switch  •  /: screen search",
	}
	if screen := m.screenHelp(); screen != "" {
		lines = append(lines, screen)
	}
	lines = append(lines, "Press any key to close")
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorBlue).
		Padding(0, 2).
		Width(m.width - 2).
		Render(strings.Join(lines, "\n"))
}

func (m *Model) screenHelp() string {
	switch m.activeTab {
	case tabPeers:
		return "Peers: / search, f cycle all/online/offline/relayed, Enter detail, x clear search"
	case tabEvents:
		return "Events: / search, f severity filter, Enter detail, x clear search"
	case tabRoutes:
		return "Routes: Enter toggles selected route after confirmation from NetBird daemon"
	case tabDiagnostics:
		return "Diagnostics: t packet trace, s daemon states, b/B debug bundle, destructive state actions confirm"
	case tabSettings:
		return "Settings: setup-key login is editable; config summary is read-only"
	case tabServices:
		return "Services: forwarding rules are read-only in this release"
	}
	return ""
}

// ─── Peers key handler ───────────────────────────────────────────────────────

func (m *Model) handlePeersKey(msg tea.KeyMsg) tea.Cmd {
	if m.peerDetail {
		switch msg.String() {
		case "esc", "enter", "q":
			if msg.String() == "q" {
				return tea.Quit
			}
			m.peerDetail = false
			m.peerSSHKey = ""
			return nil
		}
		return nil
	}

	// Search mode
	if m.peersSearching {
		switch msg.String() {
		case "esc":
			m.peersSearching = false
			m.peersSearch.SetValue("")
			m.peersSearch.Blur()
			m.rebuildPeersTableKeepingSelection()
			return nil
		case "enter":
			m.peersSearching = false
			m.peersSearch.Blur()
			m.rebuildPeersTableKeepingSelection()
			return nil
		default:
			var c tea.Cmd
			m.peersSearch, c = m.peersSearch.Update(msg)
			m.rebuildPeersTableKeepingSelection()
			return c
		}
	}

	switch msg.String() {
	case "/":
		m.peersSearching = true
		m.peersSearch.Focus()
		return nil
	case "f":
		m.peersFilter = nextPeerFilter(m.peersFilter)
		m.rebuildPeersTableKeepingSelection()
		return nil
	case "x":
		m.peersSearch.SetValue("")
		m.peersSearching = false
		m.peersSearch.Blur()
		m.rebuildPeersTableKeepingSelection()
		return nil
	}

	return m.handleGlobalKey(msg)
}
