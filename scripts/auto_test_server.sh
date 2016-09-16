#!/bin/bash

set -e
# set -x
# (( $# != 1 )) && { echo >&2 "Usage: $0 \"[COMMAND]\""; exit 1; }

echo -e " _   _                        ____ _           _       "
echo -e "| | | |_   _ _ __   ___ _ __ / ___| |__   __ _(_)_ __  "
echo -e "| |_| | | | | '_ \ / _ \ '__| |   | '_ \ / _\` | | '_ \ "
echo -e "|  _  | |_| | |_) |  __/ |  | |___| | | | (_| | | | | |"
echo -e "|_| |_|\__, | .__/ \___|_|   \____|_| |_|\__,_|_|_| |_|"
echo -e "       |___/|_|                                        "
# sudo apt-get -y install expect
if [ ! -f "/usr/bin/expect" ];then
  echo "hasn't install expect,please install expect mannualy: 'apt-get install expect'"
  exit 1
fi

PASSWD="blockchain"

SERVER_ADDR=( cn6.hyperchain.cn cn7.hyperchain.cn cn8.hyperchain.cn cn9.hyperchain.cn cn12.hyperchain.cn cn13.hyperchain.cn cn14.hyperchain.cn)

# add your local pubkey into every server
add_pubkey(){
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
}


DEPLOYCOMMAND=<<EOF
    chmod a+x /home/satoshi/deploy.sh
    cd /home/satoshi/&& /home/satoshi/deploy.sh
EOF




# auto deploy the goenv
auto_deploy(){
  for server_address in ${SERVER_ADDR[@]}; do
    scp ./deploy.sh satoshi@server_address:/home/satoshi/
    ssh satoshi@server_address "chmod a+x /home/satoshi/deploy.sh;/home/satoshi/deploy.sh"
  done
}

auto_test(){
  i=1
  for server_address in ${SERVER_ADDR[@]}; do
    i=`expr $i + 1` 
    # gnome-terminal -x bash -c "ssh satoshi@$server_address \"cd /home/satoshi/gopath/src/hyperchain/scripts/ && ./server.sh $i\""
    gnome-terminal -x bash -c "ssh satoshi@$server_address \"source /home/satoshi/.profile && cd /home/satoshi/gopath/src/hyperchain/scripts/ && ./server.sh $i\""
  done
}

auto_test
