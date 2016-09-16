#!/bin/bash
echo "┌─────────────────────────────────┐"
echo "│    LOCAL TEST FOR HYPERCHAIN    │"
echo "└─────────────────────────────────┘"

# Stop on first error
set -e

#set -x
(( $# != 1 )) && { echo -e >&2 "Usage: $0 1 : 本地测试, $0 2 : 服务器测试"; exit 1; }
CONFIG_PATH="./p2p/local_peerconfig.json"
if [ $1 == 1 ];then
    echo -e "开启本地测试"
    CONFIG_PATH="./p2p/local_peerconfig.json"
elif [ $1 == 2 ];then
    echo -e "开启服务器测试"
    CONFIG_PATH="./p2p/peerconfig.json"
else
    echo "参数错误"
    exit 1
fi
# max test node number
MAXNODE=7

echo "kill the bind port process"
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
ports5=`lsof -i :8005 | awk 'NR>=2{print $2}'`
if [ x"$ports4" != x"" ];then
    kill -9 $ports4
fi
ports6=`lsof -i :8006 | awk 'NR>=2{print $2}'`
if [ x"$ports4" != x"" ];then
    kill -9 $ports4
fi
ports7=`lsof -i :8007 | awk 'NR>=2{print $2}'`
if [ x"$ports4" != x"" ];then
    kill -9 $ports4
fi

#rebuild the application
echo "rebuild the application"
govendor build


echo "remove tmp data"
rm -rf /tmp/hyperchain/*

echo "running the application"

for((i=1;i<=$MAXNODE;i++))
do
    gnome-terminal -x bash -c "(./hyperchain -o $i -l 808$i -p $CONFIG_PATH)"
done

echo "All process are running background"