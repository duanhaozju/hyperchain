#!/usr/bin/env bash

FIRST_RUN=false
TEST_FLAG=true
PASSWD="hyperchain"
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
      spawn ssh-copy-id hyperchain@$1
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
	scp ./sub_scripts/deploy/server_addkey.sh hyperchain@$PRIMARY:/home/hyperchain/
	scp innerserverlist.txt hyperchain@$PRIMARY:/home/hyperchain/
	ssh  hyperchain@$PRIMARY "cd /home/hyperchain && chmod a+x server_addkey.sh && bash server_addkey.sh"
}

distribute_the_binary(){
    echo "Send the project to primary:"
    cd $GOPATH/src/
    rm -rf hyperchain/build
    rm hyperchain.tar.gz
    tar -zcvf hyperchain.tar.gz ./hyperchain
    scp hyperchain.tar.gz hyperchain@$PRIMARY:/home/hyperchain/
    ssh hyperchain@$PRIMARY "rm -rf go/src/hyperchain"
    ssh hyperchain@$PRIMARY "tar -C /home/hyperchain/go/src -xzf hyperchain.tar.gz"
    echo "Primary build the project:"
	ssh hyperchain@$PRIMARY "source ~/.bash_profile && cd go/src/hyperchain && govendor build && mv hyperchain /home/hyperchain"

	echo "Send the config files to primary:"
	cd $GOPATH/src/hyperchain/scripts
	scp -r ../config/ hyperchain@$PRIMARY:/home/hyperchain/
	scp ./sub_scripts/deploy/server_deploy.sh hyperchain@$PRIMARY:/home/hyperchain/

    echo "Primary send files to others:"
	ssh hyperchain@$PRIMARY "chmod a+x server_deploy.sh && bash server_deploy.sh ${MAXNODE}"
}

ni=1
auto_run(){
    echo "Auto start all nodes"
    for server_address in ${SERVER_ADDR[@]}; do
	  echo $server_address
      ssh hyperchain@$server_address "if [ ! -d /home/hyperchain/build/ ]; then mkdir -p /home/hyperchain/build/;fi"
	  gnome-terminal -x bash -c "ssh hyperchain@$server_address \" cd /home/hyperchain/ && cp -rf ./config/keystore ./build/ && ./hyperchain -o $ni -l 8001 -t 8081 || while true; do sleep 1000s; done\""
	  ni=`expr $ni + 1`
	done
}

if $FIRST_RUN; then
	add_ssh_key_into_primary
	add_ssh_key_form_primary_to_others
fi

for server_address in ${SERVER_ADDR[@]}; do
  echo "kill $server_address"
  ssh hyperchain@$server_address "ps aux | grep hyperchain -o | awk '{print \$2}' | xargs kill -9"
done

distribute_the_binary
auto_run