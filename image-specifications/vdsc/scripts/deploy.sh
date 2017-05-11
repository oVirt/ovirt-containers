#!/bin/sh

# /dev is mounted from the host so we have to fix permissions at runtime
chown root:kvm /dev/kvm
chmod 0660 /dev/kvm

# loading kernel modules for networking
modprobe bonding
modprobe 8021q
modprobe ebtables
modprobe etables_nat

mkdir -p /var/log/vdsm
mkdir -p /var/log/ovirt-imageio-daemon
chown vdsm:kvm /var/log/vdsm /var/log/ovirt-imageio-daemon

vdsm-tool configure --force

# setting link name - vdsm recognize veth_* prefix as veth type device
ip link set eth0 down
ip link set eth0 name veth_name0
ip link set veth_name0 up

sleep 5

ip route add default via 172.17.0.1 dev veth_name0
service vdsmd restart

sleep 5

python /root/add_network.py
