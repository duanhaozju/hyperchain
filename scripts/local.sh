#!/usr/bin/env bash
#set -xev
# judge system type varible may `MAC` or `LINUX`
_SYSTYPE="MAC"
case "$OSTYPE" in
  darwin*)
    echo "RUN SCRIPTS ON OSX"
    ENV=false
  ;;
  linux*)
    echo "RUN SCRIPTS ON LINUX"
    ENV=true
  ;;
  *)
    echo "unknown: $OSTYPE"
    exit -1
  ;;
esac

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

# check the local running env
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
#kill the progress
f_kill_process(){
    echo "Kill the bind port process"
    for((i=1;i<=$MAXPEERNUM;i++))
    do
        temp_port=`lsof -i :800$i | awk 'NR>=2{print $2}'`
        if [ x"$temp_port" != x"" ];then
            kill -9 $temp_port
        fi
    done
}




cd ../
# Build the project
PROJECT_PATH="${GOPATH}/src/hyperchain"
CURRENT_PATH=`pwd`
DUMP_PATH="${CURRENT_PATH}/build"
CONF_PATH="${CURRENT_PATH}/config"

# Load the config files
GLOBAL_CONFIG="${CONF_PATH}/global.yaml"
PEER_CONFIGS_FILE=`confer read global.yaml global.configs.peers`


PEER_CONFIG="${CONF_PATH}/local_peerconfig.json"
MAXPEERNUM=`cat ${PEER_CONFIG} | jq ".maxpeernode"`
echo "Node number is: ${MAXPEERNUM}"

DELETEDATA=true
REBUILD=true
MODE=true

while [ $# -gt 0 ]
do
    case "$1" in
    -h|--help)
        help; exit 1;;
    -k|--kill)
        killProcess; exit 1;;
	-d|--delete)
	    DELETEDATA=false; shift;;
	-r|--rebuild)
	    REBUILD=false; shift;;
    -m|--mode)
        MODE=false; shift;;
	--) shift; break;;
	-*) help; exit 1;;
	*) break;;
    esac
done

if $DELETEDATA; then
    # Clear the old data
    if [ -d "${DUMP_PATH}" ];then
        echo "Clear the old data..."
        rm -rf "${DUMP_PATH}/build"
        rm -rf "${DUMP_PATH}/config"
    fi

    # Creat the build dir
    if [ ! -d "${DUMP_PATH}" ];then
        echo "Auto creat the build dir..."
        mkdir -p "${DUMP_PATH}/build"
    fi
fi

# Check all the config files
echo "copy config dir into build dir.."
cp -rf "${CONF_PATH}" "${DUMP_PATH}/"
cp -rf "${CONF_PATH}/keystore" "${DUMP_PATH}/build/"

# cp the config files into nodes
for((j=1;j<=$MAXPEERNUM;j++))
do
    mkdir -p ${DUMP_PATH}/node${j}/
    cp -rf  ${CONF_PATH} ${DUMP_PATH}/node${j}/
    cp -rf  ${CONF_PATH}/peerconfigs/local_peerconfig_${j}.json ${DUMP_PATH}/node${j}/config/local_peerconfig.json

done

killProcess

if $REBUILD; then
    # Build the project
    echo "Build the project..."
    if [ -s "${DUMP_PATH}/hyperchain" ]; then
        rm ${DUMP_PATH}/hyperchain
    fi
     govendor build -o ${DUMP_PATH}/hyperchain
fi

# cp the hyperchain files into nodes

for((j=1;j<=$MAXPEERNUM;j++))
do
    cp -rf ${DUMP_PATH}/hyperchain ${DUMP_PATH}/node${j}/
done

cd ${DUMP_PATH}

# Run all the nodes
runXinXinLinux(){
    for((j=1;j<=$MAXPEERNUM;j++))
    do
        gnome-terminal -x bash -c "cd ${DUMP_PATH}/node${j} && ./hyperchain -o ${j} -t 808${j} -f 900${j}"
    done
}
runXinXinMac(){
    for((j=1;j<=$MAXPEERNUM;j++))
    do
        osascript -e 'tell app "Terminal" to do script "cd '$DUMP_PATH/node${j}' && ./hyperchain -o '${j}' -l 800'${j}' -t 808'${j}'"'
    done
}
runXin1(){
    for((j=1;j<=$MAXPEERNUM;j++))
    do
        ./hyperchain -o ${j} -l 800${j} -t 808${j} -f 900${j} -i true &
    done
}

echo "Run all the nodes..."
echo $ENV
if [ ! $MODE ]; then
    runXin1
else
    if [[ "$_SYSTYPE"x == 'MACx' ]]; then
        runXinXinMac
    else
        runXinXinLinux
    fi
fi
