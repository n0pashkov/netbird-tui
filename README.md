# netbird-tui

A terminal user interface (TUI) for managing [NetBird](https://netbird.io/) — a WireGuard-based mesh VPN.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and communicates directly with the local NetBird daemon via gRPC.

## Features

- **Status tab** — management/signal state, local peer info, relay status
- **Peers tab** — table of all peers with connection status, latency, relay info, and traffic stats
- **Routes tab** — network routes with select/deselect toggle
- **Forwarding tab** — forwarding rules (protocol, destination and translated ports/addresses)
- **Settings tab** — configure setup key and management URL, trigger login
- NetBird Up / Down with confirmation prompt
- Logout with confirmation
- Debug bundle creation
- Auto-refresh every 5 seconds
- Adaptive layout — tables resize with the terminal window

## Requirements

- NetBird daemon running locally (`netbird` service)
- Access to the NetBird socket (usually `/var/run/netbird.sock`)

## Installation

```bash
git clone https://github.com/justniklab/netbird-tui
cd netbird-tui
go build -o netbird-tui .
sudo ./netbird-tui
```

Or run directly:

```bash
go run . [socket-path]
```

Default socket path: `unix:///var/run/netbird.sock`

## Usage

```bash
# Use default socket
sudo netbird-tui

# Custom socket path
sudo netbird-tui unix:///custom/path/netbird.sock
```

## Keybindings

| Key | Action |
|-----|--------|
| `←` / `→` | Switch tabs |
| `1` – `5` | Jump to tab directly |
| `↑` / `↓` / `j` / `k` | Navigate table rows |
| `u` | NetBird Up (with confirmation) |
| `d` | NetBird Down (with confirmation) |
| `L` | Logout (with confirmation) |
| `b` | Create debug bundle (with confirmation) |
| `r` | Refresh all data |
| `Enter` | Toggle route selected/deselected (Routes tab) |
| `Tab` / `Enter` | Next input field (Settings tab) |
| `ctrl+s` | Submit settings / login (Settings tab) |
| `esc` | Cancel / go back (Settings tab) |
| `q` / `ctrl+c` | Quit |

## Tabs

| # | Tab | Description |
|---|-----|-------------|
| 1 | Status | Connection status, peer info, relays |
| 2 | Peers | All peers with stats |
| 3 | Routes | Network routes, toggle selection |
| 4 | Forwarding | Port forwarding rules |
| 5 | Settings | Setup key, management URL, login |

## License

MIT
