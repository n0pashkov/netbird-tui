package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/netbirdio/netbird/client/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestPeerFilteringSearchAndModes(t *testing.T) {
	peers := []*proto.PeerState{
		{Fqdn: "alpha.netbird.test", IP: "100.64.0.1", ConnStatus: "Connected"},
		{Fqdn: "beta.netbird.test", IP: "100.64.0.2", ConnStatus: "Disconnected"},
		{Fqdn: "relay.netbird.test", IP: "100.64.0.3", ConnStatus: "Connected", Relayed: true},
	}

	if got := filterPeers(peers, "alpha", peerFilterAll); len(got) != 1 || got[0].Fqdn != "alpha.netbird.test" {
		t.Fatalf("search filter mismatch: %#v", got)
	}
	if got := filterPeers(peers, "", peerFilterOnline); len(got) != 2 {
		t.Fatalf("online filter len = %d", len(got))
	}
	if got := filterPeers(peers, "", peerFilterOffline); len(got) != 1 || got[0].Fqdn != "beta.netbird.test" {
		t.Fatalf("offline filter mismatch: %#v", got)
	}
	if got := filterPeers(peers, "", peerFilterRelayed); len(got) != 1 || !got[0].Relayed {
		t.Fatalf("relayed filter mismatch: %#v", got)
	}
}

