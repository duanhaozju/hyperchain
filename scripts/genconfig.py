#!/usr/bin/python
# -*- coding:utf-8 -*-

import yaml

##
# 这里请加上 default_flow_style=False
# 否则将以压缩的方式输出嵌套对象
# document = """
#   a: 1
#   b:
#    c: 3
#    d: 4
# """
# 将输出
# a: 1
# b: {c: 3, d: 4}
#
# 请加上 default_flow_style=False
# print yaml.dump(yaml.load(document), default_flow_style=False)
class PeerConfig:
    def _init_(self):
        self.nodeID = 1
        self.localGRPCPort=8001
        self.localJsonRPCPort=8081
        self.localIP="127.0.0.1"
    def genPeerConfig(self,node_ID=1,local_gRPC_Port=8001,local_JsonRPC_Port=8081):
        pass
    def readServerList(self,server_list_path):
        server_list_file = open(server_list_path,"rb")
        for line in server_list_file.readlines():
            print line.strip()
        server_list_file.close()

        
if __name__ == '__main__':
    peerConfig = PeerConfig()
    peerConfig.readServerList("server.txt")