# Changelog

## v1.1.1

- Fixed the Go module path so `go install github.com/n0pashkov/netbird-tui@latest` works correctly.

## v1.1.0

- Reworked navigation around Monitor, Network, Manage, and Tools groups.
- Added a two-level tab bar, quick switch overlay, contextual help overlay, and dedicated network map screen.
- Made footers context-aware and removed unavailable Settings and Services actions from the UI.
- Improved Peers with persistent search, online/offline/relayed filters, and detail selection based on the filtered list.
- Added Events search while preserving severity filtering and detail view behavior.
- Added packet tracer input validation before daemon calls.
- Added unit coverage for navigation, filtering, helpers, trace validation, footer actions, and confirmation behavior.
- Added baseline CI for tests, vet, and build.
- Updated README for the current 11-screen layout, socket usage, troubleshooting, and installation paths.

## v1.0.0

- Initial terminal UI for NetBird daemon status, peers, routes, forwarding, and settings workflows.
