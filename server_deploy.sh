#!/usr/bin/env bash
set -e
MAXNODE=4

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
 ni=1
 for server_address in ${SERVER_ADDR[@]}; do
 expect <<EOF
       set timeout 60
       spawn ssh -t satoshi@$server_address "echo \"hello world\""
       expect {
         "yes/no" {send "yes\r";exp_continue }
         eof
       }
EOF

kellprogress

 scp -r hyperchain satoshi@$server_address:/home/satoshi/
 scp -r peerconfig.json satoshi@$server_address:/home/satoshi/
 scp -r config.yaml satoshi@$server_address:/home/satoshi/
 scp -r genesis.json satoshi@$server_address:/home/satoshi/
 ssh -t satoshi@$server_address "rm -rf keystore"
 scp -r keystore satoshi@$server_address:/home/satoshi/
 ni=`expr $ni + 1`
done