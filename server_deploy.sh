#!/usr/bin/env bash
set -e
MAXNODE=10

#kill the progress
kellprogress(){

  echo "kill the bind port process"
  for((i=1;i<=$MAXNODE;i++))
  do
      test_port=`expr 8000 + $i`
      temp_port=`lsof -i :$test_port | awk 'NR>=2{print $2}'`
      if [ x"$temp_port" != x"" ];then
          kill -9 $temp_port
      fi
  done
}

while read line;do
 SERVER_ADDR+=" ${line}"
 done < ./innerserverlist.txt

scpfile() {
 # ssh-keygen -f "/home/satoshi/.ssh/known_hosts" -R $server_address
 echo "#################"
 echo $1
 echo "#################"
 expect <<EOF
       set timeout 60
       spawn ssh -t satoshi@$1 "echo \"hello world\""
       expect {
         "yes/no" {send "yes\r";exp_continue }
         eof
       }
EOF

kellprogress
 scp -r hyperchain satoshi@$1:/home/satoshi/
 scp -r peerconfig.json satoshi@$1:/home/satoshi/
 scp -r config.yaml satoshi@$1:/home/satoshi/
 scp -r genesis.json satoshi@$1:/home/satoshi/
 ssh -t satoshi@$1 "rm -rf keystore"
 scp -r keystore satoshi@$1:/home/satoshi/
}
 for server_address in ${SERVER_ADDR[@]}; do

 scpfile $server_address &

done

wait