#!/usr/bin/env bash

FIRST=false
DELETEDATA=true
REBUILD=true
LOCAL_ENV=true
SERVER_ENV=true
MODE=true

PASSWD="hyperchain"
PRIMARY=`head -1 ./serverlist.txt`
MAXNODE=`cat serverlist.txt | wc -l`

while IFS='' read -r line || [[ -n "$line" ]]; do
	SERVER_ADDR+=" ${line}"
done < serverlist.txt

echo $SERVER_ADDR

# show the helper
help(){
    echo "local.sh helper:"
    echo "  -h, --help:     show the help for this bash script"
    echo "  -k, --kill:     just kill all the processes"
    echo "  -f, --first:    first-time run, do addkey;      default: false, add for first"
    echo "  -d, --delete:   clear the old data or not;      default: clear, add for not clear"
    echo "  -r, --rebuild:  rebuild the project or not;     default: rebuild, add for not rebuild"
    echo "  -l, --local:    which kink of local system;     default: linux, add for mac"
    echo "  -s, --server:   which kink of server system;    default: CentOS, add for SUSE"
    echo "  -m, --mode:     choose the run mode;            default: run many in many, add for many in one"
    echo "---------------------------------------------------"
    echo "Example for run many in one in mac for SUSE without rebuild:"
    echo "./server -l -m -r -s"
}

# kill all the process
killProcess(){
    echo "kill $server_address"
    for server_address in ${SERVER_ADDR[@]}; do
        ssh hyperchain@$server_address "ps aux | grep 'hyperchain -o' | awk '{print \$2}' | xargs kill -9"
    done
}

while [ $# -gt 0 ]
do
    case "$1" in
    -h|--help)
        help; exit 1;;
    -k|--kill)
        killProcess; exit 1;;
    -f|--first)
        FIRST=true; shift;;
	-d|--delete)
	    DELETEDATA=false; shift;;
	-r|--rebuild)
	    REBUILD=false; shift;;
    -l|--local)
        LOCAL_ENV=false; shift;;
    -s|--server)
        SERVER_ENV=false; shift;;
    -m|--mode)
        MODE=false; shift;;
	--) shift; break;;
	-*) help; exit 1;;
	*) break;;
    esac
done

addkeyForCentOS(){
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

addkeyForSuse(){
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

add_ssh_key_into_primary(){
    echo "Add your local ssh public key into primary node"
    if $SERVER_ENV; then
        for server_address in ${SERVER_ADDR[@]}; do
            addkeyForCentOS $server_address &
        done
    else
        for server_address in ${SERVER_ADDR[@]}; do
            addkeyForSuse $server_address &
        done
    fi
	wait
}

add_ssh_key_form_primary_to_others(){
    echo "Primary add its ssh key into others nodes"
	scp ./sub_scripts/server_addkey.sh hyperchain@$PRIMARY:/home/hyperchain/
	scp innerserverlist.txt hyperchain@$PRIMARY:/home/hyperchain/
	ssh  hyperchain@$PRIMARY "cd /home/hyperchain && chmod a+x server_addkey.sh && bash server_addkey.sh $SERVER_ENV"
}

distribute_the_binary(){
    echo "Send the project to primary:"
    cd $GOPATH/src/
    if [ -d "hyperchain/build" ]; then
        rm -rf hyperchain/build
    fi
    if [ -d "hyperchain.tar.gz" ]; then
        rm hyperchain.tar.gz
    fi
    tar -zcf hyperchain.tar.gz ./hyperchain
    scp hyperchain.tar.gz hyperchain@$PRIMARY:/home/hyperchain/
    ssh hyperchain@$PRIMARY "rm -rf go/src/hyperchain"
    ssh hyperchain@$PRIMARY "tar -C /home/hyperchain/go/src -xzf hyperchain.tar.gz"
    echo "Primary build the project:"
	ssh hyperchain@$PRIMARY "source ~/.bashrc && cd go/src/hyperchain && govendor build && mv hyperchain /home/hyperchain"

	echo "Send the config files to primary:"
	cd $GOPATH/src/hyperchain/scripts
	scp -r ../config/ hyperchain@$PRIMARY:/home/hyperchain/
	scp ./sub_scripts/server_deploy.sh hyperchain@$PRIMARY:/home/hyperchain/

    echo "Primary send files to others:"
	ssh hyperchain@$PRIMARY "chmod a+x server_deploy.sh && bash server_deploy.sh ${MAXNODE}"
}

# Run all the nodes
runXinXinLinux(){
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        gnome-terminal -x bash -c "ssh hyperchain@$server_address \" cd /home/hyperchain/ && cp -rf ./config/keystore ./build/ && ./hyperchain -o $ni -l 8001 -t 8081 -i true \""
        ni=`expr $ni + 1`
    done
}
runXinXinMac(){
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        osascript -e 'tell app "Terminal" to do script "ssh hyperchain@'$server_address' \" cd /home/hyperchain/ && cp -rf ./config/keystore ./build/ && ./hyperchain -o '$ni' -l 8001 -t 8081 -i true \""'
        ni=`expr $ni + 1`
    done
}
runXin1(){
    ni=1
    for server_address in ${SERVER_ADDR[@]}; do
        ssh hyperchain@$server_address "./hyperchain -o ${ni} -l 8001 -t 8081 -i true" &
        ni=`expr $ni + 1`
    done
}

deleteData(){
    echo "Delete all the old data"
    for server_address in ${SERVER_ADDR[@]}; do
        ssh hyperchain@$server_address "rm -rf build"
    done
}


if $FIRST; then
	add_ssh_key_into_primary
	add_ssh_key_form_primary_to_others
fi

killProcess
if $REBUILD; then
    if $DELETEDATA; then
        deleteData
    fi
    distribute_the_binary
elif $DELETEDATA; then
    deleteData
fi

echo "Run all the nodes..."
if [ ! $MODE ]; then
    runXin1
else
    if $LOCAL_ENV; then
        runXinXinLinux
    else
        runXinXinMac
    fi
fi