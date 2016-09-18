#!/usr/bin/env bash
set -e

govendor build

while read line;do
  SERVER_ADDR+=" ${line}"
done < ./serverlist.txt

ni=1
for server_address in ${SERVER_ADDR[@]}; do
   scp -r ./hyperchain satoshi@$server_address:/home/satoshi/
   scp -r ./peerconfig satoshi@$server_address:/home/satoshi/
   scp -r ./config.yml satoshi@$server_address:/home/satoshi/
   gnome-terminal -x bash -c "ssh -r -t satoshi@$server_address \"cd /home/satoshi/ && ./hyperchain -o $ni -l 8081 -p ./peerconfig.json -f ./config.yml\""
   ni=`expr $ni + 1`
done