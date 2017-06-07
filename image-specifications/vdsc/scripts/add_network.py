#!/bin/python
from vdsm import client
import socket

c = client.connect('localhost')
ip_addr = socket.gethostbyname(socket.gethostname())
gateway_addr = '{0}.{1}'.format('.'.join(ip_addr.split('.')[:3]), '1')
network_attrs = {'nic': 'veth_name0',
                 'ipaddr': ip_addr,
                 'netmask': '255.240.0.0',
                 'gateway': gateway_addr,
                 'defaultRoute': True,
                 'bridged': True}

c.Host.setupNetworks(
    networks={'ovirtmgmt': network_attrs},
    bondings={},
    options={'connectivityCheck': False}
)
