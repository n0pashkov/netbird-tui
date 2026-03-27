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
)

const tabCount = 10

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
type setConfigMsg struct{ err error }
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
	confirm string // pending confirmation action

	// ── Status tab ──
	status   *proto.StatusResponse
	features *proto.GetFeaturesResponse

	// ── Peers tab ──
	peersTable    table.Model
	peerDetail    bool
	peerSSHKey    string
	peerSSHKeyErr string

	// ── Routes tab ──
	networks    []*proto.Network
	routesTable table.Model

	// ── DNS tab ──
	// Data comes from status.FullStatus.DnsServers
	dnsSelected int

	// ── Events tab ──
	events       []*proto.SystemEvent
	eventsTable  table.Model
	eventsFilter proto.SystemEvent_Severity // -1 = all (since 0 == INFO)
	eventsDetail bool
	eventsScroll int

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
	fwdRules    []*proto.ForwardingRule
	fwdTable    table.Model
	// Service exposure sub-form is in services.go
	svcPortInput   textinput.Model
	svcProtoInput  textinput.Model
	svcGroupInput  textinput.Model
	svcDomainInput textinput.Model
	svcFocused     int
	svcEditing     bool
	svcMsg         string

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

	// Service expose inputs
	sPort := textinput.New()
	sPort.Placeholder = "8080"
	sPort.CharLimit = 5

	sProto := textinput.New()
	sProto.Placeholder = "tcp"
	sProto.CharLimit = 10

	sGroup := textinput.New()
	sGroup.Placeholder = "All (comma-separated)"
	sGroup.CharLimit = 256

	sDomain := textinput.New()
	sDomain.Placeholder = "optional.domain.example"
	sDomain.CharLimit = 256

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
		svcPortInput:     sPort,
		svcProtoInput:    sProto,
		svcGroupInput:    sGroup,
		svcDomainInput:   sDomain,
		eventsFilter:     -1,
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

