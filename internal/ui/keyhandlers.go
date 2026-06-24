package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/netbirdio/netbird/client/proto"
)

// handleSettingsKey handles keyboard input for the Settings tab.
func (m *Model) handleSettingsKey(msg tea.KeyMsg) tea.Cmd {
	if m.settingsEditing {
		switch msg.String() {
		case "esc":
			m.settingsEditing = false
			m.clearTabInputFocus()
			return nil
		case "ctrl+s":
			m.settingsMsg = ""
			m.loading = true
			return m.doLogin()
		case "tab":
			maxFields := 2
			m.settingsFocused = (m.settingsFocused + 1) % maxFields
			m.focusSettingsField()
			return nil
		default:
			return m.updateSettingsInput(msg)
		}
	}

	// Browse mode
	switch msg.String() {
	case "up", "k":
		if m.settingsFocused > 0 {
			m.settingsFocused--
		}
	case "down", "j", "tab":
		m.settingsFocused++
	case "enter":
		m.settingsEditing = true
		m.focusSettingsField()
	case "ctrl+s":
		m.settingsMsg = ""
		m.loading = true
		return m.doLogin()
	case "pgup":
		if m.settingsPage > 0 {
			m.settingsPage--
		}
	case "pgdn":
		m.settingsPage++
		if m.settingsPage > 1 {
			m.settingsPage = 1
		}
	default:
		return m.handleGlobalKey(msg)
	}
	return nil
}

func (m *Model) focusSettingsField() {
	m.setupKeyInput.Blur()
	m.mgmtURLInput.Blur()
	switch m.settingsFocused {
	case 0:
		m.setupKeyInput.Focus()
	case 1:
		m.mgmtURLInput.Focus()
	}
}

func (m *Model) updateSettingsInput(msg tea.KeyMsg) tea.Cmd {
	var cmds []tea.Cmd
	switch m.settingsFocused {
	case 0:
		var c tea.Cmd
		m.setupKeyInput, c = m.setupKeyInput.Update(msg)
		cmds = append(cmds, c)
	case 1:
		var c tea.Cmd
		m.mgmtURLInput, c = m.mgmtURLInput.Update(msg)
		cmds = append(cmds, c)
	}
	return tea.Batch(cmds...)
}

// handleProfilesKey handles keyboard input for the Profiles tab.
func (m *Model) handleProfilesKey(msg tea.KeyMsg) tea.Cmd {
	if m.profilesEditing {
		switch msg.String() {
		case "esc":
			m.profilesEditing = false
			m.clearTabInputFocus()
			m.profilesMsg = ""
			return nil
		case "ctrl+s":
			m.loading = true
			return m.doAddProfile()
		case "tab":
			m.profilesFocused = (m.profilesFocused + 1) % 2
			m.profileNameInput.Blur()
			m.profileMgmtInput.Blur()
			if m.profilesFocused == 0 {
				m.profileNameInput.Focus()
			} else {
				m.profileMgmtInput.Focus()
			}
			return nil
		default:
			var cmds []tea.Cmd
			var c tea.Cmd
			if m.profilesFocused == 0 {
				m.profileNameInput, c = m.profileNameInput.Update(msg)
			} else {
				m.profileMgmtInput, c = m.profileMgmtInput.Update(msg)
			}
			cmds = append(cmds, c)
			return tea.Batch(cmds...)
		}
	}

	// Browse mode
	switch msg.String() {
	case "up", "k":
		m.profilesTable, _ = m.profilesTable.Update(msg)
	case "down", "j":
		m.profilesTable, _ = m.profilesTable.Update(msg)
	case "enter":
		m.confirm = "switch-profile"
		return nil
	case "n":
		m.profilesEditing = true
		m.profilesFocused = 0
		m.profileNameInput.Focus()
		return nil
	case "x":
		m.confirm = "remove-profile"
		return nil
	case "r":
		return m.fetchProfiles()
	default:
		return m.handleGlobalKey(msg)
	}
	return nil
}

