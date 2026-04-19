#!/bin/bash
# NexappOS — apply Milestone 1 configs
set -e

echo "🛡️  NexappOS — applying Milestone 1 configs..."

cp configs/nftables.conf /etc/nftables.conf
cp configs/dhcpd.conf    /etc/dhcp/dhcpd.conf
cp configs/dnsmasq.conf  /etc/dnsmasq.conf

systemctl restart nftables
systemctl restart isc-dhcp-server
systemctl restart dnsmasq

echo "✅  NexappOS Milestone 1 applied successfully"