func (m *Model) doSetConfig(req *proto.SetConfigRequest) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return setConfigMsg{err: m.client.SetConfig(ctx, req)}
	}
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
		m.peersTable = buildPeersTable(m.status.FullStatus.Peers, m.width, m.height)
	}
	if m.networks != nil {
		m.routesTable = buildRoutesTable(m.networks, m.width, m.height)
	}
	if m.fwdRules != nil {
		m.fwdTable = buildFwdTable(m.fwdRules, m.width, m.height)
	}
	if m.events != nil {
		m.eventsTable = buildEventsTable(m.events, m.eventsFilter, m.width, m.height)
	}
	if m.profiles != nil {
		m.profilesTable = buildProfilesTable(m.profiles, m.activeProfile, m.width, m.height)
	}
	if m.states != nil {
		m.statesTable = buildStatesTable(m.states, m.width, m.height)
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

	case eventsMsg:
		if msg.err == nil {
			m.events = msg.events
			m.eventsTable = buildEventsTable(m.events, m.eventsFilter, m.width, m.height)
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

	case setConfigMsg:
		m.loading = false
		if msg.err != nil {
			m.settingsMsg = "Error: " + msg.err.Error()
		} else {
			m.settingsMsg = "Config saved"
			cmds = append(cmds, m.fetchConfig())
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

func (m *Model) handleGlobalKey(msg tea.KeyMsg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg.String() {
	case "q":
		return tea.Quit
	case "left":
		if m.activeTab > tabStatus {
			m.activeTab--
			m.clearTabInputFocus()
		}
	case "right":
		if m.activeTab < tab(tabCount-1) {
			m.activeTab++
			m.clearTabInputFocus()
		}
	case "1":
		m.setActiveTab(tabStatus)
	case "2":
		m.setActiveTab(tabPeers)
	case "3":
		m.setActiveTab(tabRoutes)
	case "4":
		m.setActiveTab(tabDNS)
	case "5":
		m.setActiveTab(tabEvents)
	case "6":
		m.setActiveTab(tabProfiles)
	case "7":
		m.setActiveTab(tabDiagnostics)
	case "8":
		m.setActiveTab(tabServices)
	case "9":
		m.setActiveTab(tabFwdRules)
	case "0":
		m.setActiveTab(tabSettings)
	case "r":
		m.loading = true
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
			row := m.peersTable.SelectedRow()
			if row != nil && len(row) > 1 {
				return m.doGetSSHHostKey(row[1]) // IP in col 1
			}
		}
	case tabEvents:
		if !m.eventsDetail {
			m.eventsDetail = true
			m.eventsScroll = 0
		}
	}
	return nil
}

func (m *Model) setActiveTab(t tab) {
	m.activeTab = t
	m.clearTabInputFocus()
	m.settingsEditing = false
}

func (m *Model) clearTabInputFocus() {
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
	m.svcPortInput.Blur()
	m.svcProtoInput.Blur()
	m.svcGroupInput.Blur()
	m.svcDomainInput.Blur()
	m.settingsEditing = false
	m.profilesEditing = false
	m.traceEditing = false
	m.svcEditing = false
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var sections []string
	sections = append(sections, m.renderHeader())
	sections = append(sections, m.renderTabBar())

	contentHeight := m.height - 10
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
	// Two rows of tabs: 1-5 and 6-0
	row1 := []struct {
		id    tab
		label string
		key   string
	}{
		{tabStatus, "Status", "1"},
		{tabPeers, "Peers", "2"},
		{tabRoutes, "Routes", "3"},
		{tabDNS, "DNS", "4"},
		{tabEvents, "Events", "5"},
	}
	row2 := []struct {
		id    tab
		label string
		key   string
	}{
		{tabProfiles, "Profiles", "6"},
		{tabDiagnostics, "Diagnostics", "7"},
		{tabServices, "Services", "8"},
		{tabFwdRules, "Forwarding", "9"},
		{tabSettings, "Settings", "0"},
	}

	renderRow := func(items []struct {
		id    tab
		label string
		key   string
	}) string {
		var parts []string
		for _, t := range items {
			lbl := fmt.Sprintf("[%s] %s", t.key, t.label)
			if m.activeTab == t.id {
				parts = append(parts, styleActiveTab.Render(lbl))
			} else {
				parts = append(parts, styleInactiveTab.Render(lbl))
			}
		}
		return strings.Join(parts, "  ")
	}

	r1 := renderRow(row1)
	r2 := renderRow(row2)

	return styleNeutral.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		BorderBottom(true).
		Width(m.width - 2).
		Render(r1 + "\n" + r2)
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
		row := m.peersTable.SelectedRow()
		if row != nil {
			for _, p := range m.status.FullStatus.Peers {
				if p.Fqdn == row[0] || p.IP == row[1] {
					return renderPeerDetail(p, m.width, m.peerSSHKey, m.peerSSHKeyErr)
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
	switch m.activeTab {
	case tabRoutes:
		help = "Enter:Toggle  ↑↓:Nav  c:Connect  u:Up  d:Down  r:Refresh  ←→/1-0:Tabs  q:Quit"
	case tabSettings:
		if m.settingsEditing {
			help = "Esc:Cancel  Tab:NextField  ctrl+s:Save Login  ctrl+a:Save Config"
		} else {
			help = "↑↓/Tab:Select  Enter:Edit  ctrl+s:Login  ←→/1-0:Tabs  q:Quit"
		}
	case tabPeers:
		if m.peerDetail {
			help = "Esc/Enter:Back"
		} else {
			help = "Enter:Detail  ↑↓:Nav  c:Connect  r:Refresh  ←→/1-0:Tabs  q:Quit"
		}
	case tabEvents:
		if m.eventsDetail {
			help = "Esc:Back  ↑↓:Scroll"
		} else {
			help = "Enter:Detail  ↑↓:Nav  f:Filter  r:Refresh  ←→/1-0:Tabs  q:Quit"
		}
	case tabProfiles:
		if m.profilesEditing {
			help = "Esc:Cancel  Tab:Next  ctrl+s:Add Profile"
		} else {
			help = "Enter:Switch  n:New  x:Remove  ↑↓:Nav  ←→/1-0:Tabs  q:Quit"
		}
	case tabDiagnostics:
		switch m.diagMode {
		case diagModeTrace:
			if m.traceEditing {
				help = "Esc:Cancel  Tab:Next  ctrl+s:Trace"
			} else {
				help = "t:Trace  s:States  ←/Esc:Back  ←→/1-0:Tabs  q:Quit"
			}
		case diagModeStates:
			help = "c:Clean  x:Delete  ↑↓:Nav  t:Trace  ←/Esc:Back  ←→/1-0:Tabs  q:Quit"
		default:
			help = "t:Trace Packet  s:States  l/L:LogLevel±  b:Debug  B:Debug(anon)  ←→/1-0:Tabs  q:Quit"
		}
	case tabServices:
		if m.svcEditing {
			help = "Esc:Cancel  Tab:Next  ctrl+s:Save"
		} else {
			help = "n:New Expose  ←→/1-0:Tabs  q:Quit"
		}
	default:
		help = "c:Connect  u:Up  d:Down  L:Logout  b:Debug  r:Refresh  ←→/1-0:Tabs  ↑↓:Nav  q:Quit"
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
	return m.handleGlobalKey(msg)
}
