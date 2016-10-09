#!/bin/bash
echo -e " _   _                        ____ _           _       "
echo -e "| | | |_   _ _ __   ___ _ __ / ___| |__   __ _(_)_ __  "
echo -e "| |_| | | | | '_ \ / _ \ '__| |   | '_ \ / _\` | | '_ \ "
echo -e "|  _  | |_| | |_) |  __/ |  | |___| | | | (_| | | | | |"
echo -e "|_| |_|\__, | .__/ \___|_|   \____|_| |_|\__,_|_|_| |_|"
echo -e "       |___/|_|                                        "


#kill the progress

echo "kill the bind port process"
for((i=0;i<=4;i++))
do
    temp_port=`lsof -i :800$i | awk 'NR>=2{print $2}'`
    if [ x"$temp_port" != x"" ];then
        kill -9 $temp_port
    fi
done


#rebuild the application
echo "rebuild the application"
govendor build


echo "remove tmp data"
rm -rf /tmp/hyperchain/

echo "running the application"
    for((j=1;j<=4;j++))
    do
        ./hyperchain -o $j -l 808$j -p ./p2p/peerconfig.json -f ./consensus/pbft/ -g ./core/genesis.json > output$j &
    done

python ./jsonrpc/Dashboard/simpleHttpServer.py

echo "All process are running background"

