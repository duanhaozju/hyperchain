#!/usr/bin/env bash
set -e

govendor build

PRIMARY="115.159.33.105"

./local_addkey.sh

scp -r killprocess.sh satoshi@$PRIMARY:/home/satoshi/
ssh -t satoshi@$PRIMARY "chmod a+x killprocess.sh && bash killprocess.sh && rm -rf keystore"
scp -r hyperchain satoshi@$PRIMARY:/home/satoshi/
scp -r peerconfig.json satoshi@$PRIMARY:/home/satoshi/
scp -r config.yaml satoshi@$PRIMARY:/home/satoshi/
scp -r genesis.json satoshi@$PRIMARY:/home/satoshi/
scp -r keystore satoshi@$PRIMARY:/home/satoshi/
scp -r innerserverlist.txt satoshi@$PRIMARY:/home/satoshi/
scp -r server_addkey.sh satoshi@$PRIMARY:/home/satoshi/
ssh -t satoshi@$PRIMARY "chmod a+x server_addkey.sh && bash server_addkey.sh"
scp -r server_deploy.sh satoshi@$PRIMARY:/home/satoshi/
ssh -t satoshi@$PRIMARY "chmod a+x server_deploy.sh && bash server_deploy.sh"


while read line;do
 SERVER_ADDR+=" ${line}"
done < ./serverlist.txt
ni=1

for server_address in ${SERVER_ADDR[@]}; do
  echo $server_address
  gnome-terminal -x bash -c "ssh -t satoshi@$server_address \"rm -rf /tmp/hyperchain/ && cd /home/satoshi/ && ./hyperchain -o $ni -l 8081 -p ./peerconfig.json -f ./ -g ./genesis.json\""
  ni=`expr $ni + 1`
done

python ./jsonrpc/Dashboard/simpleHttpServer.py