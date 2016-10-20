#!/usr/bin/python
# -*- coding:utf-8 -*-
#
# generator the peer config file

node = {"id":2,"address":"10.105.2.6","external_address":"115.159.26.15","port":8001,"rpc_port":8081}

server_list = open('serverlist.txt','rb')
inner_server_list = open('innerserverlist.txt','rb')
