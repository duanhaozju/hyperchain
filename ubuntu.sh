#!/bin/bash
echo -e " _   _                        ____ _           _       "
echo -e "| | | |_   _ _ __   ___ _ __ / ___| |__   __ _(_)_ __  "
echo -e "| |_| | | | | '_ \ / _ \ '__| |   | '_ \ / _\` | | '_ \ "
echo -e "|  _  | |_| | |_) |  __/ |  | |___| | | | (_| | | | | |"
echo -e "|_| |_|\__, | .__/ \___|_|   \____|_| |_|\__,_|_|_| |_|"
echo -e "       |___/|_|                                        "
echo "usage:"
echo "$0 1 : 本地测试"
echo "$0 2 : 服务器测试"
echo "$0 3 : 杀死进程"
echo "你的选择: $1"
echo "--------------"
# Stop on first error
set -e

# max test node number
MAXNODE=4
PASSWD="blockchain"

#kill the progress
kellprogress(){

    echo "kill the bind port process"
    for((i=1;i<=$MAXNODE;i++))
    do
        temp_port=`lsof -i :800$i | awk 'NR>=2{print $2}'`
        if [ x"$temp_port" != x"" ];then
            kill -9 $temp_port
        fi
    done
}

local_test(){
    kellprogress
    CONFIG_PATH=$1
    #rebuild the application
    echo "rebuild the application"
    govendor build


    echo "remove tmp data"
    rm -rf /tmp/hyperchain/*

    echo "running the application"

    for((j=1;j<=$MAXNODE;j++))
    do
        gnome-terminal -x bash -c "(./hyperchain -o $j -l 808$j -p $CONFIG_PATH)"
    done

    python ./jsonrpc/Dashboard/simpleHttpServer.py

    echo "All process are running background"

}
########################
#   server side test   #
########################
server_test(){
if [ ! -f "/usr/bin/expect" ];then
  echo "hasn't install expect,please install expect mannualy: 'apt-get install expect'"
  exit 1
fi
#SERVER_ADDR=( cn6.hyperchain.cn cn7.hyperchain.cn cn8.hyperchain.cn cn9.hyperchain.cn cn12.hyperchain.cn cn13.hyperchain.cn cn14.hyperchain.cn)
# get the server list config
while read line;do
   SERVER_ADDR+=" ${line}"
done < ./serverlist.txt

  ni=1
  for server_address in ${SERVER_ADDR[@]}; do
    i=`expr $ni + 1`
    # gnome-terminal -x bash -c "ssh satoshi@$server_address \"cd /home/satoshi/gopath/src/hyperchain/scripts/ && ./server.sh $i\""
    gnome-terminal -x bash -c "ssh satoshi@$server_address \"source /home/satoshi/.profile && cd /home/satoshi/gopath/src/hyperchain/scripts/ && ./server.sh $ni\""
  done
}

#set -x
(( $# != 1 )) && { echo -e >&2 "Usage: $0 1 : 本地测试, $0 2 : 服务器测试"; exit 1; }
CONFIG_PATH="./p2p/local_peerconfig.json"
if [ $1 == 1 ];then
    echo -e "开启本地测试"
    local_test "./p2p/local_peerconfig.json"
elif [ $1 == 2 ];then
    echo -e "开启服务器测试"
    server_test
elif [ $1 == 3 ];then
    echo -e "杀死所有进程"
    kellprogress
    exit
else
    echo "参数错误"
    exit 1
fi

python ./jsonrpc/Dashboard/simpleHttpServer.py
