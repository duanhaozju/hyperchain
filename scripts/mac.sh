#!/usr/bin/env bash
# Stop on first error
#set -e
#set -x
echo "kill process"
ps aux | grep hyperchain | awk '{print $2}' | xargs kill -9
ports1=`lsof -i :8000 | awk 'NR>=2{print $2}'`
if [ x"$ports1" != x"" ];then
    kill -9 $ports1
fi
#rebuild the application
cd ..
# clean the build folder
rm -rf ./build
mkdir -p build
echo "rebuild the application"
govendor build -o ./build/hyperchain
cp -rf ./config ./build/
mkdir -p ./build/build
cp -rf ./config/keystore ./build/build
cd -

echo "run the application"
osascript -e 'tell app "Terminal" to do script "cd $GOPATH/src/hyperchain/build && ./hyperchain -o 1 -l 8001 -t 8081"'
osascript -e 'tell app "Terminal" to do script "cd $GOPATH/src/hyperchain/build && ./hyperchain -o 2 -l 8002 -t 8082"'
osascript -e 'tell app "Terminal" to do script "cd $GOPATH/src/hyperchain/build && ./hyperchain -o 3 -l 8003 -t 8083"'
osascript -e 'tell app "Terminal" to do script "cd $GOPATH/src/hyperchain/build && ./hyperchain -o 4 -l 8004 -t 8084"'


python ../jsonrpc/Dashboard/simpleHttpServer.py

echo "All process are running background"
