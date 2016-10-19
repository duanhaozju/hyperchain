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
PRIMARY=`head -1 ./serverlist.txt`

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
#  ssh-keygen -f "/home/fox/.ssh/known_hosts" -R $1
  expect <<EOF
      set timeout 60
      spawn ssh-copy-id satoshi@$1
      expect {
        "yes/no" {send "yes\r";exp_continue }
        "s password:" {send "$PASSWD\r";exp_continue }
        eof
      }
EOF

ssh -t satoshi@$1 "sudo -i && apt-get install -y expect"
}

add_ssh_key_into_primary(){
    echo "add your local ssh public key into primary node"
#    echo "将你的本地ssh公钥添加到primary中"
    for server_address in ${SERVER_ADDR[@]}; do
	  addkey $server_address &
	done
	wait

}

# count=0
#	while read line;do
#	    if [ $count -ne 0 ]; then
#	        # SERVER_ADDR+=" ${line}"
#	        echo $line
#
#
#
#	    fi
#	    let count=count+1
#	done < ./serverlist.txt

add_ssh_key_form_primary_to_others(){
    echo "primary add its ssh key into others nodes"
#    echo "primary 将它的ssh 公钥加入到其它节点中"
	scp -r ./sub_scripts/deploy/server_addkey.sh satoshi@$PRIMARY:/home/satoshi/

	COMMANDS="cd /home/satoshi && chmod a+x server_addkey.sh && bash server_addkey.sh"

	ssh  satoshi@$PRIMARY $COMMANDS

# count=0
#	while read line;do
#	    if [ $count -ne 0 ]; then
#	        # SERVER_ADDR+=" ${line}"
#	        ssh -t
#
#
#
#	    fi
#	    let count=count+1
#	done < ./innerlist.txt

}


build(){
	echo "Compiling and generating the configuration files..."
    cd ..
    govendor build -o build/hyperchain
    cd scripts
}


distribute_the_binary(){
    echo "Sending the local complied file and configuration files to primary"
	scp -r ../build/hyperchain satoshi@$PRIMARY:/home/satoshi/
	scp -r ../config/ satoshi@$PRIMARY:/home/satoshi/

	scp -r innerserverlist.txt satoshi@$PRIMARY:/home/satoshi/
	scp -r ./sub_scripts/deploy/killprocess.sh satoshi@$PRIMARY:/home/satoshi/
	scp -r ./sub_scripts/deploy/server_deploy.sh satoshi@$PRIMARY:/home/satoshi/

	ssh  satoshi@$PRIMARY "ps aux | grep hyperchain | awk '{print $2}' | xargs kill -9"
	ssh  satoshi@$PRIMARY "chmod a+x server_deploy.sh && bash server_deploy.sh ${MAXNODE}"
}

#clean(){
#    echo "清除本地生成的文件(build 文件夹)"
#    rm -rf $deploy_dir
#}

#read_server_list(){
#    echo "读取server_list"
#    while read line; do
#        SERVER_ADDR+=" ${line}"
#    done < serverlist.txt < "\n"
#    echo $SERVER_ADDR
#}

ni=1
auto_run(){
    echo "Auto start all nodes"
#    echo "自动运行相应命令，启动全节点"
    for server_address in ${SERVER_ADDR[@]}; do
	  echo $server_address

	  gnome-terminal -x bash -c "ssh satoshi@$server_address \" cd /home/satoshi/ && ./hyperchain -o $ni -l 8001 -t 8081 || while true; do ifconfig && sleep 100; done\""
	  ni=`expr $ni + 1`
	done
}

if $FIRST_RUN; then
	add_ssh_key_into_primary
	add_ssh_key_form_primary_to_others
fi

build
distribute_the_binary
auto_run