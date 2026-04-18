# InstaGuard

A FortiGate-inspired firewall OS built from scratch on Debian 13.

**"Instant protection for your network."**

## Status

- [x] Milestone 1 — Base router (nftables + DHCP + DNS)
- [x] Milestone 2 — Go config engine + REST API
- [ ] Milestone 3 — Web UI + CLI
- [ ] Milestone 4 — IPS (Suricata), VPN (WireGuard), BGP/OSPF (FRR)
- [ ] Milestone 5 — ISO packaging with Debian live-build
- [ ] Milestone 6 — v1.0 release

## Architecture

Request flow: curl -> REST API (Go) -> SQLite -> template -> nftables.conf -> Linux kernel

- Base OS: Debian 13 Trixie, Linux kernel 6.12 LTS
- Packet filter: nftables (kernel-native)
- Management: Go engine, single 8 MB binary
- Config store: SQLite
- API: RESTful HTTP on port 8080

## API endpoints

System:
- GET /api/v1/status — engine health
- GET /api/v1/stats — row counts

Policies:
- GET /api/v1/policies — list all
- POST /api/v1/policies — create
- GET /api/v1/policies/{id} — read one
- PUT /api/v1/policies/{id} — update
- DELETE /api/v1/policies/{id} — delete

Apply:
- POST /api/v1/apply — regenerate nftables.conf (dry-run)
- POST /api/v1/apply?commit=true — regenerate and reload live firewall

## Build and run

cd engine
go build -o bin/instaguard-engine ./cmd/instaguard-engine
./bin/instaguard-engine

## Author

Built by Raushan — learning by building.
