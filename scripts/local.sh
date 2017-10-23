#!/usr/bin/env bash
#set -e
# set environment
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

# help prompt message
f_help(){
    echo "local.sh helper:"
    echo "  -h, --help:     show the help for this bash script"
    echo "  -k, --kill:     just kill all the processes"
    echo "  -d, --delete:   clear the old data or not; default: clear. add for not clear"
    echo "  -r, --rebuild:  rebuild the project or not; default: rebuild, add for not rebuild"
    echo "  -c, --hypercli: rebuild hypercli or not; default: not rebuild, add for rebuild"
    echo "  -m, --mode:     choose the run mode; default: run many in many, add for many in one"
    echo "---------------------------------------------------"
    echo "Example for run many in one in mac without rebuild:"
    echo "./local -m -r"
}

# check local run environment
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
        echo -e "Please install `confer` to read global.toml config"
        exit 1
    fi
}

# kill hyperchain process
f_kill_process(){
	pkill hyperchain
}

# clear data
f_delete_data(){
echo "Delete old node data"
for (( j=1; j<=$MAXPEERNUM; j++ ))
do
    # Clear the old data
    if [ -d "${DUMP_PATH}/node${j}" ];then
        rm -rf "${DUMP_PATH}/node${j}"
    fi
done
}

# rebuild the function
f_rebuild(){
# Build the project
echo "Rebuild the project..."
if [ -s "${DUMP_PATH}/hyperchain" ]; then
    rm ${DUMP_PATH}/hyperchain
fi
cd ${PROJECT_PATH} && govendor build -ldflags -s -o ${DUMP_PATH}/hyperchain -tags=embed
}

# rebuild hypercli
f_rebuild_hypercli(){
echo "Rebuild hypercli ..."
cd ${CLI_PATH} && govendor build
}

# build executor
f_rebuild_executor(){
    echo "Build executor ..."
    cd ${EXECUTOR_PATH} && govendor build
    for (( j=1; j<=$MAXPEERNUM; j++ ))
    do
#        if [ ! -d "${DUMP_PATH}/node${j}/executor" ]; then
#            mkdir ${DUMP_PATH}/node${j}/executor
#        fi
        cp ${EXECUTOR_PATH}/executor ${DUMP_PATH}/node${j}
#        cp ${DUMP_PATH}/node${j}/global.toml ${DUMP_PATH}/node${j}/executor
#        cp -rf ${DUMP_PATH}/node${j}/namespaces ${DUMP_PATH}/node${j}/executor
    done

}


