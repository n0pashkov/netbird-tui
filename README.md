# netbird-tui

A terminal UI for the local [NetBird](https://netbird.io/) daemon. It uses the daemon gRPC socket directly and keeps common monitoring and management tasks in one keyboard-driven interface.

## Features

- Two-level navigation with 4 groups and 11 screens.
- Monitor screens for status, peers, system events, and network map.
- Network screens for routes, DNS, and forwarding rules.
- Manage screens for profiles and setup-key login settings.
- Tools screens for diagnostics, packet tracing, daemon states, and service status.
- Context-aware footer actions, confirmation prompts for destructive operations, and quick switch/help overlays.
- Auto-refresh for daemon data with explicit refresh available from the UI.

## Screens

| Group | Screens |
| --- | --- |
| Monitor | Status, Peers, Events, Map |
| Network | Routes, DNS, Forwarding |
| Manage | Profiles, Settings |
| Tools | Diagnostics, Services |

## Requirements

- Go 1.25 or newer for building from source.
- A running NetBird daemon.
- Permission to access the daemon socket, usually `unix:///var/run/netbird.sock`.

On most Linux systems the socket is owned by root. Run with `sudo`, adjust group permissions, or pass a socket path that your user can read.

## Installation

Install with Go:

```bash
go install github.com/n0pashkov/netbird-tui@latest
```

Build from source:

```bash
git clone https://github.com/n0pashkov/netbird-tui
cd netbird-tui
go build -o netbird-tui .
```

## Usage

Use the default socket:

```bash
sudo netbird-tui
```

Use a custom socket:

```bash
netbird-tui unix:///custom/path/netbird.sock
```

Default socket path: `unix:///var/run/netbird.sock`.

## Keybindings

| Key | Action |
| --- | --- |
| `Tab` / `Shift+Tab` | Switch Monitor, Network, Manage, Tools groups |
| `1` - `4` | Jump to a group |
| `Left` / `Right` | Switch screens inside the active group |
| `g` | Open quick switch overlay for all screens |
| `m` | Open Map from quick switch |
| `?` | Open contextual help |
| `/` | Search inside the current searchable screen |
| `r` | Refresh current daemon data |
| `c` | Connect or disconnect depending on current state |
| `u` / `d` | NetBird up / down with confirmation |
| `L` | Logout with confirmation |
| `q` / `ctrl+c` | Quit |

Screen-specific keys are shown in the footer. The footer only advertises actions available on the current screen.

## Search And Filters

- Peers: `/` searches by FQDN, IP, or connection status; `f` cycles all, online, offline, and relayed peers; `x` clears search.
- Events: `/` searches event text; `f` cycles severity filters; `Enter` opens event detail.

## Troubleshooting

- `Failed to connect to NetBird daemon`: verify the daemon is running and the socket path is correct.
- `permission denied`: run with `sudo` or grant your user access to `/var/run/netbird.sock`.
- Empty screens: refresh with `r`; some screens depend on daemon or management-server support.
- Settings config flags are read-only in this release. Use `netbird config set` for config changes.
- Services shows forwarding rules and local service status. Creating exposed services is not enabled in this release.

## Development

```bash
GOCACHE=/tmp/netbird-tui-gocache go test ./...
go vet ./...
go build -o /tmp/netbird-tui .
```

## Demo

A screenshot or terminal recording can be added to this section when publishing release assets.

## License

MIT
