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
    echo -e "Please install the `govendor`, just type:\ngo get -u github.com/kardianos/govendor"
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
echo -e "copy config dir  into build dir.."
cp -rf "${CONF_PATH}" "${DUMP_PATH}/"


# 读取配置文件
PEER_CONFIG="${CONF_PATH}/peerconfig.json"
MAXPEERNUM=`cat ${PEER_CONFIG} | jq ".maxpeernode"`
echo "Node number is: ${MAXPEERNUM}"


# 杀死所有进程
#kill the progress
echo "kill the bind port process"
for((i=1;i<=$MAXPEERNUM;i++))
do
    temp_port=`lsof -i :800$i | awk 'NR>=2{print $2}'`
    if [ x"$temp_port" != x"" ];then
        kill -9 $temp_port
    fi
done
# 编译项目
echo "build the project"
govendor build -o ${DUMP_PATH}/hyperchain

#清空数据
echo "clear the old data"
rm -rf "${DUMP_PATH}/build"

#执行测试
for((j=1;j<=$MAXPEERNUM;j++))
do
	gnome-terminal -x bash -c "cd ${DUMP_PATH} && ./hyperchain -o ${j} -l 800${j} -t 808${j}"
done

#count=0
#while read line;do
#    if [ $count -ne 0 ]; then
#        # SERVER_ADDR+=" ${line}"
#        echo $line
#
#

#    fi
#    let count=count+1
#done < ./serverlist.txt


#for (( i=0; i<$MAXPEERNUM;i++))
#do
#	gnome-terminal -x bash -c "cd ${DUMP_PATH} && ./hyperchain -o ${i} -l 800${i} -t 808${i}"
#done