func TestNavigationGroupsAndScreens(t *testing.T) {
	m := New(nil)
	m.width = 100
	m.height = 30

	if cmd := m.handleGlobalKey(tea.KeyMsg{Type: tea.KeyTab}); cmd != nil {
		t.Fatalf("unexpected command")
	}
	if m.activeTab != tabRoutes {
		t.Fatalf("Tab should move to Network/Routes, got %v", m.activeTab)
	}
	m.handleGlobalKey(tea.KeyMsg{Type: tea.KeyRight})
	if m.activeTab != tabDNS {
		t.Fatalf("right should move within group to DNS, got %v", m.activeTab)
	}
	m.handleGlobalKey(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.activeTab != tabStatus {
		t.Fatalf("shift+tab should move to Monitor/Status, got %v", m.activeTab)
	}
	m.handleGlobalKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	if m.activeTab != tabDiagnostics {
		t.Fatalf("4 should move to Tools/Diagnostics, got %v", m.activeTab)
	}
	m.setActiveGroup(groupMonitor)
	m.handleGlobalKey(tea.KeyMsg{Type: tea.KeyRight})
	m.handleGlobalKey(tea.KeyMsg{Type: tea.KeyRight})
	m.handleGlobalKey(tea.KeyMsg{Type: tea.KeyRight})
	if m.activeTab != tabNetMap {
		t.Fatalf("right should include Monitor/Map, got %v", m.activeTab)
	}
}

func TestQuickSwitchAndHelpKeys(t *testing.T) {
	m := New(nil)
	m.handleGlobalKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if !m.quickSwitch {
		t.Fatalf("g should open quick switch")
	}
	m.handleQuickSwitch(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}})
	if m.quickSwitch || m.activeTab != tabFwdRules {
		t.Fatalf("quick switch 6 should select forwarding, tab=%v open=%v", m.activeTab, m.quickSwitch)
	}
	m.quickSwitch = true
	m.handleQuickSwitch(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	if m.quickSwitch || m.activeTab != tabNetMap {
		t.Fatalf("quick switch m should select map, tab=%v open=%v", m.activeTab, m.quickSwitch)
	}
	m.handleGlobalKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !m.helpOverlay {
		t.Fatalf("? should open help overlay")
	}
	m.handleHelpOverlay(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if m.helpOverlay {
		t.Fatalf("any key should close help overlay")
	}
}

func TestNetworkMapRendersDedicatedScreen(t *testing.T) {
	m := New(nil)
	m.width = 120
	m.height = 32
	m.status = &proto.StatusResponse{FullStatus: &proto.FullStatus{
		LocalPeerState: &proto.LocalPeerState{IP: "100.64.0.1", Fqdn: "local.example"},
		Peers: []*proto.PeerState{
			{Fqdn: "direct.example", IP: "100.64.0.2", ConnStatus: "Connected"},
			{Fqdn: "relay.example", IP: "100.64.0.3", ConnStatus: "Connected", Relayed: true},
		},
	}}
	m.networks = []*proto.Network{{ID: "home-network", Range: "10.0.0.0/24", Selected: true}}
	m.setActiveTab(tabNetMap)

	view := m.renderContent()
	for _, want := range []string{"Network Map", "local.example", "direct.example", "home-network"} {
		if !strings.Contains(view, want) {
			t.Fatalf("map missing %q in %q", want, view)
		}
	}
}

func TestNetworkMapDoesNotOverflowWidth(t *testing.T) {
	m := New(nil)
	m.width = 80
	m.height = 24
	m.status = &proto.StatusResponse{FullStatus: &proto.FullStatus{
		LocalPeerState: &proto.LocalPeerState{IP: "100.64.0.1", Fqdn: "very-long-local-hostname.example"},
		Peers: []*proto.PeerState{
			{Fqdn: "very-long-peer-hostname-one.example", IP: "100.64.0.2", ConnStatus: "Connected"},
			{Fqdn: "very-long-peer-hostname-two.example", IP: "100.64.0.3", ConnStatus: "Connected", Relayed: true},
		},
	}}
	m.networks = []*proto.Network{{ID: "very-long-home-network-route-name", Range: "10.0.0.0/24", Selected: true}}

	view := renderNetworkMap(m, m.width-4)
	for i, line := range strings.Split(view, "\n") {
		if width := lipgloss.Width(line); width > m.width {
			t.Fatalf("map line %d width = %d, want <= %d: %q", i, width, m.width, line)
		}
	}
}

func TestPeersSearchSelectionAndDetailUseFilteredList(t *testing.T) {
	m := New(nil)
	m.width = 100
	m.height = 30
	m.status = &proto.StatusResponse{FullStatus: &proto.FullStatus{Peers: []*proto.PeerState{
		{Fqdn: "alpha", IP: "100.64.0.1", ConnStatus: "Connected"},
		{Fqdn: "beta", IP: "100.64.0.2", ConnStatus: "Disconnected", Latency: durationpb.New(0)},
	}}}
	m.peersSearch.SetValue("beta")
	m.rebuildPeersTableKeepingSelection()

	peer := m.selectedPeer()
	if peer == nil || peer.Fqdn != "beta" {
		t.Fatalf("selected filtered peer = %#v", peer)
	}
	m.peerDetail = true
	view := m.renderPeers()
	if !strings.Contains(view, "beta") || strings.Contains(view, "alpha") {
		t.Fatalf("detail should render filtered selection, got %q", view)
	}
}

func TestHelpers(t *testing.T) {
	if got := formatBytes(1536); got != "1.5KB" {
		t.Fatalf("formatBytes = %q", got)
	}
	if got := wordWrap("alpha beta gamma", 8); got != "alpha\nbeta\ngamma" {
		t.Fatalf("wordWrap = %q", got)
	}
	if got := nextLogLevel(proto.LogLevel_INFO, true); got != proto.LogLevel_DEBUG {
		t.Fatalf("nextLogLevel up = %v", got)
	}
	if got := nextLogLevel(proto.LogLevel_INFO, false); got != proto.LogLevel_WARN {
		t.Fatalf("nextLogLevel down = %v", got)
	}
}

func TestTraceInputValidation(t *testing.T) {
	if err := validateTraceInput("100.64.0.1", "100.64.0.2", "tcp", "0", "443", "in"); err != nil {
		t.Fatalf("valid trace input rejected: %v", err)
	}
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"bad source", validateTraceInput("bad", "100.64.0.2", "tcp", "0", "443", "in")},
		{"bad proto", validateTraceInput("100.64.0.1", "100.64.0.2", "sctp", "0", "443", "in")},
		{"bad port", validateTraceInput("100.64.0.1", "100.64.0.2", "tcp", "70000", "443", "in")},
		{"bad dir", validateTraceInput("100.64.0.1", "100.64.0.2", "tcp", "0", "443", "sideways")},
	} {
		if tc.err == nil {
			t.Fatalf("%s should fail", tc.name)
		}
	}
}

func TestDiagnosticsOverviewShowsDebugCommands(t *testing.T) {
	m := New(nil)
	m.width = 120
	m.height = 30

	view := renderDiagnosticsOverview(m)
	for _, want := range []string{"Debug Commands", "c:dump config", "a:capture packets for 10s", "p:persistence on", "f:debug for 1m"} {
		if !strings.Contains(view, want) {
			t.Fatalf("diagnostics overview missing %q in %q", want, view)
		}
	}
}

func TestEventsScreenFallsBackToStatusEvents(t *testing.T) {
	m := New(nil)
	m.width = 120
	m.height = 30
	m.status = &proto.StatusResponse{FullStatus: &proto.FullStatus{
		Events: []*proto.SystemEvent{{UserMessage: "status event"}},
	}}
	m.events = nil
	m.eventsTable = buildEventsTable(m.eventsForDisplay(), m.eventsFilter, m.eventsSearch.Value(), m.width, m.height)

	view := renderEvents(m)
	if !strings.Contains(view, "1 total") || !strings.Contains(view, "status event") {
		t.Fatalf("events screen should render status events fallback, got %q", view)
	}
}

