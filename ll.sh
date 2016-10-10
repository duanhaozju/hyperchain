#!/usr/bin/env bash

set -e

if [ x"$1" == x"" ];then
	echo "error param"
	exit 1;
fi

if [ x"$1" == x"1" ] ;then
	echo "kill the bind port process"
	rm -rf /tmp/hyperchain/
		for((i=0;i<=4;i++))
		do
		    temp_port=`lsof -i :800$i | awk 'NR>=2{print $2}'`
		    if [ x"$temp_port" != x"" ];then
		        kill -9 $temp_port
		    fi
		done
fi

clear

./hyperchain -o $1 -l 808$1 -p ./p2p/peerconfig.json -f ./consensus/pbft/ -g ./core/genesis.json

