#!/bin/bash
if [ 'x'$1 == 'x1' ];then
	echo 'clear database'
	rm -rf /tmp/hyperchain/
fi

./hyperchain -o $1 -l 808$1 -p p2p/local_peerconfig.json ./ -g ./genesis.json