package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderSettings(m *Model) string {
	if m.settingsPage == 1 {
		return renderConfigFlags(m)
	}
	return renderLoginSettings(m)
}

func renderLoginSettings(m *Model) string {
	var sb strings.Builder

	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Render("Settings") +
		"  " + styleNeutral.Render("PgDn/PgUp: switch pages  [1/2] Login") + "\n\n")

	// Current config info (read-only overview)
	if m.config != nil {
		sb.WriteString(styleSectionHeader.Render("Current Configuration") + "\n")
		lbl := lipgloss.NewStyle().Foreground(colorGray).Width(26)
		val := styleValue

		sb.WriteString("  " + lbl.Render("Management URL:") + val.Render(m.config.ManagementUrl) + "\n")
		if m.config.AdminURL != "" {
			sb.WriteString("  " + lbl.Render("Admin URL:") + val.Render(m.config.AdminURL) + "\n")
		}
		if m.config.ConfigFile != "" {
			sb.WriteString("  " + lbl.Render("Config file:") + val.Render(m.config.ConfigFile) + "\n")
		}
		if m.config.LogFile != "" {
			sb.WriteString("  " + lbl.Render("Log file:") + val.Render(m.config.LogFile) + "\n")
		}
		if m.config.InterfaceName != "" {
			sb.WriteString("  " + lbl.Render("Interface:") + val.Render(m.config.InterfaceName) + "\n")
		}
		if m.config.WireguardPort > 0 {
			sb.WriteString("  " + lbl.Render("WireGuard port:") + val.Render(fmt.Sprintf("%d", m.config.WireguardPort)) + "\n")
		}
		if m.config.Mtu > 0 {
			sb.WriteString("  " + lbl.Render("MTU:") + val.Render(fmt.Sprintf("%d", m.config.Mtu)) + "\n")
		}
		sb.WriteString("\n")

		// Quick flags row
		sb.WriteString("  ")
		flags := []struct {
			label string
			val   bool
		}{
			{"Rosenpass", m.config.RosenpassEnabled},
			{"Auto-connect", !m.config.DisableAutoConnect},
			{"SSH server", m.config.ServerSSHAllowed},
			{"Network monitor", m.config.NetworkMonitor},
			{"Routes", !m.config.DisableClientRoutes},
			{"DNS", !m.config.DisableDns},
		}
		for i, f := range flags {
			if i > 0 {
				sb.WriteString("  ")
			}
			if f.val {
				sb.WriteString(styleOnline.Render("● "))
			} else {
				sb.WriteString(styleNeutral.Render("○ "))
			}
			sb.WriteString(styleNeutral.Render(f.label))
		}
		sb.WriteString("\n\n")
	}

	// Login / Setup key section
	sb.WriteString(styleSectionHeader.Render("Login with Setup Key") + "\n")

	skiLabel := labelForField(0, m.settingsFocused, m.settingsEditing, "Setup Key:")
	sb.WriteString(skiLabel + "\n")
	sb.WriteString("  " + m.setupKeyInput.View() + "\n\n")

	muiLabel := labelForField(1, m.settingsFocused, m.settingsEditing, "Management URL:")
	sb.WriteString(muiLabel + "\n")
	sb.WriteString("  " + m.mgmtURLInput.View() + "\n\n")

	// Status message
	if m.settingsMsg != "" {
		if strings.HasPrefix(m.settingsMsg, "Error") {
			sb.WriteString(styleError.Render(m.settingsMsg) + "\n\n")
		} else {
			sb.WriteString(styleOnline.Render(m.settingsMsg) + "\n\n")
		}
	}

	if m.settingsEditing {
		sb.WriteString(styleNeutral.Render("ctrl+s:Login  Esc:Cancel  Tab:Next field") + "\n")
	} else {
		sb.WriteString(styleNeutral.Render("↑↓/Tab:Select  Enter:Edit  ctrl+s:Login  PgDn:Config flags page") + "\n")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

func renderConfigFlags(m *Model) string {
	var sb strings.Builder

	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Render("Settings") +
		"  " + styleNeutral.Render("PgDn/PgUp: switch pages  [2/2] Configuration") + "\n\n")

	if m.config == nil {
		sb.WriteString(styleNeutral.Render("Loading configuration..."))
		return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
	}

	// Full config display
	lbl := lipgloss.NewStyle().Foreground(colorGray).Width(30)

	sb.WriteString(styleSectionHeader.Render("Network") + "\n")
	sb.WriteString("  " + lbl.Render("Interface name:") + styleValue.Render(m.config.InterfaceName) + "\n")
	sb.WriteString("  " + lbl.Render("WireGuard port:") + styleValue.Render(fmt.Sprintf("%d", m.config.WireguardPort)) + "\n")
	sb.WriteString("  " + lbl.Render("MTU:") + styleValue.Render(fmt.Sprintf("%d", m.config.Mtu)) + "\n")
	sb.WriteString("  " + lbl.Render("Block inbound:") + boolFlag(m.config.BlockInbound) + "\n")
	sb.WriteString("  " + lbl.Render("Block LAN access:") + boolFlag(m.config.BlockLanAccess) + "\n")
	sb.WriteString("  " + lbl.Render("Network monitor:") + boolFlag(m.config.NetworkMonitor) + "\n")
	sb.WriteString("\n")

	sb.WriteString(styleSectionHeader.Render("Firewall & Routes") + "\n")
	sb.WriteString("  " + lbl.Render("Disable client routes:") + boolFlag(m.config.DisableClientRoutes) + "\n")
	sb.WriteString("  " + lbl.Render("Disable server routes:") + boolFlag(m.config.DisableServerRoutes) + "\n")
	sb.WriteString("\n")

	sb.WriteString(styleSectionHeader.Render("DNS") + "\n")
	sb.WriteString("  " + lbl.Render("Disable DNS:") + boolFlag(m.config.DisableDns) + "\n")
	sb.WriteString("\n")

	sb.WriteString(styleSectionHeader.Render("Connection") + "\n")
	sb.WriteString("  " + lbl.Render("Disable auto-connect:") + boolFlag(m.config.DisableAutoConnect) + "\n")
	sb.WriteString("  " + lbl.Render("Lazy connection:") + boolFlag(m.config.LazyConnectionEnabled) + "\n")
	sb.WriteString("  " + lbl.Render("Block inbound:") + boolFlag(m.config.BlockInbound) + "\n")
	sb.WriteString("  " + lbl.Render("Disable notifications:") + boolFlag(m.config.DisableNotifications) + "\n")
	sb.WriteString("\n")

	sb.WriteString(styleSectionHeader.Render("Security — Rosenpass") + "\n")
	sb.WriteString("  " + lbl.Render("Rosenpass enabled:") + boolFlag(m.config.RosenpassEnabled) + "\n")
	sb.WriteString("  " + lbl.Render("Rosenpass permissive:") + boolFlag(m.config.RosenpassPermissive) + "\n")
	sb.WriteString("\n")

	sb.WriteString(styleSectionHeader.Render("SSH Server") + "\n")
	sb.WriteString("  " + lbl.Render("SSH server allowed:") + boolFlag(m.config.ServerSSHAllowed) + "\n")
	sb.WriteString("  " + lbl.Render("SSH root access:") + boolFlag(m.config.EnableSSHRoot) + "\n")
	sb.WriteString("  " + lbl.Render("SFTP enabled:") + boolFlag(m.config.EnableSSHSFTP) + "\n")
	sb.WriteString("  " + lbl.Render("Local port forwarding:") + boolFlag(m.config.EnableSSHLocalPortForwarding) + "\n")
	sb.WriteString("  " + lbl.Render("Remote port forwarding:") + boolFlag(m.config.EnableSSHRemotePortForwarding) + "\n")
	sb.WriteString("\n")

	sb.WriteString(styleSectionHeader.Render("Paths") + "\n")
	sb.WriteString("  " + lbl.Render("Management URL:") + styleValue.Render(m.config.ManagementUrl) + "\n")
	if m.config.AdminURL != "" {
		sb.WriteString("  " + lbl.Render("Admin URL:") + styleValue.Render(m.config.AdminURL) + "\n")
	}
	sb.WriteString("  " + lbl.Render("Config file:") + styleValue.Render(m.config.ConfigFile) + "\n")
	sb.WriteString("  " + lbl.Render("Log file:") + styleValue.Render(m.config.LogFile) + "\n")

	sb.WriteString("\n" + styleNeutral.Render("PgUp: Login page  (Edit via CLI: netbird config set)"))

	if m.settingsMsg != "" {
		sb.WriteString("\n")
		if strings.HasPrefix(m.settingsMsg, "Error") {
			sb.WriteString(styleError.Render(m.settingsMsg))
		} else {
			sb.WriteString(styleOnline.Render(m.settingsMsg))
		}
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(sb.String())
}

func boolFlag(b bool) string {
	if b {
		return styleOnline.Render("Yes")
	}
	return styleOffline.Render("No")
}

// labelForField returns the styled label for a settings field.
func labelForField(fieldIdx, focused int, editing bool, label string) string {
	if fieldIdx != focused {
		return styleNeutral.Render(label)
	}
	if editing {
		return lipgloss.NewStyle().Foreground(colorBlue).Bold(true).Render("✎ " + label)
	}
	return lipgloss.NewStyle().Foreground(colorBlue).Bold(true).Render("▶ " + label)
}
