#!/usr/bin/env bash
MAXNODE=1
killprocess(){

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
killprocess