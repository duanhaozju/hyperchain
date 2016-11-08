#!/usr/bin/env bash
# this script is used to auto deploy the compiled binary code.
# and auto run the predefined command.
# Author: Chen Quan
# Update Date: 2016-10-19
# Features:
# 1. auto add the ssh key into the primary sever
# 2. auto add the primary's ssh key charo the non-primary server
# 3. accelerate the distributes speed
# 4. auto read the server list file and auto run the suit command

# stop when error
#set -e
#set -x
#parse the arguments
# some arguments don't have a corresponding value to go with it such
# as in the --default example).
# note: if this is set to -gt 0 the /etc/hosts part is not recognized ( may be a bug )

FIRST_RUN=false
TEST_FLAG=true
PASSWD="blockchain"
TEST_MACHINE="115.159.27.148"

MAXNODE=`cat serverlist.txt | wc -l`

#SERVER_LIST_CONTENT=`cat ./serverlist.txt`
while IFS='' read -r line || [[ -n "$line" ]]; do
	SERVER_ADDR+=" ${line}"
done < serverlist.txt

echo $SERVER_ADDR

while [[ $# -ge 1 ]]
do
key="$1"
case $key in
    -f|--first)
    FIRST_RUN=true
    shift # past argument
    ;;
    -t|--test)
    TEST_FLAG=false
    shift # past argument
    ;;
    *)
    shift # past argument or value
    ;;
esac
done

#echo $FIRST_RUN
#echo $TEST_FLAG

#deploy_dir=`pwd`"/../deploy"
#if [ ! -d $deploy_dir ];then
#    mkdir -p $deploy_dir
#fi


addkey(){
  expect <<EOF
      set timeout 60
      spawn ssh-copy-id satoshi@$1
      expect {
        "yes/no" {send "yes\r";exp_continue }
        "s password:" {send "$PASSWD\r";exp_continue }
        eof
      }
EOF
}

add_ssh_key_into_all(){
    echo "add your local ssh public key into primary node"
    for server_address in ${SERVER_ADDR[@]}; do
	  addkey $server_address
	done
}

ni=1
auto_run(){
    echo "Auto start all nodes"
    for server_address in ${SERVER_ADDR[@]}; do
	  echo $server_address
	  ssh satoshi@$server_address "ps aux | grep hyperchain | awk '{print \$2}' | xargs kill -9"
      ssh satoshi@$server_address "if [ ! -d /home/satoshi/build/ ]; then mkdir -p /home/satoshi/build/;fi"
	  gnome-terminal -x bash -c "ssh satoshi@$server_address \" cd /home/satoshi/ && cp -rf ./config/keystore ./build/ && ./hyperchain -o $ni -l 8001 -t 8081 || while true; do ifconfig && sleep 100; done\""
	  ni=`expr $ni + 1`
	done
}

scp_binary_to_test_machine(){
    cd $GOPATH/src/ && rm hyperchain.zip && zip -r hyperchain.zip ./hyperchain
    scp hyperchain.zip satoshi@$TEST_MACHINE:/home/satoshi/
    ssh satoshi@$TEST_MACHINE "rm -rf go/src/hyperchain && unzip -o hyperchain.zip -d /home/satoshi/go/src/"
    ssh satoshi@$TEST_MACHINE "cd /home/satoshi/go/src/hyperchain/scripts && bash centos_server.sh $1"
}

if $FIRST_RUN; then
    addkey $TEST_MACHINE
	add_ssh_key_into_all
	scp_binary_to_test_machine "-f"
else
    scp_binary_to_test_machine
fi

auto_run