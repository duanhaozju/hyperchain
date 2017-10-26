#!/bin/bash

USERNAME="hyperchain"
PASSWD="hyperchain"
# debug flag
#set -evx
################
# pwd vars
################
CURRENT_DIR=`pwd`
HYPERCHAIN_DIR="$GOPATH/src/hyperchain"
# judge system type varible may `MAC` or `LINUX`
_SYSTYPE="MAC"
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

##################
# TOOL FUNCTIONS #
##################
# TOOL functions will start with 'f_'

f_trim(){
    word=$1
    echo -e "${word}" | tr -d '[:space:]'
}
f_trim_tail(){
    word=$1
    echo -e "${word}" | sed -e 's/^[[:space:]]*//'
}


#######################
# env check function  #
#######################
# env Check function start with `env`

env_check_serverlist_length(){
    serverlistlen=`cat serverlist.txt | wc -l`
    innerserverlistlen=`cat innerserverlist.txt | wc -l`
    if [ serverlistlen -ne innerserverlistlen ]; then
        echo "serverlist length not equal inner server list"
    fi
    exit 1
}

env_check_local_go_env(){
if ! type go > /dev/null; then
    echo -e "Please install the go env correctly!"
    exit 1
fi

if ! type govendor > /dev/null; then
    # install foobar here
    echo -e "Please install the 'govendor', just type:\ngo get -u github.com/kardianos/govendor"
    exit 1
fi
}

env_check_local_linux_env(){
if ! type jq > /dev/null; then
    echo -e "Please install the jq to parse the json file \n just type: \n sudo apt-get install jq / sudo yum -y install jq / brew install jq "
    exit 1
fi
if ! type confer > /dev/null; then
    echo -e "Please install the confer to generate the peer config json file"
    echo "now auto install the confer:"
    mkdir -p $GOPATH/src/git.hyperchain.cn/chenquan/ && cd $GOPATH/src/git.hyperchain.cn/chenquan/
    git clone git@git.hyperchain.cn:chenquan/confer.git
    cd $GOPATH/src/git.hyperchain.cn/chenquan/confer
    go install
    confer -h
fi
echo "check confer again:"
if ! type confer > /dev/null; then
    echo -e "please manully install confer,just follow those steps:"
    echo "mkdir -p $GOPATH/src/git.hyperchain.cn/chenquan/ && cd $GOPATH/src/git.hyperchain.cn/chenquan/"
    echo "git clone git@git.hyperchain.cn:chenquan/confer.git"
    echo "cd $GOPATH/src/git.hyperchain.cn/chenquan/confer"
    echo "go install"
    exit 1
fi

if [ -d $GOPATH/src/git.hyperchain.cn/chenquan/confer ]; then
    echo "update the 'confer' "
    cd $GOPATH/src/git.hyperchain.cn/chenquan/confer
    git clean -df && git checkout -- .
    git pull origin master
    go install
fi

cd $CURRENT_DIR
}

#################
# global flags  #
#################
# first time run this script or not, if first, will run addkey function
FIRST=false
# if delete the data or not
DELETEDATA=true
# if rebuild the hyperchain or not
REBUILD=true
# true: centos false: suse
SERVER_ENV=true
# false: open many terminals, true: output all logs in one terminal
MODE=false

PRIMARY=`head -1 ./serverlist.txt`
MAXNODE=`cat serverlist.txt | wc -l`

while IFS='' read -r line || [[ -n "$line" ]]; do
	SERVER_ADDR+=" ${line}"
done < serverlist.txt

while IFS='' read -r line || [[ -n "$line" ]]; do
	INNER_SERVER_ADDR+=" ${line}"
done < innerserverlist.txt


#################################
# functional support function
#################################
# those function will start with `fs_`

# show the helper
fs_help(){
    echo "local.sh helper:"
    echo "  -h, --help:     show the help for this bash script"
    echo "  -k, --kill:     just kill all the processes"
    echo "  -f, --first:    first-time run, do addkey;      default: false, add for first"
    echo "  -d, --delete:   clear the old data or not;      default: true, add for not to delete the data"
    echo "  -r, --rebuild:  rebuild the project or not;     default: rebuild, add for not rebuild"
    echo "  -s, --server:   which kink of server system;    default: CentOS, add for SUSE"
    echo "  -m, --mode:     choose the run mode;            default: run many in N terminals, add for run  in one terminal "
    echo "---------------------------------------------------"
    echo "Example for run many in one in mac for SUSE without rebuild:"
    echo "./server.sh -m -r -s"
    echo "means run server.sh for suse and in one terminal not rebuild"
}
# kill all the process
fs_kill_process(){
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        echo "kill process in ${server_address}"
        ssh ${USERNAME}@${server_address} "pkill hyperchain"
        ni=`expr ${ni} + 1`
    done
   }


