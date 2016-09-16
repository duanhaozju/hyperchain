#!/usr/bin/env bash
# Author: ChenQuan
# Description: this script is used for deploy code env in the server side.
# Date: 2016-09-15
# Happy mid-moon day!

set -e
echo -e " _   _                        ____ _           _       "
echo -e "| | | |_   _ _ __   ___ _ __ / ___| |__   __ _(_)_ __  "
echo -e "| |_| | | | | '_ \ / _ \ '__| |   | '_ \ / _\` | | '_ \ "
echo -e "|  _  | |_| | |_) |  __/ |  | |___| | | | (_| | | | | |"
echo -e "|_| |_|\__, | .__/ \___|_|   \____|_| |_|\__,_|_|_| |_|"
echo -e "       |___/|_|                                        "

if [ ! -f "/usr/bin/expect" ];then
  echo "hasn't install expect,please install expect mannualy: 'apt-get install expect'"
  exit 1
fi

PASSWD="blockchain"

# get the server list config
while read line;do
   SERVER_ADDR+=" ${line}"
done < ../serverlist.txt

#########################
# authorization         #
#########################

# add your local pubkey into every server
    echo "┌────────────────────────┐"
    echo "│    auto add ssh key    │"
    echo "└────────────────────────┘"
    for server_address in ${SERVER_ADDR[@]}; do
    expect <<EOF
        set timeout 60
        spawn ssh-copy-id satoshi@$server_address
        expect {
          "s password:" {send "$PASSWD\r";exp_continue }
          eof
        }
EOF
    done

#########################
# deploy                #
#########################

echo "┌────────────────────────┐"
echo "│    auto deploy         │"
echo "└────────────────────────┘"
for server_address in ${SERVER_ADDR[@]}; do
    scp ./goenv_install.sh satoshi@server_address:/home/satoshi/
    nome-terminal -x bash -c "ssh satoshi@server_address \"chmod a+x /home/satoshi/goenv_install.sh;/home/satoshi/goenv_install.sh\""
done