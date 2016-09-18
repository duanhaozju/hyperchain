#!/usr/bin/env bash
set -e

govendor build

while read line;do
  SERVER_ADDR+=" ${line}"
done < ./serverlist.txt

ni=1
for server_address in ${SERVER_ADDR[@]}; do
#    ssh -t satoshi@$server_address "cd /home/satoshi/ && ./hyperchain -o $ni -l 8081 -p ./peerconfig.json -f ./"
   gnome-terminal -x bash -c "scp -r ./hyperchain satoshi@$server_address:/home/satoshi/ && \
                              scp -r ./peerconfig.json satoshi@$server_address:/home/satoshi/ && \
                              scp -r ./config.yaml satoshi@$server_address:/home/satoshi/ && \
                              scp -r ./genesis.json satoshi@$server_address:/home/satoshi/ && \
                              ssh -t satoshi@$server_address \"rm -rf /tmp/hyperchain/ && cd /home/satoshi/ && ./hyperchain -o $ni -l 8081 -p ./peerconfig.json -f ./ -g ./genesis.json\""
   ni=`expr $ni + 1`
done