# check the env
fs_checkenv(){
    env_check_serverlist_length
    env_check_local_go_env
    env_check_local_linux_env
}

# add ssh-key for CentOS
fs___addkey_for_centos(){
    expect <<EOF
        set timeout 60
        spawn ssh-copy-id ${USERNAME}@$1
        expect {
            "yes/no" {send "yes\r";exp_continue }
            "password:" {send "$PASSWD\r";exp_continue }
            eof
        }
EOF
}

# add ssh-key for OpenSuse
fs___addkey_for_suse(){
    expect <<EOF
        set timeout 60
        spawn ssh-copy-id ${USERNAME}@$1
        expect {
            "yes/no" {send "yes\r";exp_continue }
            "Password:" {send "$PASSWD\r";exp_continue }
            eof
        }
EOF
}

# add ssh-key into primary
fs_add_ssh_key_into_primary(){
    echo "Add your local ssh public key into primary node"
    if $SERVER_ENV; then
        for server_address in ${SERVER_ADDR[@]}; do
            fs___addkey_for_centos $server_address &
        done
    else
        for server_address in ${SERVER_ADDR[@]}; do
            fs___addkey_for_suse $server_address &
        done
    fi
	wait
}

# distribute the  Primary ssh-key into others
fs_add_ssh_key_form_primary_to_others(){
    echo "Primary add its ssh key into others nodes"
	scp ./sub_scripts/server_addkey.sh ${USERNAME}@$PRIMARY:/home/${USERNAME}/
	scp innerserverlist.txt ${USERNAME}@$PRIMARY:/home/${USERNAME}/
	ssh  -T ${USERNAME}@$PRIMARY "cd /home/${USERNAME} && chmod a+x server_addkey.sh && bash server_addkey.sh $SERVER_ENV"
}

# distribute the binary into primary
# the dir variblies
HPC_PRI_HYPERCHAIN_HOME="/home/${USERNAME}/"
HPC_PRI_HYPERCHAIN_GO_SRC="/home/${USERNAME}/go/src"
HPC_PRI_HYPERCHAIN_DIR="/home/${USERNAME}/go/src/hyperchain"
HPC_OTHER_HYPERCHAIN_DIR="/home/${USERNAME}"
fs_distribute_the_binary(){
    echo "Send the project to primary:"
    cd ${GOPATH}/src/
    if [ -d "hyperchain/build" ]; then
        rm -rf hyperchain/build
    fi
    if [ -d "hyperchain.tar.gz" ]; then
        rm hyperchain.tar.gz
    fi
    tar -zcf hyperchain.tar.gz ./hyperchain
    scp hyperchain.tar.gz ${USERNAME}@${PRIMARY}:${HPC_PRI_HYPERCHAIN_HOME}
    ssh ${USERNAME}@$PRIMARY "rm -rf $HPC_PRI_HYPERCHAIN_DIR"
    ssh ${USERNAME}@$PRIMARY "tar -C $HPC_PRI_HYPERCHAIN_GO_SRC -zxmf hyperchain.tar.gz"
    ssh ${USERNAME}@$PRIMARY "rm -rf hyperchain.tar.gz"
    echo "Primary build the project:"
	ssh -T ${USERNAME}@$PRIMARY <<EOF
    if ! type go > /dev/null; then
        echo -e "Please install the go env correctly!"
        exit 1
    fi

    if [ `which govendor`x == "x" ]; then
        echo -e "Please install the govendor, just type:\ngo get -u github.com/kardianos/govendor"
        exit 1
    fi
    if [ ! -d "/home/${USERNAME}" ]; then
        mkdir /home/${USERNAME}/
    fi
    source ~/.bashrc && \
    cd go/src/hyperchain && \
    govendor build -tags=embed && \
    mv hyperchain /home/${USERNAME}/
EOF
    scp ${GOPATH}/src/hyperchain/scripts/innerserverlist.txt ${USERNAME}@${PRIMARY}:${HPC_PRI_HYPERCHAIN_HOME}
	scp ${HYPERCHAIN_DIR}/scripts/sub_scripts/server_deploy.sh ${USERNAME}@${PRIMARY}:${HPC_PRI_HYPERCHAIN_HOME}

    echo "Primary send binary to others:"
	ssh ${USERNAME}@${PRIMARY} "chmod a+x server_deploy.sh && bash server_deploy.sh ${MAXNODE}"
}

# modifiy the global config value
#fs_modifi_global(){
#    if [ ${_SYSTYPE} = "MAC" ]; then
#        sed -i "" "s/local_peerconfig.json/peerconfig.json/g" ${HYPERCHAIN_DIR}/scripts/namespace/config/template/config/global.yaml
#    else
#        sed -i "s/local_peerconfig.json/peerconfig.json/g" ${HYPERCHAIN_DIR}/scripts/namespace/config/template/config/global.yaml
#    fi
#}

