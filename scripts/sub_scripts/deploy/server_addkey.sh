#!/usr/bin/env bash

set -e

if [ ! -f "/usr/bin/expect" ];then
echo "hasn't install expect,please install expect mannualy: 'apt-get install expect'"
exit 1
fi

PASSWD="hyperchain"

while IFS='' read -r line || [[ -n "$line" ]]; do
   SERVER_ADDR+=" ${line}"
done < innerserverlist.txt

#########################
# authorization         #
#########################

addkey(){
#  ssh-keygen -f "/home/hyperchain/.ssh/known_hosts" -R $1
  expect <<EOF
      set timeout 60
      spawn ssh-copy-id hyperchain@$1
      expect {
        "yes/no" {send "yes\r";exp_continue }
        "s password:" {send "$PASSWD\r";exp_continue }
        eof
      }
EOF
}

for server_address in ${SERVER_ADDR[@]}; do
  addkey $server_address &
done

wait