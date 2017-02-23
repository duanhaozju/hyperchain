#!/bin/bash
rm -rf build db.log
cp ../node4/hyperchain ./
cp ../node4/config/local_peerconfig.json ./config/local_peerconfig.json
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.node_id 5 -t int -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.grpc_port 8005 -t int -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.jsonrpc_port 8085 -t int -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.introducer_id 1 -t int -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.introducer_port 8001 -t int -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.introducer_rpcport 8081 -t int -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json self.is_origin false -t bool -y
confer write ./config/local_peerconfig.json ./config/local_peerconfig.json maxpeernode 5 -t int -y
confer write ./config/pbft.yaml ./config/pbft.yaml pbft.nodes 4 -t int -y
./hyperchain -o 5 -l 8005 -t 8085