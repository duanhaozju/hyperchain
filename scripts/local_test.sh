#!/usr/bin/env bash
# Author: Chen Quan
# Date  : 2016-10-12
# Hint  :this shell script is used for local test and it will clean the tmp dir and rebuild the project
#        please ensure `govendor` installed, this script is suitable for MacOS and Ubuntu 16.04

# exit if error occurred
set -e
#set -x

# test the env
# 检查环境
if ! type go > /dev/null; then
    echo -e "Please install the go env correctly!"
    exit 1
fi

if ! type govendor > /dev/null; then
    # install foobar here
    echo -e "Please install the govendor, just type:\ngo get -u github.com/kardianos/govendor"
    exit 1
fi

if ! type jq > /dev/null; then
    echo -e "Please install the `jq` to parse the json file \n just type: \n sudo apt-get install jq"
    exit 1
fi

cd ../
# build the project
PROJECT_PATH=`pwd`
DUMP_PATH="${PROJECT_PATH}/build"
CONF_PATH="${PROJECT_PATH}/config"

if [ ! -d "${DUMP_PATH}" ];then
    echo "auto creat the build dir"
    mkdir -p ${DUMP_PATH}
fi

# 检查所有的配置文件

echo -e "copy the peerconfig form config dir files into build dir.."
cp -rf "${CONF_PATH}/peerconfig.json" "${DUMP_PATH}/"


echo -e "copy the config form config dir files into build dir.."
cp -rf "${CONF_PATH}/config.yaml" "${DUMP_PATH}/"

echo -e "copy the membersrvc form config dir files into build dir.."
cp -rf "${CONF_PATH}/membersrvc.yaml" "${DUMP_PATH}/"

echo -e "copy the cert files"

cp -rf "${CONF_PATH}/cert" "${DUMP_PATH}/"



# 读取配置文件
PEER_CONFIG="${DUMP_PATH}/peerconfig.json"
MAXPEERNUM=`cat ${PEER_CONFIG} | jq ".maxpeernode"`
echo $MAXPEERNUM
#    sed -e 's/[{}]/''/g' |
#    awk -v k="text" '{print $1}'
#    awk -v k="text" '{n=split($0,a,","); for (i=1; i<=1; i++) print a[i]}'


count=0
while read line;do
    if [ $count -ne 0 ]; then
        # SERVER_ADDR+=" ${line}"
        echo $line


    fi
    let count=count+1
done < ./serverlist.txt


# 杀死所有进程
#kill the progress
echo "kill the bind port process"
for((i=0;i<=$MAXPEERNUM;i++))
do
    temp_port=`lsof -i :800$i | awk 'NR>=2{print $2}'`
    if [ x"$temp_port" != x"" ];then
        kill -9 $temp_port
    fi
done

# 编译项目
echo "build the project"
govendor build -o ${DUMP_PATH}/hyperchain

#执行测试


for (( i=0; i<$MAXPEERNUM;i++))
do
	gnome-terminal -x bash -c "cd ${DUMP_PATH} && ./hyperchain -o ${i} -l 808${i} -p ./peerconfig.json -f ./"
done
