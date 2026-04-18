# InstaGuard 🛡️

A FortiGate-inspired firewall OS built from scratch on Debian 13.

**"Instant protection for your network."**

## Status: Milestone 1 ✅ — Base router working

### Network topology
- WAN interface: enp1s0 (upstream NAT, DHCP client)
- LAN interface: enp2s0 (192.168.81.1/24)
- LAN DHCP pool: 192.168.81.100 – 192.168.81.200

### Services running
- **nftables** — stateful firewall + NAT masquerade
- **isc-dhcp-server** — DHCP for LAN clients
- **dnsmasq** — forwarding DNS resolver (8.8.8.8, 1.1.1.1)
- **ssh** — management access

### Validated
- Alpine Linux client gets DHCP lease from InstaGuard
- Full internet reachability via NAT masquerade
- DNS resolution via dnsmasq
- Stateful connection tracking via conntrack

## Architecture

InstaGuard is built on:
- **Base OS:** Debian 13 (Trixie) with Linux kernel 6.12 LTS
- **Packet filter:** nftables (Linux kernel native)
- **Management:** Go-based REST API + React web UI (Milestone 2+)

## Roadmap

- [x] **Milestone 1** — Base router (nftables + DHCP + DNS)
- [ ] **Milestone 2** — Go config engine + REST API
- [ ] **Milestone 3** — Web UI + CLI
- [ ] **Milestone 4** — IPS (Suricata), VPN (WireGuard), BGP/OSPF (FRR)
- [ ] **Milestone 5** — ISO packaging with Debian live-build
- [ ] **Milestone 6** — v1.0 release

## Author

Built by **Raushan-Nexapp** — learning by building.
# InstaGuard
# InstaGuard
