#!/usr/bin/env bash

FIRST_RUN=false
TEST_FLAG=true
PASSWD="blockchain"
PRIMARY=`head -1 ./serverlist.txt`
MAXNODE=`cat serverlist.txt | wc -l`

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

add_ssh_key_into_primary(){
    echo "Add your local ssh public key into primary node"
    for server_address in ${SERVER_ADDR[@]}; do
	  addkey $server_address
	done
}

add_ssh_key_form_primary_to_others(){
    echo "Primary add its ssh key into others nodes"
	scp ./sub_scripts/deploy/server_addkey.sh satoshi@$PRIMARY:/home/satoshi/
	scp innerserverlist.txt satoshi@$PRIMARY:/home/satoshi/
	ssh  satoshi@$PRIMARY "cd /home/satoshi && chmod a+x server_addkey.sh && bash server_addkey.sh"
}

distribute_the_binary(){
    echo "Send the project to primary:"
    cd $GOPATH/src/ && rm -rf hyperchain/build && rm hyperchain.tar.gz && tar -zcvf hyperchain.tar.gz ./hyperchain
    scp hyperchain.tar.gz satoshi@$PRIMARY:/home/satoshi/
    ssh satoshi@$PRIMARY "rm -rf go/src/hyperchain && tar -C /home/satoshi/go/src -xzf hyperchain.tar.gz"

    echo "Primary build the project:"
	ssh satoshi@$PRIMARY "source ~/.bash_profile && cd go/src/hyperchain && govendor build && mv hyperchain /home/satoshi"

	echo "Send the config files to primary:"
	cd $GOPATH/src/hyperchain/scripts
	scp -r ../config/ satoshi@$PRIMARY:/home/satoshi/
	scp ./sub_scripts/deploy/server_deploy.sh satoshi@$PRIMARY:/home/satoshi/

    echo "Primary send files to others:"
	ssh satoshi@$PRIMARY "chmod a+x server_deploy.sh && bash server_deploy.sh ${MAXNODE}"
}

ni=1
auto_run(){
    echo "Auto start all nodes"
    for server_address in ${SERVER_ADDR[@]}; do
	  echo $server_address
      ssh satoshi@$server_address "if [ ! -d /home/satoshi/build/ ]; then mkdir -p /home/satoshi/build/;fi"
	  gnome-terminal -x bash -c "ssh satoshi@$server_address \" cd /home/satoshi/ && cp -rf ./config/keystore ./build/ && ./hyperchain -o $ni -l 8001 -t 8081 || while true; do ifconfig && sleep 100; done\""
	  ni=`expr $ni + 1`
	done
}

if $FIRST_RUN; then
	add_ssh_key_into_primary
	add_ssh_key_form_primary_to_others
fi

for server_address in ${SERVER_ADDR[@]}; do
  echo "kill $server_address"
  ssh satoshi@$server_address "ps aux | grep hyperchain | awk '{print \$2}' | xargs kill -9"
done

distribute_the_binary
auto_run