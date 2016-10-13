#!/usr/bin/env bash
echo "┌─────────────────────────────────┐"
echo "│    LOCAL TEST FOR HYPERCHAIN    │"
echo "└─────────────────────────────────┘"

# Stop on first error
set -e

#set -x


echo "kill the bind port process"
ports1=`lsof -i :8000 | awk 'NR>=2{print $2}'`
if [ x"$ports1" != x"" ];then
    kill -9 $ports1
fi
ports1=`lsof -i :8001 | awk 'NR>=2{print $2}'`
if [ x"$ports1" != x"" ];then
    kill -9 $ports1
fi
ports2=`lsof -i :8002 | awk 'NR>=2{print $2}'`
if [ x"$ports2" != x"" ];then
    kill -9 $ports2
fi
ports3=`lsof -i :8003 | awk 'NR>=2{print $2}'`
if [ x"$ports3" != x"" ];then
    kill -9 $ports3
fi
ports4=`lsof -i :8004 | awk 'NR>=2{print $2}'`
if [ x"$ports4" != x"" ];then
    kill -9 $ports4
fi

#rebuild the application
echo "rebuild the application"
govendor build

echo "delete the database"
rm -rf /tmp/hyperchain/

echo "run the application"
osascript -e 'tell app "Terminal" to do script "cd $GOPATH/src/hyperchain && ./hyperchain -o 1 -l 8081 -p ./p2p/peerconfig.json -f ./consensus/pbft/ -g ./core/genesis.json"'
osascript -e 'tell app "Terminal" to do script "cd $GOPATH/src/hyperchain && ./hyperchain -o 2 -l 8082 -p ./p2p/peerconfig.json -f ./consensus/pbft/ -g ./core/genesis.json"'
osascript -e 'tell app "Terminal" to do script "cd $GOPATH/src/hyperchain && ./hyperchain -o 3 -l 8083 -p ./p2p/peerconfig.json -f ./consensus/pbft/ -g ./core/genesis.json"'
osascript -e 'tell app "Terminal" to do script "cd $GOPATH/src/hyperchain && ./hyperchain -o 4 -l 8084 -p ./p2p/peerconfig.json -f ./consensus/pbft/ -g ./core/genesis.json"'


python ./jsonrpc/Dashboard/simpleHttpServer.py

echo "All process are running background"
