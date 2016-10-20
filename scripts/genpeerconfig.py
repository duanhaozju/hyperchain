#!/usr/bin/python
# -*- coding:utf-8 -*-
#
# generator the peer config file
import json

node = {"id":2,"address":"10.105.2.6","external_address":"115.159.26.15","port":8001,"rpc_port":8081}

server_list_files = open('serverlist.txt','rb')
inner_server_list_files = open('innerserverlist.txt','rb')

server_list = [line.strip() for line in server_list_files.readlines()]
inner_server_list = [ inner_line.strip() for inner_line in inner_server_list_files.readlines()]

# print server_list
# print inner_server_list

server_list_files.close()
inner_server_list_files.close()

nodes = []

for i in range (0,len(server_list)):
    # print server_list[i]
    # print inner_server_list[i]
    node = {"id":i+1,"address":inner_server_list[i],"external_address":server_list[i],"port":8001,"rpc_port":8081}
    nodes.append(node)
print nodes

peerconfig = {"maxpeernode":len(server_list),"nodes":nodes}

config_contents = json.dumps(peerconfig)

output = file("../config/peerconfig.json","wb")
output.write(config_contents)
output.close()