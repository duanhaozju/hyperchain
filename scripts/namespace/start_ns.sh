#!/usr/bin/env bash

# check local run environment
f_check_local_env(){
    if ! type shyaml > /dev/null; then
        echo -e "Please install 'shyaml', just type:\nsudo pip install shyaml"
        exit 1
    fi

    # check if config files exist
    for name in ${NS_NAMES}
    do
        if [ ! -f ${NS_CONFIG_PATH}/${name}${NS_CONFIG_SUFFIX} ]; then
            echo "config file ${NS_CONFIG_PATH}/${name}${NS_CONFIG_SUFFIX} does't exist!"
            exit 1
        fi
    done
}

# help prompt message
f_help(){
    echo "start_ns.sh helper:"
    echo "specify namespace config file name to start a namespace"
    echo "---------------------------------------------------"
    echo "Example for start a namespace:"
    echo "./start_ns ns1"
}

# rebuild hypercli
f_rebuild(){
    echo "rebuild hypercli..."
    cd ${CLI_PATH} && govendor build -tags=embed
}

f_get_ip_nodes(){
    NS_CONFIG_FILE=${NS_CONFIG_PATH}/${1}${NS_CONFIG_SUFFIX}
    NS_MAXNODE=`cat ${NS_CONFIG_FILE} | shyaml get-value maxpeernode`
    NODES=`cat ${NS_CONFIG_FILE} | grep -v "^#" | grep "^node[0-9]\{1,\}:" | awk -F ':' '{print $1}'`
    NODE_NUMS=`cat ${NS_CONFIG_FILE} | grep -v "^#" | grep "^node[0-9]\{1,\}:" | grep -c "node"`
    if [ ${NS_MAXNODE} -ne ${NODE_NUMS} ]; then
        echo "maxpeernode not equal to actual nodes provided in ${NS_CONFIG_FILE}"
        exit 1
    fi

    i=0
    for node in ${NODES}
    do
        IPS[$i]=`cat ${NS_CONFIG_FILE} | shyaml get-value ${node}.ip`
        PORTS[$i]=`cat ${NS_CONFIG_FILE} | shyaml get-value ${node}.jsonrpc_port`
        ((i+=1))
    done
}

# start namespace by specified ns config file
f_start_ns(){
    echo "ready to start namespace: $1"
    f_get_ip_nodes $1

    # register nodes
    for i in `seq 0 $[${NS_MAXNODE}-1]`
    do
        if [ $? -eq 0 ]; then
            echo "register ${IPS[$i]}:${PORTS[$i]} to namespace: $1"
            ${CLI_PATH}/hypercli -H ${IPS[$i]} -P ${PORTS[$i]} namespace register $1
        else
            echo "register failed!!!"
            exit 1
        fi
    done

    # start nodes
    for i in `seq 0 $[${NS_MAXNODE}-1]`
    do
        if [ $? -eq 0 ] ; then
            echo "start ${IPS[$i]}:${PORTS[$i]} in namespace: $1"
            ${CLI_PATH}/hypercli -H ${IPS[$i]} -P ${PORTS[$i]} namespace start $1
        else
            echo "start failed!!!"
            exit 1
        fi
    done
}

#####################################
#                                   #
#  MAIN INVOKE AREA                 #
#                                   #
#####################################
PROJECT_PATH="${GOPATH}/src/hyperchain"

# hypercli path
CLI_PATH="${PROJECT_PATH}/hypercli"

# namespace config file name
NS_CONFIG_SUFFIX=".yaml"

# namespace config path
NS_CONFIG_PATH="${PROJECT_PATH}/scripts/namespace/config"

# rebuild hypercli or not
REBUILD=false

num=0
# exe by extra input params
if [ $# -gt 0 ]; then
    while [ $# -gt 0 ]
    do
        case "$1" in
        -h|--help)
            f_help; exit 0;;
        -r|--rebuild)
            REBUILD=true;
            shift;;
        -*) f_help; exit 1;;
        *) NS_NAMES[$num]=$1; let num+=1; shift;;
        esac
    done
else
    echo "not enough params, at least one param"
    exit 1
fi
let num-=1

# 1.check local env
f_check_local_env

if ${REBUILD}; then
    f_rebuild
fi

for n in `seq 0 ${num}`
do
    f_start_ns ${NS_NAMES[$n]}
done
