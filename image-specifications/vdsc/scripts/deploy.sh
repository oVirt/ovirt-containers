#!/bin/sh

# /dev is mounted from the host so we have to fix permissions at runtime
chown root:kvm /dev/kvm
chmod 0660 /dev/kvm

# loading kernel modules for networking
modprobe bonding
modprobe 8021q
modprobe ebtables

mkdir -p /var/log/vdsm
mkdir -p /var/log/ovirt-imageio-daemon
chown vdsm:kvm /var/log/vdsm /var/log/ovirt-imageio-daemon

vdsm-tool configure --force

IFNAME="veth_name0"

# setting link name - vdsm recognize veth_* prefix as veth type device
ip link set eth0 down
ip link set eth0 name $IFNAME
ip link set $IFNAME up

# Configure routes
IP_ADDRESS=$(ip addr show $IFNAME | grep "inet\b" | awk '{print $2}' | cut -d/ -f1)
IFS=. read ip_oct_1 ip_oct_2 ip_oct_3 ip_oct_4 <<< "$IP_ADDRESS"

DEFAULT_ROUTE="$ip_oct_1.$ip_oct_2.$ip_oct_3.1"
CLUSTER_ROUTE="10.128.0.0/14"
MULTICAST_ROUTE="224.0.0.0/4"

ip route add default via $DEFAULT_ROUTE dev $IFNAME
ip route add $CLUSTER_ROUTE dev $IFNAME scope global
ip route add $MULTICAST_ROUTE dev $IFNAME scope global

sleep 5

service vdsmd restart

sleep 5

python /root/add_network.py

# ovirt registration flow.
declare $(xargs -n 1 -0 < /proc/1/environ | grep MY_NODE_NAME)
MYNAME=$(hostname -f)
SSHPORT=22222
ENGINE_SERVICE=ovirt-engine.ovirt.svc
ENGINE_HTTPS_PORT=443
ENGINE_FQDN=$(echo "Q" | openssl s_client -connect $ENGINE_SERVICE:$ENGINE_HTTPS_PORT 2>&1 | grep "subject=" | sed 's/^.*CN=//g')
vdsm-tool register --engine-fqdn $ENGINE_FQDN \
                   --engine-https-port $ENGINE_HTTPS_PORT \
                   --check-fqdn false \
                   --node-name $MYNAME \
                   --node-address $MY_NODE_NAME \
                   --ssh-port $SSHPORT