// handleDiagnosticsKey handles keyboard input for the Diagnostics tab.
func (m *Model) handleDiagnosticsKey(msg tea.KeyMsg) tea.Cmd {
	switch m.diagMode {
	case diagModeTrace:
		return m.handleTraceKey(msg)
	case diagModeStates:
		return m.handleStatesKey(msg)
	case diagModeOutput:
		return m.handleDebugOutputKey(msg)
	}

	// Overview mode
	switch msg.String() {
	case "t":
		m.diagMode = diagModeTrace
		return nil
	case "s":
		m.diagMode = diagModeStates
		return tea.Batch(m.fetchStates())
	case "c":
		m.loading = true
		m.loadingWhat = "Dumping config"
		return m.doDebugConfigDump()
	case "a":
		m.loading = true
		m.loadingWhat = "Capturing packets"
		return m.doDebugCommand("Packet Capture", 15*time.Second, "capture", "--duration", "10s")
	case "f":
		m.confirm = "debug-for"
		return nil
	case "p":
		m.loading = true
		m.loadingWhat = "Setting persistence"
		return m.doSetSyncPersistence(true)
	case "P":
		m.loading = true
		m.loadingWhat = "Setting persistence"
		return m.doSetSyncPersistence(false)
	case "l":
		// Increase log level
		if m.logLevelKnown {
			next := nextLogLevel(m.logLevel, true)
			m.loading = true
			return m.doSetLogLevel(next)
		}
	case "L":
		// Decrease log level
		if m.logLevelKnown {
			next := nextLogLevel(m.logLevel, false)
			m.loading = true
			return m.doSetLogLevel(next)
		}
	case "b":
		m.confirm = "debug"
		return nil
	case "B":
		m.confirm = "debug-anon"
		return nil
	case "r":
		return tea.Batch(m.fetchStates(), m.fetchLogLevel())
	default:
		return m.handleGlobalKey(msg)
	}
	return nil
}

func (m *Model) handleDebugOutputKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "enter":
		m.diagMode = diagModeOverview
	case "q":
		return tea.Quit
	default:
		return m.handleGlobalKey(msg)
	}
	return nil
}

func (m *Model) handleTraceKey(msg tea.KeyMsg) tea.Cmd {
	if m.traceEditing {
		switch msg.String() {
		case "esc":
			m.traceEditing = false
			m.clearTraceInputs()
			return nil
		case "ctrl+s":
			m.loading = true
			m.traceResult = nil
			m.traceErr = ""
			return m.doTracePacket()
		case "tab":
			m.traceFocused = (m.traceFocused + 1) % 6
			m.focusTraceField()
			return nil
		default:
			return m.updateTraceInput(msg)
		}
	}

	switch msg.String() {
	case "enter", "e":
		m.traceEditing = true
		m.traceFocused = 0
		m.focusTraceField()
	case "esc", "q":
		if msg.String() == "q" {
			return tea.Quit
		}
		m.diagMode = diagModeOverview
		m.traceResult = nil
	default:
		return m.handleGlobalKey(msg)
	}
	return nil
}

func (m *Model) clearTraceInputs() {
	m.traceSrcIP.Blur()
	m.traceDstIP.Blur()
	m.traceProto.Blur()
	m.traceSrcPort.Blur()
	m.traceDstPort.Blur()
	m.traceDir.Blur()
}

func (m *Model) focusTraceField() {
	m.clearTraceInputs()
	switch m.traceFocused {
	case 0:
		m.traceSrcIP.Focus()
	case 1:
		m.traceDstIP.Focus()
	case 2:
		m.traceProto.Focus()
	case 3:
		m.traceSrcPort.Focus()
	case 4:
		m.traceDstPort.Focus()
	case 5:
		m.traceDir.Focus()
	}
}

