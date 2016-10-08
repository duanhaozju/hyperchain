#!/usr/bin/env bash

set -e

if [ ! -f "/usr/bin/expect" ];then
echo "hasn't install expect,please install expect mannualy: 'apt-get install expect'"
exit 1
fi

PASSWD="blockchain"

# get the server list config
while read line;do
 SERVER_ADDR+=" ${line}"
done < ./serverlist.txt

#########################
# authorization         #
#########################

# add your local pubkey into every server
echo "┌────────────────────────┐"
echo "│    auto add ssh key    │"
echo "└────────────────────────┘"

addkey(){
#  ssh-keygen -f "/home/satoshi/.ssh/known_hosts" -R $1
  expect <<EOF
      set timeout 60
      spawn ssh-copy-id satoshi@$1
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
