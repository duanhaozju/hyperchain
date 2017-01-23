#!/usr/bin/env bash
set -e

# DIRS
CURRENT_DIR=`pwd`
HYPERCHAIN_DIR="/home/hyperchain"

if [ ! -f innerserverlist.txt ]; then
    echo "innerserverlist.txt is not exists!"
    exit -1
fi

MAXNODE=`cat innerserverlist.txt | wc -l`

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
 scp -r config hyperchain@$1:$HYPERCHAIN_DIR
 scp hyperchain hyperchain@$1:$HYPERCHAIN_DIR
# ssh hyperchain@$1 "rm -rf build"
}

for server_address in ${SERVER_ADDR[@]}; do
 scpfile $server_address &
done

wait