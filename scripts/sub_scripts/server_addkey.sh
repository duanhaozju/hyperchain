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

addkeyForCentOS(){
  expect <<EOF
      set timeout 60
      spawn ssh-copy-id hyperchain@$1
      expect {
        "yes/no" {send "yes\r";exp_continue }
        "password:" {send "$PASSWD\r";exp_continue }
        eof
      }
EOF
}

addkeyForSuse(){
  expect <<EOF
      set timeout 60
      spawn ssh-copy-id hyperchain@$1
      expect {
        "yes/no" {send "yes\r";exp_continue }
        "Password:" {send "$PASSWD\r";exp_continue }
        eof
      }
EOF
}

echo $1

if $1; then
    for server_address in ${SERVER_ADDR[@]}; do
      addkeyForCentOS $server_address &
    done
else
    for server_address in ${SERVER_ADDR[@]}; do
      addkeyForSuse $server_address &
    done
fi
wait