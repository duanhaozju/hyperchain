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
PRIMARY=`head -1 ./innerserverlist.txt`

MAXNODE=`cat innerserverlist.txt | wc -l`

#SERVER_LIST_CONTENT=`cat ./serverlist.txt`
while IFS='' read -r line || [[ -n "$line" ]]; do
	SERVER_ADDR+=" ${line}"
done < innerserverlist.txt

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

add_ssh_key_form_primary_to_others(){
    echo "primary add its ssh key into others nodes"
	scp ./sub_scripts/deploy/server_addkey.sh satoshi@$PRIMARY:/home/satoshi/
	scp innerserverlist.txt satoshi@$PRIMARY:/home/satoshi/

	COMMANDS="cd /home/satoshi && chmod a+x server_addkey.sh && bash server_addkey.sh"

	ssh  satoshi@$PRIMARY $COMMANDS
}

build(){
	echo "Compiling and generating the configuration files..."
    cd ..
    source ~/.bash_profile
    govendor build -o build/hyperchain
    cd scripts
}

distribute_the_binary(){
    echo "Sending the local complied file and configuration files to primary"
	cp -f ../build/hyperchain /home/satoshi
	cp -f -r ../config/ /home/satoshi/

	cp -f innerserverlist.txt /home/satoshi/
	cp -f ./sub_scripts/deploy/server_deploy.sh /home/satoshi/

	cd /home/satoshi && chmod a+x server_deploy.sh && bash server_deploy.sh $MAXNODE
	echo "finish"
}

if $FIRST_RUN; then
    add_ssh_key_into_all
	add_ssh_key_form_primary_to_others
fi

build
distribute_the_binary