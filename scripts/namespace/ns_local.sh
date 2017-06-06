#!/usr/bin/env bash
set -e
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
        echo -e "Please install the 'govendor', just type:\ngo get -u github.com/kardianos/govendor"
        exit 1
    fi

    if ! type jq > /dev/null; then
        echo -e "Please install the 'jq' to parse the json file \n just type: \n sudo apt-get install jq / sudo yum -y install jq / brew install jq "
        exit 1
    fi
    # confer
    if ! type confer > /dev/null; then
        echo -e "Please install 'confer' to read global.yaml config"
        exit 1
    fi
    # shyaml
    if ! type shyaml > /dev/null; then
        echo -e "Please install 'shyaml', just type:\nsudo pip install shyaml"
        exit 1
    fi
}

# kill hyperchain process
f_kill_process(){
    echo "kill the bind port process"
    PID=`ps -ax | grep hyperchain | grep -v grep | awk '{print $1}'`
    if [ "$PID" != "" ]
    then
        ps -ax | grep hyperchain | grep -v grep | awk '{print $1}' | xargs kill -9
    fi
}

# clear data
f_delete_data(){
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
    cd ${PROJECT_PATH} && govendor build -o ${DUMP_PATH}/hyperchain -tags=embed
}

# rebuild hypercli
f_rebuild_hypercli(){
echo "Rebuild hypercli ..."
cd ${CLI_PATH} && govendor build
}

# modify peerconfig.json to local_peerconfig.json in global.yaml
f_modif_global(){
    sed -i "s/\/peerconfig.json/\/local_peerconfig.json/g" ${PROJECT_PATH}/scripts/namespace/config/template/config/global.yaml
}

start_hyperjvm() {
    cd ${PROJECT_PATH}/core/vm/jcee/java && ./build.sh
    for j in  1 2 3 4
    do
        cp -rf ${PROJECT_PATH}/core/vm/jcee/java/hyperjvm ${DUMP_PATH}/node$j/
    done
}

f_all_in_one_cmd(){
    cd $DUMP_PATH/node${1} && ./hyperchain -o ${1} -l 800${1} -t 808${1} -f 900${1} &
}

f_x_in_linux_cmd(){
    gnome-terminal -x bash -c "cd ${DUMP_PATH}/node${1} && ./hyperchain"
}

f_x_in_mac_cmd(){
    osascript -e 'tell app "Terminal" to do script "cd '$DUMP_PATH/node${1}' && ./hyperchain "'
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

# namespaces name, default is global
NS="global"

PROJECT_PATH="${GOPATH}/src/hyperchain"

# work path
CURRENT_PATH=`pwd`

# output root dir
DUMP_PATH="${PROJECT_PATH}/build"

# config file path
CONF_PATH="${PROJECT_PATH}/configuration"

# global config path
GLOBAL_CONFIG="${CONF_PATH}/namespaces/global/config/global.yaml"

# hypercli root path
CLI_PATH="${PROJECT_PATH}/hypercli"

# peerconfig
PEER_CONFIG_FILE=${PROJECT_PATH}/scripts/namespace/config/global.yaml

# delete data? default = true
DELETEDATA=true

# rebuild the project or not? default = true
REBUILD=true

# rebuild hypercli or not? default = false
HYPERCLI=false

# 1.check local env
f_check_local_env

# node num
MAXPEERNUM=`cat ${PEER_CONFIG_FILE} | shyaml get-value maxpeernode`
echo "Node number is: ${MAXPEERNUM}"

# 2.set system type
f_set_env

# exe by extra input params
while [ $# -gt 0 ]
do
    case "$1" in
    -h|--help)
        f_help; exit 0;;
    -k|--kill)
        f_kill_process; exit 1;;
    -d|--delete)
        DELETEDATA=false;
        shift;;
    -r|--rebuild)
        REBUILD=false;
        shift;;
    -c|--hypercli)
        HYPERCLI=true;
        shift;;
    -m|--mode)
        MODE=true;
        shift;;
    --) shift; break;;
    -*) f_help; exit 1;;
    *) NS="$NS $@"; break;;
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

f_modif_global

if $HYPERCLI ; then
    f_rebuild_hypercli
fi

# distribute files
for ns in $NS
do
    ${PROJECT_PATH}/scripts/namespace/gen_config.sh -l ${ns}
done

# run hyperchain node
start_hyperjvm

# run hyperchain node
f_run_process