func (m *Model) updateTraceInput(msg tea.KeyMsg) tea.Cmd {
	var c tea.Cmd
	switch m.traceFocused {
	case 0:
		m.traceSrcIP, c = m.traceSrcIP.Update(msg)
	case 1:
		m.traceDstIP, c = m.traceDstIP.Update(msg)
	case 2:
		m.traceProto, c = m.traceProto.Update(msg)
	case 3:
		m.traceSrcPort, c = m.traceSrcPort.Update(msg)
	case 4:
		m.traceDstPort, c = m.traceDstPort.Update(msg)
	case 5:
		m.traceDir, c = m.traceDir.Update(msg)
	}
	return c
}

func (m *Model) handleStatesKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up", "k":
		m.statesTable, _ = m.statesTable.Update(msg)
	case "down", "j":
		m.statesTable, _ = m.statesTable.Update(msg)
	case "c":
		row := m.statesTable.SelectedRow()
		if row != nil {
			// Find index in states
			for i, s := range m.states {
				if s.Name == row[0] {
					m.statesFocused = i
					break
				}
			}
			m.confirm = "clean-state"
		}
	case "x":
		row := m.statesTable.SelectedRow()
		if row != nil {
			for i, s := range m.states {
				if s.Name == row[0] {
					m.statesFocused = i
					break
				}
			}
			m.confirm = "delete-state"
		}
	case "r":
		return m.fetchStates()
	case "esc":
		m.diagMode = diagModeOverview
	default:
		return m.handleGlobalKey(msg)
	}
	return nil
}

// handleEventsKey handles keyboard input for the Events tab.
func (m *Model) handleEventsKey(msg tea.KeyMsg) tea.Cmd {
	if m.eventsSearching {
		switch msg.String() {
		case "esc":
			m.eventsSearching = false
			m.eventsSearch.SetValue("")
			m.eventsSearch.Blur()
			m.eventsTable = buildEventsTable(m.eventsForDisplay(), m.eventsFilter, m.eventsSearch.Value(), m.width, m.height)
			return nil
		case "enter":
			m.eventsSearching = false
			m.eventsSearch.Blur()
			m.eventsTable = buildEventsTable(m.eventsForDisplay(), m.eventsFilter, m.eventsSearch.Value(), m.width, m.height)
			return nil
		default:
			var c tea.Cmd
			m.eventsSearch, c = m.eventsSearch.Update(msg)
			m.eventsTable = buildEventsTable(m.eventsForDisplay(), m.eventsFilter, m.eventsSearch.Value(), m.width, m.height)
			return c
		}
	}

	if m.eventsDetail {
		switch msg.String() {
		case "esc", "enter":
			m.eventsDetail = false
			m.eventsScroll = 0
			return nil
		case "up", "k":
			if m.eventsScroll > 0 {
				m.eventsScroll--
			}
		case "down", "j":
			m.eventsScroll++
		case "q":
			return tea.Quit
		}
		return nil
	}

	switch msg.String() {
	case "up", "k":
		m.eventsTable, _ = m.eventsTable.Update(msg)
	case "down", "j":
		m.eventsTable, _ = m.eventsTable.Update(msg)
	case "enter":
		m.eventsDetail = true
		m.eventsScroll = 0
		return nil
	case "f":
		// Cycle through severity filters: -1=all, WARNING, ERROR, CRITICAL
		switch m.eventsFilter {
		case -1:
			m.eventsFilter = proto.SystemEvent_WARNING
		case proto.SystemEvent_WARNING:
			m.eventsFilter = proto.SystemEvent_ERROR
		case proto.SystemEvent_ERROR:
			m.eventsFilter = proto.SystemEvent_CRITICAL
		default:
			m.eventsFilter = -1
		}
		m.eventsTable = buildEventsTable(m.eventsForDisplay(), m.eventsFilter, m.eventsSearch.Value(), m.width, m.height)
		return nil
	case "r":
		return m.fetchEvents()
	case "/":
		m.eventsSearching = true
		m.eventsSearch.Focus()
		return nil
	case "x":
		m.eventsSearch.SetValue("")
		m.eventsSearching = false
		m.eventsSearch.Blur()
		m.eventsTable = buildEventsTable(m.eventsForDisplay(), m.eventsFilter, m.eventsSearch.Value(), m.width, m.height)
		return nil
	default:
		return m.handleGlobalKey(msg)
	}
	return nil
}

