#!/bin/bash

# Check if all required arguments are provided
if [ $# -ne 4 ]; then
    echo "Usage: $0 <netip> <hostip> <netport> <hostport>"
    exit 1
fi

# Assigning arguments to variables
netip=$1
hostip=$2
netport=$3
hostport=$4

echo "netip: $netip"
echo "hostip: $hostip"
echo "netport: $netport"
echo "hostport: $hostport"

# Execute commands
ipvsadm -D -t "$netip":"$netport"
iptables -t nat -D POSTROUTING -m ipvs --vaddr "$netip" --vport "$netport" -j MASQUERADE
iptables -t nat -D PREROUTING -p tcp -d "$hostip" --dport "$hostport" -j DNAT --to-destination "$netip":"$netport"
ip addr del "$netip"/24 dev flannel.1