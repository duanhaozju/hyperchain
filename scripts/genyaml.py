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

# global config
# this section should give some params

glb = {"global":{}}
nodeid = {'nodeid':1}
grpcport = {'grpcprort':8001}
httpport = {'httpport':8081}

glb['global'].update(nodeid)
glb['global'].update(grpcport)
glb['global'].update(httpport)
#print glb


# data storage config
# this section should not modified!
account = {'keystoredir':'"config/keystore"','keynodesdir':'"./build/keynodes"'}

logs = {'logs':{'dumpfile':True,'logsdir':'"./build/logs"','logllevel':'NOTICE'}}

database = {'database':{'dir':'"./build/database"'}}
glb['global'].update(account)
glb['global'].update(logs)
glb['global'].update(database)

#print glb
# configs section, this section should use default configs
configs = {'configs':{}}
peers = {'peers':'"configs/peerconfig.json"'}
genesis = {'genesis':'"config/genesis.json"'}
memebersrvc = {'membersrvc':'"config/membersrvc.yaml"'}
pbft = {'pbft':'"config/pbft.yaml"'}
configs['configs'].update(peers)
configs['configs'].update(genesis)
configs['configs'].update(memebersrvc)
configs['configs'].update(pbft)

glb["global"].update(configs)
#print glb


#print yaml.dump(yaml.load(glb),default_flow_style=False)
print yaml.dump(glb,default_flow_style=False)
