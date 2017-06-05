#!/usr/bin/env bash
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

PASSWD="hyperchain"
PRIMARY=`head -1 ./serverlist.txt`
MAXNODE=`cat serverlist.txt | wc -l`

while IFS='' read -r line || [[ -n "$line" ]]; do
	SERVER_ADDR+=" ${line}"
done < serverlist.txt


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
        ssh hyperchain@${server_address} "pkill hyperchain"
        ssh hyperchain@${server_address} " cd /home/hyperchain/node${ni}/hyperjvm/bin && ./stop_hyperjvm.sh  "
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
        spawn ssh-copy-id hyperchain@$1
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
        spawn ssh-copy-id hyperchain@$1
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
	scp ./sub_scripts/server_addkey.sh hyperchain@$PRIMARY:/home/hyperchain/
	scp innerserverlist.txt hyperchain@$PRIMARY:/home/hyperchain/
	ssh  -T hyperchain@$PRIMARY "cd /home/hyperchain && chmod a+x server_addkey.sh && bash server_addkey.sh $SERVER_ENV"
}

# distribute the binary into primary
# the dir variblies
HPC_PRI_HYPERCHAIN_HOME="/home/hyperchain/"
HPC_PRI_HYPERCHAIN_GO_SRC="/home/hyperchain/go/src"
HPC_PRI_HYPERCHAIN_DIR="/home/hyperchain/go/src/hyperchain"
HPC_OTHER_HYPERCHAIN_DIR="/home/hyperchain"
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
    scp hyperchain.tar.gz hyperchain@${PRIMARY}:${HPC_PRI_HYPERCHAIN_HOME}
    ssh hyperchain@$PRIMARY "rm -rf $HPC_PRI_HYPERCHAIN_DIR"
    ssh hyperchain@$PRIMARY "tar -C $HPC_PRI_HYPERCHAIN_GO_SRC -zxmf hyperchain.tar.gz"
    echo "Primary build the project:"
	ssh -T hyperchain@$PRIMARY <<EOF
    if ! type go > /dev/null; then
        echo -e "Please install the go env correctly!"
        exit 1
    fi

    if [ `which govendor`x == "x" ]; then
        echo -e "Please install the govendor, just type:\ngo get -u github.com/kardianos/govendor"
        exit 1
    fi
    if [ ! -d "/home/hyperchain" ]; then
        mkdir /home/hyperchain/
    fi
    source ~/.bashrc && \
    cd go/src/hyperchain && \
    govendor build -tags=embed && \
    mv hyperchain /home/hyperchain/
EOF
    scp innerserverlist.txt hyperchain@${PRIMARY}:${HPC_PRI_HYPERCHAIN_HOME}
	scp ${HYPERCHAIN_DIR}/scripts/sub_scripts/server_deploy.sh hyperchain@${PRIMARY}:${HPC_PRI_HYPERCHAIN_HOME}

    echo "Primary send binary to others:"
	ssh hyperchain@${PRIMARY} "chmod a+x server_deploy.sh && bash server_deploy.sh ${MAXNODE}"
}

# modifiy the global config value
fs_modifi_global(){
    if [ ${_SYSTYPE} = "MAC" ]; then
        sed -i "" "s/local_peerconfig.json/peerconfig.json/g" ${HYPERCHAIN_DIR}/scripts/namespace/config/template/config/global.yaml
    else
        sed -i "s/local_peerconfig.json/peerconfig.json/g" ${HYPERCHAIN_DIR}/scripts/namespace/config/template/config/global.yaml
    fi
}

# generate the peer configs and distribute it
fs_gen_and_distribute_peerconfig(){
    fs__generate_node_peer_configs
    fs__distribute_peerconfigs
}

# generate peer configs
fs__generate_node_peer_configs(){
    ${GOPATH}/src/hyperchain/scripts/namespace/gen_config.sh global
}

# distribute config files
fs__distribute_peerconfigs(){
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
    echo "distribute config files to ${server_address}"
    ssh -T hyperchain@${server_address} <<EOF
    if [ -d ${HPC_OTHER_HYPERCHAIN_DIR}/node${ni} ]; then
        rm -rf ${HPC_OTHER_HYPERCHAIN_DIR}/node${ni}
    fi
EOF

#    copy config files
    cd ${GOPATH}/src/hyperchain/build
    tar cvzf node${ni}.tar.gz node${ni}
    scp node${ni}.tar.gz hyperchain@${server_address}:${HPC_OTHER_HYPERCHAIN_DIR}
    ssh -T hyperchain@${server_address} <<EOF
    if [ -d ${HPC_OTHER_HYPERCHAIN_DIR}/node${ni} ]; then
        rm -rf ${HPC_OTHER_HYPERCHAIN_DIR}/node${ni}
    fi
    tar zxvmf node${ni}.tar.gz
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
        scp -r ${PROJECT_PATH}/core/vm/jcee/java/hyperjvm hyperchain@${server_address}:/home/hyperchain/node${ni}
        ni=`expr ${ni} + 1`
    done
}
# Run all the nodes
# Open N Terminals in linux
fs_run_N_terminals_linux(){
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        gnome-terminal -x bash -c \
        "ssh hyperchain@$server_address \" cd /home/hyperchain/node${ni} && ./hyperchain \""
        ni=`expr ${ni} + 1`
    done
}
# Open N Terminals in mac
fs_run_N_terminals_mac(){
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        osascript -e 'tell app "Terminal" to do script "ssh hyperchain@'${server_address}' \" cd /home/hyperchain/node'${ni}' && ./hyperchain \""'
        ni=`expr ${ni} + 1`
    done
}

# run hyperchain N node in one Terminal
fs_run_one_terminal(){
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        ssh -T hyperchain@${server_address} "./hyperchain" &
        ni=`expr ${ni} + 1`
    done
}

# clean the data
fs_delete_data(){
    echo "Delete all the old data"
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        ssh -T hyperchain@${server_address} "rm -rf node${ni}"
        ni=`expr ${ni} + 1`
    done
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

fs_modifi_global

fs_gen_and_distribute_peerconfig

fs_distribute_jvm

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
