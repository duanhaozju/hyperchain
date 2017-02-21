#!/bin/bash
rm -rf build db.log
cp ../node4/hyperchain ./
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.node_id ${1} -t int -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.grpc_port 800${1} -t int -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.jsonrpc_port 808${1} -t int -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.introducer_id 1 -t int -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.introducer_port 8001 -t int -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.introducer_rpcport 8081 -t int -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.is_origin false -t bool -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json maxpeernode ${1} -t int -y
./hyperchain -o ${1} -l 800${1} -t 808${1}
