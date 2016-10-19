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

if ! type tmux > /dev/null; then
    echo -e "Please install the `tmux` to parse the json file \n just type: \n sudo apt-get install tmux"
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

#清空数据
echo "clear the old data"
rm -rf "${DUMP_PATH}/build"

#执行测试
tmux_init()
{
	tmux new-session -s "hyperchain" -d -n "local"    # 开启一个会话
	tmux new-window -n "other"
	tmux split-window -v
	tmux select-pane -t 1
	tmux split-window -h
	tmux select-pane -t 3
	tmux split-window -h
	sleep 2
	count=0
	for _pane in $(tmux list-panes  -t hyperchain -F '#P'); do
		let count=count+1
        tmux send-keys -t ${_pane} "cd ${DUMP_PATH} && ./hyperchain -o ${count} -l 800${count} -t 808${count}" Enter
	done
    tmux -2 attach-session -d           # tmux -2强制启用256color，连接已开启的tmux
}

# 判断是否已有开启的tmux会话，没有则开启
if which tmux 2>&1 >/dev/null; then
	if [ "hyperchain:" ==  `tmux ls | awk '{print$1}'` ];then
		echo "exist tmux session"
		tmux kill-session -t hyperchain
	fi
	echo "going to open 'tmux' to running auto local test,if you want to exit, just type 'Ctrl-b d'"
	echo "将要开启'tmux' 运行自动化本地测试，如果你想退出，请按'Ctrl-b d'"
	echo "如果你不想使用本脚本重新启动测试，请手动运行命令 ' tmux kill-session -t hyperchain '"
	sleep 2
    test -z "$TMUX" && (tmux attach || tmux_init)
fi


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


