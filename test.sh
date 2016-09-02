#!/bin/bash
echo "##########################"
echo "#   TEST FOR HYPERCHAIN  #"
echo "##########################"

echo "kill the bind port process"

kill -9 $(lsof -i :8001 | awk 'NR>=2{print $2}')
kill -9 $(lsof -i :8002 | awk 'NR>=2{print $2}')
kill -9 $(lsof -i :8003 | awk 'NR>=2{print $2}')
kill -9 $(lsof -i :8004 | awk 'NR>=2{print $2}')

#rebuild the application
echo "rebuild the application"

govendor build

echo "run the application"

./hyperchain -o 1 -l 8081 > ./node_1_log.txt &
./hyperchain -o 2 -l 8082 > ./node_2_log.txt &
./hyperchain -o 3 -l 8083 > ./node_3_log.txt &
./hyperchain -o 4 -l 8084 > ./node_4_log.txt &

echo "All process are running background"