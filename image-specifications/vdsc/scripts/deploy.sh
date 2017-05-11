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
