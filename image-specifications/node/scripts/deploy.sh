#!/bin/sh
modprobe bonding
modprobe 8021q

vdsm-tool configure --force
service vdsmd restart

# setting link name - vdsm recognize veth_* prefix as veth type device
ip link set eth0 down
ip link set eth0 name veth_name0
ip link set veth_name0 up
sleep 10
ip route add default via $CLUSTER_GATEWAY dev veth_name0
python /root/add_network.py

# ovirt registration flow.
MYADDR=$(ip address show dev veth_name0 | grep 'inet ' | xargs | cut -d' ' -f2 | cut -d/ -f1)
MYNAME=$(hostname -f)
vdsm-tool register --engine-fqdn $ENGINE_FQDN --check-fqdn false --node-address $MYADDR --node-name $MYNAME
