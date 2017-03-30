#!/usr/bin/env bash

set -e
# check local run environment
f_check_local_env(){
    if ! type shyaml > /dev/null; then
        echo -e "Please install 'shyaml', just type:\nsudo pip install shyaml"
        exit 1
    fi

    if [ ! -f ${NS_CONFIG_FILE} ]; then
    echo "config file ${NS_CONFIG_FILE} does't exist!"
    exit 1
    fi
}

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


f_get_nodes(){
    NODES=`cat ${NS_CONFIG_FILE} | grep -v "^#" | grep "node[0-9]\{1,\}:" | awk -F ':' '{print $1}'`
    NS_NUMS=`cat ${NS_CONFIG_FILE} | grep -v "^#" | grep "node[0-9]\{1,\}:" | grep -c "node"`
    if [ ! ${NS_MAXNODE} == ${NS_NUMS} ]; then
        echo "maxpeernode not equal to actual nodes provided in ${NS_CONFIG_FILE}"
        exit 1
    fi

    i=1
    for node in ${NODES}
    do
        NS_NODES[$i]=${node}
        ((i+=1))
    done
}

# help prompt message
f_help(){
    echo "gen_config.sh helper:"
    echo "specify namespace config file name to generate and distribute nodes dirs"
    echo "---------------------------------------------------"
    echo "Example for generate and distribute a namespace:"
    echo "./gen_config ns1"
}