# generate the peer configs and distribute it
fs_gen_and_distribute_peerconfig(){
    fs__generate_node_peer_configs
    fs__distribute_peerconfigs
}

# generate peer configs
fs__generate_node_peer_configs(){
#    ${GOPATH}/src/hyperchain/scripts/namespace/gen_config.sh global
#    ${GOPATH}/src/hyperchain/scripts/namespace/gen_config.sh -c ns1
#    ${GOPATH}/src/hyperchain/scripts/namespace/gen_config.sh -c ns2
#    ${GOPATH}/src/hyperchain/scripts/namespace/gen_config.sh -c ns3
    ni=1
    mkdir ${GOPATH}/src/hyperchain/build
    for server_address in ${INNER_SERVER_ADDR[@]}; do
    echo "gen config"
    cd ${GOPATH}/src/hyperchain/build
    mkdir node${ni}
    cd node${ni}
    cp -rf ${GOPATH}/src/hyperchain/configuration/* ./
    cp -rf ${GOPATH}/src/hyperchain/configuration/peerconfigs/peerconfig_${ni}.toml ./namespaces/global/config/peerconfig.toml
    cp -rf ${GOPATH}/src/hyperchain/configuration/global.toml ./global.toml
    cp -rf ${GOPATH}/src/hyperchain/configuration/peerconfigs/addr_1.toml ./addr.toml
    if [ ${_SYSTYPE} = "MAC" ]; then
        sed -i "" "s/domain=\"domain1\"/domain=\"domain${ni}\"/g" ./addr.toml
    else
        sed -i "s/domain=\"domain1\"/domain=\"domain${ni}\"/g" ./addr.toml
    fi
    cp -rf  ${GOPATH}/src/hyperchain/configuration/peerconfigs/cert${ni}/* ./namespaces/global/config/certs/
    cp -rf  ${GOPATH}/src/hyperchain/configuration/tls ./
    cp  ${GOPATH}/src/hyperchain/configuration/tls/tls_peer${ni}.cert ./tls/tls_peer.cert
    cp  ${GOPATH}/src/hyperchain/configuration/tls/tls_peer${ni}.priv ./tls/tls_peer.priv
    cp  ${GOPATH}/src/hyperchain/configuration/tls/tlsca.ca ./tls/tlsca.ca
    mkdir ./bin
    cp ${GOPATH}/src/hyperchain/scripts/sub_scripts/start.sh ./bin
    ((ni+=1))
done
    ni=1
    for server_address in ${INNER_SERVER_ADDR[@]}; do
    if [ ${_SYSTYPE} = "MAC" ]; then
        sed -i "" "s/domain${ni} 127.0.0.1:50011/domain${ni} ${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node1/addr.toml
        sed -i "" "s/domain${ni} 127.0.0.1:50011/domain${ni} ${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node2/addr.toml
        sed -i "" "s/domain${ni} 127.0.0.1:50011/domain${ni} ${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node3/addr.toml
        sed -i "" "s/domain${ni} 127.0.0.1:50011/domain${ni} ${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node4/addr.toml
    else
        sed -i "s/domain${ni} 127.0.0.1:50011/domain${ni} ${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node1/addr.toml
        sed -i "s/domain${ni} 127.0.0.1:50011/domain${ni} ${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node2/addr.toml
        sed -i "s/domain${ni} 127.0.0.1:50011/domain${ni} ${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node3/addr.toml
        sed -i "s/domain${ni} 127.0.0.1:50011/domain${ni} ${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node4/addr.toml
    fi
    ((ni+=1))
done
    ni=1
    for server_address in ${INNER_SERVER_ADDR[@]}; do
    if [ ${_SYSTYPE} = "MAC" ]; then
        sed -i "" "s/127.0.0.1:5001${ni}/${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node1/hosts.toml
        sed -i "" "s/127.0.0.1:5001${ni}/${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node2/hosts.toml
        sed -i "" "s/127.0.0.1:5001${ni}/${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node3/hosts.toml
        sed -i "" "s/127.0.0.1:5001${ni}/${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node4/hosts.toml
    else
        sed -i "s/127.0.0.1:5001${ni}/${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node1/hosts.toml
        sed -i "s/127.0.0.1:5001${ni}/${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node2/hosts.toml
        sed -i "s/127.0.0.1:5001${ni}/${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node3/hosts.toml
        sed -i "s/127.0.0.1:5001${ni}/${server_address}:50011/g" ${GOPATH}/src/hyperchain/build/node4/hosts.toml
    fi
    ((ni+=1))
done
}

# distribute config files
fs__distribute_peerconfigs(){
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
    echo "distribute config files to ${server_address}"
    ssh -T ${USERNAME}@${server_address} <<EOF
    if [ -d ${HPC_OTHER_HYPERCHAIN_DIR}/node${ni} ]; then
        rm -rf ${HPC_OTHER_HYPERCHAIN_DIR}/node${ni}
    fi
EOF

#    copy config files
    cd ${GOPATH}/src/hyperchain/build
    tar cvzf node${ni}.tar.gz node${ni}
    scp node${ni}.tar.gz ${USERNAME}@${server_address}:${HPC_OTHER_HYPERCHAIN_DIR}
    ssh -T ${USERNAME}@${server_address} <<EOF
    if [ -d ${HPC_OTHER_HYPERCHAIN_DIR}/node${ni} ]; then
        rm -rf ${HPC_OTHER_HYPERCHAIN_DIR}/node${ni}
    fi
    tar zxvmf node${ni}.tar.gz
    rm -rf node${ni}.tar.gz
    cp ${HPC_OTHER_HYPERCHAIN_DIR}/hyperchain ${HPC_OTHER_HYPERCHAIN_DIR}/node${ni}
EOF
    rm -rf node${ni}.tar.gz
    ((ni+=1))
done
}

# Build jvm in local machine and then distribute to servers
fs_distribute_jvm(){
    echo "build hyperjvm in local"
    cd ${PROJECT_PATH}/core/vm/jcee/java && ./build.sh

    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        echo "distribute hyperjvm to ${server_address}"
        scp -r ${PROJECT_PATH}/core/vm/jcee/java/hyperjvm ${USERNAME}@${server_address}:/home/${USERNAME}/node${ni}
        ssh ${USERNAME}@${server_address} " cd /home/${USERNAME}/node${ni}/hyperjvm/bin && ./stop_hyperjvm.sh"
        ni=`expr ${ni} + 1`
    done
}
# Run all the nodes
# Open N Terminals in linux
fs_run_N_terminals_linux(){
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        gnome-terminal -x bash -c \
        "ssh ${USERNAME}@$server_address \" export PATH=$PATH:/usr/sbin && cd /home/${USERNAME}/node${ni} && ./hyperchain 2>error.log \""
        ni=`expr ${ni} + 1`
    done
}
# Open N Terminals in mac
fs_run_N_terminals_mac(){
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        osascript -e 'tell app "Terminal" to do script "ssh '${USERNAME}'@'${server_address}' \" cd /home/'${USERNAME}'/node'${ni}' && export LD_LIBRARY_PATH=/usr/local/ssl/lib && ./hyperchain --pprof --pport 10091 2>error.log \""'
        ni=`expr ${ni} + 1`
    done
}

# run hyperchain N node in one Terminal
fs_run_one_terminal(){
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        ssh -T ${USERNAME}@${server_address} "./hyperchain" &
        ni=`expr ${ni} + 1`
    done
}

# clean the data
fs_delete_data(){
    echo "Delete all the old data"
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        ssh -T ${USERNAME}@${server_address} "rm -rf node${ni}"
        ni=`expr ${ni} + 1`
    done
}

# clean files in primary node
fs_delete_files(){
    echo "Delete codes on primary"
    ssh ${USERNAME}@$PRIMARY "rm -rf $HPC_PRI_HYPERCHAIN_DIR"
}


######################
# RUN ALL PROGRAMS
#####################

PROJECT_PATH="${GOPATH}/src/hyperchain"

# parse the flags
while [ $# -gt 0 ]
do
    case "$1" in
    -h|--help)
        fs_help; exit 1;;
    -k|--kill)
        fs_kill_process; exit 1;;
    -f|--first)
        FIRST=true; shift;;
	-d|--delete)
	    DELETEDATA=false; shift;;
	-r|--rebuild)
	    REBUILD=false; shift;;
    -s|--server)
        SERVER_ENV=false; shift;;
    -m|--mode)
        MODE=true; shift;;
	--) shift; break;;
	-*) help; exit 1;;
	*) break;;
    esac
done

#echo "run this script first time? $FIRST"
#echo "delete the data? $DELETEDATA"
#echo "rebuild and redistribute binary? $REBUILD"
#echo "server env,true: suse,false: centos: $SERVER_ENV"

if ${FIRST}; then
    fs_add_ssh_key_into_primary
    fs_add_ssh_key_form_primary_to_others
    exit 0
fi
# kill all processes
fs_kill_process

if ${DELETEDATA}; then
    fs_delete_data
fi

if ${REBUILD}; then
    fs_distribute_the_binary
fi

#fs_modifi_global

fs_gen_and_distribute_peerconfig

fs_distribute_jvm

fs_delete_files

echo "Running nodes"

if ${MODE}; then
    fs_run_one_terminal
else
    if [ "${_SYSTYPE}x" == "MACx" ]; then
        fs_run_N_terminals_mac
    else
        fs_run_N_terminals_linux
    fi
fi