# distribute node package
f_distribute(){
# cp the config files into nodes
for (( j=1; j<=$1; j++ ))
do
    # distribute the project
    if [ ! -d "${DUMP_PATH}/node${j}" ];then
        mkdir -p ${DUMP_PATH}/node${j}
    fi
    if [ -d "${DUMP_PATH}/node${j}/namespaces" ];then
        rm -rf ${DUMP_PATH}/node${j}/namespaces
    fi

    cp -rf  ${CONF_PATH}/* ${DUMP_PATH}/node${j}/
    rm -rf ${DUMP_PATH}/node${j}/peerconfigs
    rm -rf ${DUMP_PATH}/node${j}/tls
    #peerconfig.toml
    cp -rf  ${CONF_PATH}/peerconfigs/peerconfig_${j}.toml ${DUMP_PATH}/node${j}/namespaces/global/config/peerconfig.toml
    cp -rf  ${CONF_PATH}/peerconfigs/peerconfig_${j}.toml ${DUMP_PATH}/node${j}/namespaces/ns_2e6160583867/config/peerconfig.toml
    #namespace's global

    cp -rf  ${CONF_PATH}/global.toml ${DUMP_PATH}/node${j}/global.toml
    if [ ${_SYSTYPE} = "MAC" ]; then
        sed -i "" "s/8081/808${j}/g" ${DUMP_PATH}/node${j}/global.toml
        sed -i "" "s/9001/900${j}/g" ${DUMP_PATH}/node${j}/global.toml
        sed -i "" "s/50081/5008${j}/g" ${DUMP_PATH}/node${j}/global.toml
        sed -i "" "s/50051/5005${j}/g" ${DUMP_PATH}/node${j}/global.toml
        sed -i "" "s/50011/5001${j}/g" ${DUMP_PATH}/node${j}/global.toml
        sed -i "" "s/10001/1000${j}/g" ${DUMP_PATH}/node${j}/global.toml
        sed -i "" "s/50061/5006${j}/g" ${DUMP_PATH}/node${j}/global.toml

    else
        sed -i "s/8081/808${j}/g" ${DUMP_PATH}/node${j}/global.toml
        sed -i "s/9001/900${j}/g" ${DUMP_PATH}/node${j}/global.toml
        sed -i "s/50081/5008${j}/g" ${DUMP_PATH}/node${j}/global.toml
        sed -i "s/50051/5005${j}/g" ${DUMP_PATH}/node${j}/global.toml
        sed -i "s/50011/5001${j}/g" ${DUMP_PATH}/node${j}/global.toml
        sed -i "s/10001/1000${j}/g" ${DUMP_PATH}/node${j}/global.toml
        sed -i "s/50061/5006${j}/g" ${DUMP_PATH}/node${j}/global.toml
    fi

    cp -rf  ${CONF_PATH}/peerconfigs/addr_${j}.toml ${DUMP_PATH}/node${j}/addr.toml
    cp -rf  ${DUMP_PATH}/hyperchain ${DUMP_PATH}/node${j}/
    #tls configuration
    cp -rf  ${CONF_PATH}/peerconfigs/cert${j}/* ${DUMP_PATH}/node${j}/namespaces/global/config/certs/
    cp -rf  ${CONF_PATH}/peerconfigs/cert${j}/* ${DUMP_PATH}/node${j}/namespaces/ns_2e6160583867/config/certs/

    #certs
    mkdir ${DUMP_PATH}/node${j}/tls
    cp -rf  ${CONF_PATH}/tls/tls_peer${j}.cert ${DUMP_PATH}/node${j}/tls/tls_peer.cert
    cp -rf  ${CONF_PATH}/tls/tls_peer${j}.priv ${DUMP_PATH}/node${j}/tls/tls_peer.priv
    cp -rf  ${CONF_PATH}/tls/tlsca.ca ${DUMP_PATH}/node${j}/tls/tlsca.ca

    # distribute hypercli
    if [ ! -d "${DUMP_PATH}/node${j}/hypercli" ];then
        mkdir ${DUMP_PATH}/node${j}/hypercli
    fi
    if [ ! -e "${CLI_PATH}/hypercli" ]; then
        f_rebuild_hypercli
    fi
    cp -rf  ${CLI_PATH}/hypercli ${DUMP_PATH}/node${j}/hypercli
    cp -rf  ${CLI_PATH}/keyconfigs ${DUMP_PATH}/node${j}/hypercli

    BIN_PATH=${DUMP_PATH}/node${j}/bin

    # distribute bin
    if [ -d ${BIN_PATH} ];then
        rm -rf ${BIN_PATH}
    fi
    mkdir -p ${BIN_PATH}
    cp ${PROJECT_PATH}/scripts/sub_scripts/start.sh ${BIN_PATH}
    cp ${PROJECT_PATH}/scripts/sub_scripts/stop.sh ${BIN_PATH}
    cp ${PROJECT_PATH}/scripts/sub_scripts/stop.sh ${BIN_PATH}/stop_local.sh
    if [ ${_SYSTYPE} = "MAC" ]; then
        sed -i "" "s/8081/808${j}/g" ${BIN_PATH}/stop_local.sh
    else
        sed -i "s/8081/808${j}/g" ${BIN_PATH}/stop_local.sh
    fi
done
}

f_all_in_one_cmd(){
    cd $DUMP_PATH/node${1} && ./hyperchain -o ${1} -l 800${1} -t 808${1} -f 900${1} &
}

f_x_in_linux_cmd(){
    gnome-terminal -x bash -c "cd $DUMP_PATH/node${1}/bin && ./start.sh"
}

f_x_in_mac_cmd(){
    osascript -e 'tell app "Terminal" to do script "cd '$DUMP_PATH/node${1}/bin' && ./start.sh"'
#    cd ${DUMP_PATH}/node${1}/executor && ./executor &
    osascript -e 'tell app "Terminal" to do script "cd '$DUMP_PATH/node${1}' && ./executor"'
}

# run process by os type
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

start_hyperjvm() {
    cd ${PROJECT_PATH}/core/vm/jcee/java && ./build.sh
    for j in  1 2 3 4
    do
        cp -rf ${PROJECT_PATH}/core/vm/jcee/java/hyperjvm ${DUMP_PATH}/node$j/
    done
    cd ${DUMP_PATH}/node1/hyperjvm/bin/ && ./stop_hyperjvm.sh
}

f_sleep(){
    sleep ${1}s
}

#####################################
#                                   #
#  MAIN INVOKE AREA                 #
#                                   #
#####################################


# system type
_SYSTYPE="MAC"

PROJECT_PATH="${GOPATH}/src/hyperchain"

# work path
CURRENT_PATH=`pwd`

# output root dir
DUMP_PATH="${PROJECT_PATH}/build"

# config file path
CONF_PATH="${PROJECT_PATH}/configuration"

# global config path
GLOBAL_CONFIG="${CONF_PATH}/namespaces/global/config/namespace.toml"

# hypercli root path
CLI_PATH="${PROJECT_PATH}/hypercli"

# executor root path
EXECUTOR_PATH="${PROJECT_PATH}/service/executor"

# peerconfig
PEER_CONFIG_FILE_NAME=`confer read ${GLOBAL_CONFIG} global.configs.peers |sed 's/"//g'`
PEER_CONFIG_FILE_NAME="configuration/"$PEER_CONFIG_FILE_NAME
PEER_CONFIG_FILE=${PROJECT_PATH}/${PEER_CONFIG_FILE_NAME}

# node num
# MAXPEERNUM=`cat ${PEER_CONFIG_FILE} | jq ".maxpeernode"`
# echo "Node number is: ${MAXPEERNUM}"
MAXPEERNUM=4

# delete data? default = true
DELETEDATA=true

# rebuild the project or not? default = true
REBUILD=true

# rebuild hypercli or not? default = false
HYPERCLI=false

# distribute jvm or not? default = false
HYPERJVM=true

# run process or not? default = true
RUN=true

# 1.check local env
f_check_local_env

# 2.set system type
f_set_env

# exe by extra input params
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
    -c|--hypercli)
        HYPERCLI=true; shift;;
    -j|--jvm)
        HYPERJVM=true; shift;;
    -m|--mode)
        MODE=true; shift;;
    -n|--run)
        RUN=false; shift;;
    --) shift; break;;
    -*) help; exit 1;;
    *) break;;
    esac
done

# kill existing process
f_kill_process

# handle data delete
if  $DELETEDATA ; then
    f_delete_data
fi

# handle rebuild issues

if  $REBUILD ; then
    f_rebuild
fi

if [[ $? != 0 ]]; then
echo "compile failed, script stopped."
exit 1
fi
if $HYPERCLI ; then
    f_rebuild_hypercli
fi

# distribute files
f_distribute $MAXPEERNUM

f_rebuild_executor


# run hyperchain node
if ${HYPERJVM}; then
    start_hyperjvm
fi

if ${RUN}; then
    f_run_process
fi
