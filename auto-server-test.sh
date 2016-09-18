#!/usr/bin/env bash
while read line;do
 SERVER_ADDR+=" ${line}"
done < ./serverlist.txt

ni=1
for server_address in ${SERVER_ADDR[@]}; do
  ni=`expr $ni + 1`
  gnome-terminal -x bash -c "ssh -t satoshi@$server_address \"cd /home/satoshi/gopath/src/hyperchain/scripts && git checkout -- server.sh && git pull origin develop && pwd && chmod a+x server.sh && ./server.sh $ni\""
#   ssh -t satoshi@$server_address "cd /home/satoshi/gopath/src/hyperchain/scripts && git checkout -- server.sh && git pull origin develop && pwd && chmod a+x server.sh && ./server.sh $ni"
done