// handleServicesKey handles keyboard input for the Services tab.
func (m *Model) handleServicesKey(msg tea.KeyMsg) tea.Cmd {
	if m.exposeEditing {
		switch msg.String() {
		case "esc":
			m.exposeEditing = false
			m.clearExposeInputs()
			return nil
		case "ctrl+s":
			m.loading = true
			m.loadingWhat = "Starting expose"
			m.exposeMsg = ""
			return m.doExposeService()
		case "tab":
			m.exposeFocused = (m.exposeFocused + 1) % 8
			m.focusExposeField()
			return nil
		default:
			return m.updateExposeInput(msg)
		}
	}

	switch msg.String() {
	case "n":
		m.exposeEditing = true
		m.exposeFocused = 0
		m.focusExposeField()
	case "x":
		m.stopExposeService()
	case "r":
		return m.fetchFwdRules()
	default:
		return m.handleGlobalKey(msg)
	}
	return nil
}

func (m *Model) clearExposeInputs() {
	m.exposePortInput.Blur()
	m.exposeProtocolInput.Blur()
	m.exposeExternalInput.Blur()
	m.exposeDomainInput.Blur()
	m.exposeNameInput.Blur()
	m.exposePasswordInput.Blur()
	m.exposePinInput.Blur()
	m.exposeGroupsInput.Blur()
}

func (m *Model) focusExposeField() {
	m.clearExposeInputs()
	switch m.exposeFocused {
	case 0:
		m.exposePortInput.Focus()
	case 1:
		m.exposeProtocolInput.Focus()
	case 2:
		m.exposeExternalInput.Focus()
	case 3:
		m.exposeDomainInput.Focus()
	case 4:
		m.exposeNameInput.Focus()
	case 5:
		m.exposePasswordInput.Focus()
	case 6:
		m.exposePinInput.Focus()
	case 7:
		m.exposeGroupsInput.Focus()
	}
}

func (m *Model) updateExposeInput(msg tea.KeyMsg) tea.Cmd {
	var c tea.Cmd
	switch m.exposeFocused {
	case 0:
		m.exposePortInput, c = m.exposePortInput.Update(msg)
	case 1:
		m.exposeProtocolInput, c = m.exposeProtocolInput.Update(msg)
	case 2:
		m.exposeExternalInput, c = m.exposeExternalInput.Update(msg)
	case 3:
		m.exposeDomainInput, c = m.exposeDomainInput.Update(msg)
	case 4:
		m.exposeNameInput, c = m.exposeNameInput.Update(msg)
	case 5:
		m.exposePasswordInput, c = m.exposePasswordInput.Update(msg)
	case 6:
		m.exposePinInput, c = m.exposePinInput.Update(msg)
	case 7:
		m.exposeGroupsInput, c = m.exposeGroupsInput.Update(msg)
	}
	return c
}

// nextLogLevel cycles log levels up or down.
func nextLogLevel(current proto.LogLevel, up bool) proto.LogLevel {
	order := []proto.LogLevel{
		proto.LogLevel_PANIC,
		proto.LogLevel_FATAL,
		proto.LogLevel_ERROR,
		proto.LogLevel_WARN,
		proto.LogLevel_INFO,
		proto.LogLevel_DEBUG,
		proto.LogLevel_TRACE,
	}
	for i, l := range order {
		if l == current {
			if up && i < len(order)-1 {
				return order[i+1]
			}
			if !up && i > 0 {
				return order[i-1]
			}
			return current
		}
	}
	return proto.LogLevel_INFO
}
