#!/usr/bin/env bash

USERNAME="hyperchain"

set -e
# DIRS
CURRENT_DIR=`pwd`
HYPERCHAIN_DIR="/home/${USERNAME}"

if [ ! -f innerserverlist.txt ]; then
    echo "innerserverlist.txt is not exists!"
    exit -1
fi

MAXNODE=`cat innerserverlist.txt | wc -l`

count=0
while IFS='' read -r line || [[ -n "$line" ]]; do
    let count=$count+1
    if [ ${count} -eq 0 ]; then
        echo ${count}
        continue
    fi

   SERVER_ADDR+=" ${line}"
done < innerserverlist.txt

scpfile() {
    scp ${HYPERCHAIN_DIR}/hyperchain ${USERNAME}@$1:${HYPERCHAIN_DIR}
    scp ${HYPERCHAIN_DIR}/executor ${USERNAME}@$1:${HYPERCHAIN_DIR}
}

for server_address in ${SERVER_ADDR[@]}; do
    echo "scp hyperchain and executor binary to ${server_address}"
    scpfile ${server_address} &
done

wait