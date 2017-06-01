#!/usr/bin/env bash
set -e

# DIRS
CURRENT_DIR=`pwd`
HYPERCHAIN_DIR="/home/hyperchain"
HPC_PRI_HYPERCHAIN_DIR="/home/hyperchain/go/src/hyperchain"

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
 scp -r ${HPC_PRI_HYPERCHAIN_DIR}/build/node${ni} hyperchain@$1:$HYPERCHAIN_DIR/node${ni}
 ni=`expr $ni + 1`
# ssh hyperchain@$1 "rm -rf build"
}

ni=1
for server_address in ${SERVER_ADDR[@]}; do
#     scp -r config hyperchain@$server_address:$HYPERCHAIN_DIR
     scp -r ${HPC_PRI_HYPERCHAIN_DIR}/build/node${ni} hyperchain@$server_address:$HYPERCHAIN_DIR/
     scp hyperchain hyperchain@$server_address:$HYPERCHAIN_DIR/node${ni}
     ni=`expr $ni + 1`
done

echo "generate config..."
echo "${HPC_PRI_HYPERCHAIN_DIR}/scripts/namespace/gen_config.sh"
${HPC_PRI_HYPERCHAIN_DIR}/scripts/namespace/gen_config.sh global

ni=1
for server_address in ${SERVER_ADDR[@]}; do
     scp -r ${HPC_PRI_HYPERCHAIN_DIR}/build/node${ni}/namespaces/global/config/local_peerconfig.json hyperchain@$server_address:$HYPERCHAIN_DIR/node${ni}/namespaces/global/config
     ni=`expr $ni + 1`
done

wait