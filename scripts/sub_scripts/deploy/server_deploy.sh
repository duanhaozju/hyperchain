#!/usr/bin/env bash
set -e
MAXNODE=$1

count=0
while IFS='' read -r line || [[ -n "$line" ]]; do
   let count=$count+1
   if [ $count -eq 0 ]; then
    echo $count

    continue
   fi

   SERVER_ADDR+=" ${line}"
done < innerserverlist.txt

scpfile() {
 scp -r config hyperchain@$1:/home/hyperchain/
 scp hyperchain hyperchain@$1:/home/hyperchain/
# ssh hyperchain@$1 "rm -rf build"
}

for server_address in ${SERVER_ADDR[@]}; do
 scpfile $server_address
done