f_copy_tmp(){
    if [ -d ${NS_PATH}/${NS_NAME} ]; then
        echo "delete existed title namespace config root files '${NS_NAME}'..."
        rm -rf ${NS_PATH}/${NS_NAME}
    fi
    echo "create namespace directory '${NS_NAME}'..."
    mkdir -p ${NS_PATH}/${NS_NAME}

    # copy template config
    cp -r ${TMP_PATH}/* ${NS_PATH}/${NS_NAME}/

    # global is the template namespace name
    if [ ${_SYSTYPE} = "MAC" ]; then
        sed -i "" "s/namespaces\/template/namespaces\/${NS_NAME}/" ${NS_PATH}/${NS_NAME}/config/db.yaml
        sed -i "" "s/namespaces\/template/namespaces\/${NS_NAME}/" ${NS_PATH}/${NS_NAME}/config/global.yaml
        sed -i "" "s/namespaces\/template/namespaces\/${NS_NAME}/" ${NS_PATH}/${NS_NAME}/config/caconfig.toml
    else
        sed -i "s/namespaces\/template/namespaces\/${NS_NAME}/g" ${NS_PATH}/${NS_NAME}/config/db.yaml
        sed -i "s/namespaces\/template/namespaces\/${NS_NAME}/" ${NS_PATH}/${NS_NAME}/config/global.yaml
        sed -i "s/namespaces\/template/namespaces\/${NS_NAME}/" ${NS_PATH}/${NS_NAME}/config/caconfig.toml
    fi
}

f_gen_config(){
    # remove useless local {peerconfig}
    rm -rf ${NS_PATH}/${NS_NAME}/config/peerconfigs/local_peerconfig_*

    for ((i=1; i<= $NS_MAXNODE; i++)); do
        # create cert dir if not exist
        if [ -d ${NS_PATH}/${NS_NAME}/config/peerconfigs/${NS_NODES[$i]} ]; then
            rm -rf ${NS_PATH}/${NS_NAME}/config/peerconfigs/${NS_NODES[$i]}
        fi
        mkdir ${NS_PATH}/${NS_NAME}/config/peerconfigs/${NS_NODES[$i]}
        cp -r ${TMP_PATH}/config/peerconfigs/node1/* ${NS_PATH}/${NS_NAME}/config/peerconfigs/${NS_NODES[$i]}/

        peerconfig=${NS_PATH}/${NS_NAME}/config/peerconfigs/local_peerconfig_${i}.json
        touch ${peerconfig}
        ip=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$i]}.ip`
        node_id=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$i]}.node_id`
        is_origin=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$i]}.is_origin`
        is_origin=$(echo ${is_origin} | tr A-Z a-z)
        is_vp=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$i]}.is_vp`
        is_vp=$(echo ${is_vp} | tr A-Z a-z)
        grpc_port=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$i]}.grpc_port`
        jsonrpc_port=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$i]}.jsonrpc_port`
        restful_port=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$i]}.restful_port`
        introducer_ip=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$i]}.introducer_ip`
        introducer_port=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$i]}.introducer_port`
        introducer_id=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$i]}.introducer_id`

        echo "{"                                        >>${peerconfig}
        echo "  \"self\":{"                             >>${peerconfig}
        echo "  \"is_origin\":$is_origin,"              >>${peerconfig}
        echo "  \"is_vp\":$is_vp,"                      >>${peerconfig}
        echo "  \"node_id\":$node_id,"                  >>${peerconfig}
        echo "  \"grpc_port\":$grpc_port,"              >>${peerconfig}
        echo "  \"local_ip\":\"127.0.0.1\","            >>${peerconfig}
        echo "  \"jsonrpc_port\":$jsonrpc_port,"        >>${peerconfig}
        echo "  \"restful_port\":$restful_port,"        >>${peerconfig}
        echo "  \"introducer_ip\":\"$introducer_ip\","  >>${peerconfig}
        echo "  \"untroducer_port\":$introducer_port,"  >>${peerconfig}
        echo "  \"introducer_id\":$introducer_id"       >>${peerconfig}
        echo " },"                                      >>${peerconfig}
        echo "  \"maxpeernode\":$NS_MAXNODE,"           >>${peerconfig}
        echo "  \"nodes\":["                            >>${peerconfig}

        for ((j = 1; j <= $NS_MAXNODE; j++)); do
            # reachable ip address
            ip=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$j]}.ip`
            grpc_port=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$j]}.grpc_port`
            jsonrpc_port=`cat ${NS_CONFIG_FILE} | shyaml get-value ${NS_NODES[$j]}.jsonrpc_port`
            echo "    {"                                    >>${peerconfig}
            echo "    \"id\":$j,"                           >>${peerconfig}
            echo "    \"address\":\"$ip\","                 >>${peerconfig}
            echo "    \"external_address\":\"127.0.0.1\","  >>${peerconfig}
            echo "    \"port\":$grpc_port,"                 >>${peerconfig}
            echo "    \"rpc_port\":$jsonrpc_port"           >>${peerconfig}
            echo "  },"                                     >>${peerconfig}
        done
        if [ ${_SYSTYPE} = "MAC" ]; then
            sed -i "" '$d' ${peerconfig}
        else
            sed -i '$d' ${peerconfig}
        fi
        echo "  }"                                      >>${peerconfig}
        echo "]"                                        >>${peerconfig}
        echo "}"                                        >>${peerconfig}
    done
}

# distribute node package
f_distribute(){
    # cp the config files into nodes
    for (( j=1; j<=$NS_MAXNODE; j++ ))
    do
        if [ ! -d "${DUMP_PATH}/${NS_NODES[$j]}/namespaces" ]; then
            mkdir -p ${DUMP_PATH}/${NS_NODES[$j]}/namespaces
        fi

        if [ ${NS_NAME} == "global" ];then
            cp -rf  ${NS_CONFIG_PATH}/global.yaml ${DUMP_PATH}/${NS_NODES[$j]}
            cp -rf  ${NS_CONFIG_PATH}/LICENSE ${DUMP_PATH}/${NS_NODES[$j]}
        fi

        if [ -d "${DUMP_PATH}/${NS_NODES[$j]}/namespaces/${NS_NAME}" ];then
            echo "delete existed title namespace files in dir ${NS_NODES[$j]}"
            rm -rf ${DUMP_PATH}/${NS_NODES[$j]}/namespaces/${NS_NAME}
        fi
        mkdir ${DUMP_PATH}/${NS_NODES[$j]}/namespaces/${NS_NAME}

        cp -rf  ${NS_PATH}/${NS_NAME}/* ${DUMP_PATH}/${NS_NODES[$j]}/namespaces/${NS_NAME}
        cp -rf  ${NS_PATH}/${NS_NAME}/config/peerconfigs/local_peerconfig_${j}.json ${DUMP_PATH}/${NS_NODES[$j]}/namespaces/${NS_NAME}/config/local_peerconfig.json
        cp -rf  ${NS_PATH}/${NS_NAME}/config/peerconfigs/node${j}/* ${DUMP_PATH}/${NS_NODES[$j]}/namespaces/${NS_NAME}/config/cert/
        cp -rf  ${DUMP_PATH}/hyperchain ${DUMP_PATH}/${NS_NODES[$j]}
    done
}

#####################################
#                                   #
#  MAIN INVOKE AREA                 #
#                                   #
#####################################
# system type
_SYSTYPE="LINUX"

PROJECT_PATH="${GOPATH}/src/hyperchain"

# output root dir
DUMP_PATH="${PROJECT_PATH}/build"

#template config file path
TMP_PATH="${PROJECT_PATH}/scripts/namespace/config/template"

# namespace config file name
NS_CONFIG_SUFFIX=".yaml"
NS_CONFIG_FILE=""

NS_CONFIG_PATH="${GOPATH}/src/hyperchain/configuration"

# config ns path
NS_PATH="${NS_CONFIG_PATH}/namespaces"

# exe by extra input params
if [ $# -gt 0 ]; then
    case "$1" in
    -h|--help)
        f_help; exit 0;;
    -*) f_help; exit 1;;
    *) NS_CONFIG_FILE="${PROJECT_PATH}/scripts/namespace/config/$1${NS_CONFIG_SUFFIX}";;
    esac
else
    echo "not enough params, at least one param"
    exit 1
fi

# 1.check local env
f_check_local_env

# 2.set system type
f_set_env

# get namespace name from config file
NS_NAME=`cat ${NS_CONFIG_FILE} | shyaml get-value namespace`

# get max node number from config file
NS_MAXNODE=`cat ${NS_CONFIG_FILE} | shyaml get-value maxpeernode`
f_get_nodes

# 3.copy template config files to generate new config files
f_copy_tmp

# 4.generate new config files
f_gen_config

# 5.distribute config files to corresponding nodes
f_distribute