func TestStatusMsgRebuildsEventsFallbackTable(t *testing.T) {
	m := New(nil)
	m.width = 120
	m.height = 30

	_, _ = m.Update(statusMsg{status: &proto.StatusResponse{FullStatus: &proto.FullStatus{
		Events: []*proto.SystemEvent{{UserMessage: "status event"}},
	}}})

	view := renderEvents(m)
	if !strings.Contains(view, "status event") {
		t.Fatalf("status update should rebuild fallback events table, got %q", view)
	}
}

func TestFooterShowsExposeActionsAndConfirmIntercepts(t *testing.T) {
	m := New(nil)
	m.width = 120
	m.height = 30
	m.setActiveTab(tabServices)
	if footer := m.renderFooter(); !strings.Contains(footer, "New expose") || !strings.Contains(footer, "Stop expose") {
		t.Fatalf("services footer missing expose actions: %q", footer)
	}
	m.setActiveTab(tabSettings)
	m.settingsEditing = true
	if footer := m.renderFooter(); strings.Contains(footer, "Save Config") || strings.Contains(footer, "ctrl+a") {
		t.Fatalf("settings footer advertises config save: %q", footer)
	}
	m.confirm = "down"
	m.handleTabKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if m.quickSwitch {
		t.Fatalf("confirmation should intercept global keys")
	}
}

func TestBuildExposeRequestValidation(t *testing.T) {
	m := New(nil)
	m.exposePortInput.SetValue("8080")
	m.exposeProtocolInput.SetValue("tcp")
	m.exposeExternalInput.SetValue("4433")
	req, protocol, err := m.buildExposeRequest()
	if err != nil {
		t.Fatalf("valid expose request rejected: %v", err)
	}
	if protocol != "tcp" || req.Port != 8080 || req.ListenPort != 4433 || req.Protocol != proto.ExposeProtocol_EXPOSE_TCP {
		t.Fatalf("request mismatch: protocol=%s req=%#v", protocol, req)
	}

	m.exposeProtocolInput.SetValue("http")
	if _, _, err := m.buildExposeRequest(); err == nil {
		t.Fatalf("external port should be rejected for http")
	}

	m.exposeExternalInput.SetValue("")
	m.exposePinInput.SetValue("12345x")
	if _, _, err := m.buildExposeRequest(); err == nil {
		t.Fatalf("non-numeric pin should be rejected")
	}
}

func TestViewDoesNotOverflowTerminalHeight(t *testing.T) {
	m := New(nil)
	m.width = 100
	m.height = 20
	m.status = &proto.StatusResponse{FullStatus: &proto.FullStatus{
		LocalPeerState: &proto.LocalPeerState{IP: "100.64.0.1", Fqdn: "host.example", KernelInterface: true},
		Peers: []*proto.PeerState{
			{Fqdn: "p1", IP: "100.64.0.2", ConnStatus: "Connected"},
			{Fqdn: "p2", IP: "100.64.0.3", ConnStatus: "Connected"},
			{Fqdn: "p3", IP: "100.64.0.4", ConnStatus: "Disconnected"},
		},
	}}
	m.networks = []*proto.Network{{ID: "route-1", Range: "10.0.0.0/24", Selected: true}}
	m.events = []*proto.SystemEvent{{UserMessage: "event"}}

	view := m.View()
	if got := lipgloss.Height(view); got != m.height {
		t.Fatalf("view height = %d, want %d", got, m.height)
	}
	if !strings.Contains(view, "[1] Monitor") || !strings.Contains(view, "Status") {
		t.Fatalf("view should keep header and tabs visible: %q", view)
	}
}

func TestViewDoesNotOverflowTerminalWidth(t *testing.T) {
	m := New(nil)
	m.width = 60
	m.height = 20
	m.status = &proto.StatusResponse{FullStatus: &proto.FullStatus{
		LocalPeerState: &proto.LocalPeerState{IP: "100.88.36.8/16", Fqdn: "very-long-hostname.nb.vm644479.eurodir.example", KernelInterface: true},
		Peers:          []*proto.PeerState{{Fqdn: "peer", IP: "100.64.0.2", ConnStatus: "Connected"}},
	}}

	for i, line := range strings.Split(m.View(), "\n") {
		if width := lipgloss.Width(line); width > m.width {
			t.Fatalf("line %d width = %d, want <= %d: %q", i, width, m.width, line)
		}
	}
}
