#!/usr/bin/env bash
set -e
MAXNODE=$1

#kill the progress
killprogress(){

  echo "kill the bind port process"
  for((i=0;i<=$MAXNODE;i++))
  do
      test_port=`expr 8000 + $i`
      temp_port=`lsof -i :$test_port | awk 'NR>=2{print $2}'`
      if [ x"$temp_port" != x"" ];then
          kill -9 $temp_port
      fi
  done
}

count=0
while IFS='' read -r line || [[ -n "$line" ]]; do
   let count=$count+1
   if [ $count -eq 0 ]; then
    continue
   fi

   SERVER_ADDR+=" ${line}"
done < innerserverlist.txt

scpfile() {
 # ssh-keygen -f "/home/satoshi/.ssh/known_hosts" -R $server_address
 expect <<EOF
       set timeout 60
       spawn ssh -t satoshi@$1"echo \"hello world\""
       expect {
         "yes/no" {send "yes\r";exp_continue }
         eof
       }
EOF

 killprogress $MAXNODE
 scp -r config satoshi@$1:/home/satoshi/
 scp -r hyperchain satoshi@$1:/home/satoshi/
 ssh -t satoshi@$1 "rm -rf build"
}

for server_address in ${SERVER_ADDR[@]}; do
 scpfile $server_address &
done

wait