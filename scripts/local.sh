#!/usr/bin/env bash
set -e
# 设置系统环境类型
f_set_env(){
    case "$OSTYPE" in
      darwin*)
        echo "RUN SCRIPTS ON OSX"
        _SYSTYPE="MAC"
      ;;
      linux*)
        echo "RUN SCRIPTS ON LINUX"
        _SYSTYPE="LINUX"
      ;;
      *)
        echo "unknown: $OSTYPE"
        exit -1
      ;;
    esac
}

# 帮助函数
f_help(){
    echo "local.sh helper:"
    echo "  -h, --help:     show the help for this bash script"
    echo "  -k, --kill:     just kill all the processes"
    echo "  -d, --delete:   clear the old data or not; default: clear. add for not clear"
    echo "  -r, --rebuild:  rebuild the project or not; default: rebuild, add for not rebuild"
    echo "  -m, --mode:     choose the run mode; default: run many in many, add for many in one"
    echo "---------------------------------------------------"
    echo "Example for run many in one in mac without rebuild:"
    echo "./local -m -r"
}

# 检查本地运行环境
f_check_local_env(){
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
        echo -e "Please install the `jq` to parse the json file \n just type: \n sudo apt-get install jq / sudo yum -y install jq / brew install jq "
        exit 1
    fi
    # confer
    if ! type confer > /dev/null; then
        echo -e "Please install `confer` to read global.yaml config"
        exit 1
    fi
}


# 杀进程函数
f_kill_process(){
    echo "kill the bind port process"
    ps -ax | grep hyperchain | grep -v grep | awk '{print $1}' |  xargs  kill -9
}

# 删除数据函数
f_delete_data(){
for (( j=1; j<=$MAXPEERNUM; j++ ))
do
    # Clear the old data
    if [ -d "${DUMP_PATH}/node${j}/build" ];then
        rm -rf "${DUMP_PATH}/node${j}/build"
    fi

    # Creat the build dir
    if [ ! -d "${DUMP_PATH}/node${j}/build" ];then
        mkdir -p "${DUMP_PATH}/node${j}/build"
    fi
done
}

# 重新编译函数
f_rebuild(){
# Build the project
echo "Rebuild the project..."
if [ -s "${DUMP_PATH}/hyperchain" ]; then
    rm ${DUMP_PATH}/hyperchain
fi
cd ${PROJECT_PATH} && govendor build -o ${DUMP_PATH}/hyperchain
}

# 数据分发函数
f_distribute(){
# cp the config files into nodes
for (( j=1; j<=$1; j++ ))
do
    if [ ! -d "${DUMP_PATH}/node${j}" ];then
        mkdir -p ${DUMP_PATH}/node${j}
    fi
    if [ -d "${DUMP_PATH}/node${j}/config" ];then
        rm -rf ${DUMP_PATH}/node${j}/config
    fi
    cp -rf  ${CONF_PATH} ${DUMP_PATH}/node${j}/
    cp -rf  ${CONF_PATH}/peerconfigs/local_peerconfig_${j}.json ${DUMP_PATH}/node${j}/config/local_peerconfig.json
    cp -rf  ${CONF_PATH}/peerconfigs/node${j}/* ${DUMP_PATH}/node${j}/config/cert/
    cp -rf ${DUMP_PATH}/hyperchain ${DUMP_PATH}/node${j}/
done
}

# 运行命令
f_all_in_one_cmd(){
    cd $DUMP_PATH/node${1} && ./hyperchain -o ${1} -l 800${1} -t 808${1} -f 900${1} &
}

f_x_in_linux_cmd(){
    gnome-terminal -x bash -c "cd ${DUMP_PATH}/node${1} && ./hyperchain"
}

f_x_in_mac_cmd(){
    osascript -e 'tell app "Terminal" to do script "cd '$DUMP_PATH/node${1}' && ./hyperchain "'
}


# 组织运行命令函数
f_run_process(){
    for((j=1;j<=$MAXPEERNUM;j++))
    do
        case "$_SYSTYPE" in
          MAC*)
             f_x_in_mac_cmd $j
          ;;
          LINUX*)
             f_x_in_linux_cmd $j
          ;;
          ALLINONE*)
            f_all_in_one_cmd $j
          ;;
          *)
            echo "unknow OS TYPE $OSTYPE"
            exit -1
          ;;
        esac
    done
}

f_sleep(){
    sleep ${1}s
}

#####################################
#                                   #
#  MAIN INVOKE AREA  主要调用区域   #
#                                   #
#####################################


# 系统类型
_SYSTYPE="MAC"

# 项目路径
PROJECT_PATH="${GOPATH}/src/hyperchain"
# 当前脚本路径

CURRENT_PATH=`pwd`

# 输出基本路径
DUMP_PATH="${PROJECT_PATH}/build"

# 配置文件路径
CONF_PATH="${PROJECT_PATH}/config"

# 全局配置文件路径
GLOBAL_CONFIG="${CONF_PATH}/global.yaml"

# peerconfig 配置文件路径
PEER_CONFIG_FILE_NAME=`confer read ${GLOBAL_CONFIG} global.configs.peers |sed 's/"//g'`

PEER_CONFIG_FILE=${PROJECT_PATH}/${PEER_CONFIG_FILE_NAME}

# 节点数目
MAXPEERNUM=`cat ${PEER_CONFIG_FILE} | jq ".maxpeernode"`
echo "Node number is: ${MAXPEERNUM}"
# 是否删除数据（注意，默认为true）
DELETEDATA=true

# 是否重新编译
REBUILD=true

# 检查本地环境
f_check_local_env

# 设置操作系统类型
f_set_env

# 解析输出参数
while [ $# -gt 0 ]
do
    case "$1" in
    -h|--help)
        help; exit 0;;
    -k|--kill)
        f_kill_process; exit 1;;
    -d|--delete)
        DELETEDATA=false; shift;;
    -r|--rebuild)
        REBUILD=false; shift;;
    -m|--mode)
        MODE=true; shift;;
    --) shift; break;;
    -*) help; exit 1;;
    *) break;;
    esac
done

# 杀进程
f_kill_process

# 判断是否删除数据
if  $DELETEDATA ; then
    f_delete_data
fi

# 判断是否重新编译

if  $REBUILD ; then
    f_rebuild
fi

# 分发配置文件
f_distribute $MAXPEERNUM

# 运行hyperchain 进程
f_run_process


