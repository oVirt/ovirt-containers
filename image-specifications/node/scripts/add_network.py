#!/bin/python
from vdsm import vdscli
import socket

c = vdscli.connect()
network_attrs = {'nic': 'veth_name0',
                 'ipaddr': socket.gethostbyname(socket.gethostname()),
                 'netmask': '255.240.0.0',
                 'gateway': '172.17.0.1',
                 'defaultRoute': True,
                 'bridged': True}

c.setupNetworks({'ovirtmgmt': network_attrs}, {}, {'connectivityCheck': False})
