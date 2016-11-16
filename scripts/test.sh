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

distribute_the_binary(){
    echo "Send the project to primary:"
    cd $GOPATH/src/
    rm -rf hyperchain/build
    rm hyperchain.tar.gz
    tar -zcf hyperchain.tar.gz ./hyperchain
    scp hyperchain.tar.gz hyperchain@$PRIMARY:/home/hyperchain/
    ssh hyperchain@$PRIMARY "rm -rf go/src/hyperchain"
    ssh hyperchain@$PRIMARY "tar -C /home/hyperchain/go/src -xzf hyperchain.tar.gz"
    echo "Primary build the project:"
	ssh hyperchain@$PRIMARY "source ~/.bashrc && cd go/src/hyperchain && govendor build && mv hyperchain /home/hyperchain"

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
	  # this line for ubuntu
	  gnome-terminal -x bash -c "ssh hyperchain@$server_address \" cd /home/hyperchain/ && cp -rf ./config/keystore ./build/ && ./hyperchain -o $ni -l 8001 -t 8081 \""

	  # this line for mac
#	  osascript -e 'tell app "Terminal" to do script "ssh hyperchain@'$server_address' \" cd /home/hyperchain/ && cp -rf ./config/keystore ./build/ && ./hyperchain -o '$ni' -l 8001 -t 8081 \""'

	  ni=`expr $ni + 1`
	done
}


for server_address in ${SERVER_ADDR[@]}; do
  echo "kill $server_address"
  ssh hyperchain@$server_address "rm -rf build"
  ssh hyperchain@$server_address "pkill hyperchains"
  ssh hyperchain@$server_address "ps aux | grep 'hyperchain -o' | awk '{print \$2}' | xargs kill -9"
done

distribute_the_binary
auto_run