#!/bin/sh
modprobe bonding
modprobe 8021q

mkdir -p /var/log/vdsm
mkdir -p /var/log/ovirt-imageio-daemon
chown vdsm:kvm /var/log/vdsm /var/log/ovirt-imageio-daemon

vdsm-tool configure --force
service vdsmd restart

# setting link name - vdsm recognize veth_* prefix as veth type device
ip link set eth0 down
ip link set eth0 name veth_name0
ip link set veth_name0 up
sleep 10
ip route add default via $CLUSTER_GATEWAY dev veth_name0
python /root/add_network.py

# /dev is mounted from the host so we have to fix permissions at runtime
chown root:kvm /dev/kvm
chmod 0660 /dev/kvm

# ovirt registration flow.
MYNAME=$(hostname -f)
SSHPORT=22222
vdsm-tool register --engine-fqdn $ENGINE_FQDN --check-fqdn false \
	--node-name $MYNAME --node-address $MY_NODE_NAME --ssh-port $SSHPORT
