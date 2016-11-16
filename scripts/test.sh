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


ni=1
auto_run(){
    echo "Auto start all nodes"
    for server_address in ${SERVER_ADDR[@]}; do
	  echo $server_address
      ssh hyperchain@$server_address "if [ ! -d /home/hyperchain/build/ ]; then mkdir -p /home/hyperchain/build/;fi"
	  # this line for ubuntu
	  #gnome-terminal -x bash -c "ssh hyperchain@$server_address \" cd /home/hyperchain/ && cp -rf ./config/keystore ./build/ && ./hyperchain -o $ni -l 8001 -t 8081 || while true; do sleep 1000s; done\""

	  # this line for mac
	  osascript -e 'tell app "Terminal" to do script "ssh hyperchain@'$server_address' \" cd /home/hyperchain/ && cp -rf ./config/keystore ./build/ && ./hyperchain -o '$ni' -l 8001 -t 8081 || while true; do sleep 1000s; done\""'

	  ni=`expr $ni + 1`
	done
}


for server_address in ${SERVER_ADDR[@]}; do
  echo "kill $server_address"
  ssh hyperchain@$server_address "pkill hyperchains"
  ssh hyperchain@$server_address "ps aux | grep hyperchain -o | awk '{print \$2}' | xargs kill -9"
done

auto_run