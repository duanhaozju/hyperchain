#!/usr/bin/env bash
set -e

govendor build

FIRSTSERVER="114.55.64.132"
MAXNODE=16

 scp -r killprocess.sh satoshi@$FIRSTSERVER:/home/satoshi/
 ssh -t satoshi@$FIRSTSERVER "chmod a+x killprocess.sh && bash killprocess.sh"
 scp -r hyperchain satoshi@$FIRSTSERVER:/home/satoshi/
 scp -r peerconfig.json satoshi@$FIRSTSERVER:/home/satoshi/
 scp -r config.yaml satoshi@$FIRSTSERVER:/home/satoshi/
 scp -r genesis.json satoshi@$FIRSTSERVER:/home/satoshi/
 scp -r server_deploy.sh satoshi@$FIRSTSERVER:/home/satoshi/
 scp -r innerserverlist.txt satoshi@$FIRSTSERVER:/home/satoshi/
 ssh -t satoshi@$FIRSTSERVER "chmod a+x server_deploy.sh && bash server_deploy.sh"


while read line;do
 SERVER_ADDR+=" ${line}"
done < ./serverlist.txt
ni=1
for server_address in ${SERVER_ADDR[@]}; do
echo $server_address
 #    ssh -t satoshi@$server_address "cd /home/satoshi/ && ./hyperchain -o $ni -l 8081 -p ./peerconfig.json -f ./"

#osascript -e 'tell app "Terminal" to do script "ssh -t satoshi@'$server_address' \"rm -rf /tmp/hyperchain/ && cd /home/satoshi/ &&chmod a+x hyperchain && ./hyperchain -o $ni -l 8081 -p ./peerconfig.json -f ./ -g ./genesis.json\""'

 gnome-terminal -x bash -c "ssh -t satoshi@$server_address \"rm -rf /tmp/hyperchain/ && cd /home/satoshi/ && ./hyperchain -o $ni -l 8081 -p ./peerconfig.json -f ./ -g ./genesis.json\""
 ni=`expr $ni + 1`
done