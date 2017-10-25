#!/usr/bin/env bash

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

    if ! type git > /dev/null; then
        echo -e "Please install the git env correctly!"
        exit 1
    fi
}

# build hyperchain with version information
f_build_hyperchain(){
    branch=`git rev-parse --abbrev-ref HEAD`
    commitID=`git log --pretty=format:"%h" -1`
    date=`date +%Y%m%d`

    if [ -f ${1}/hyperchain ]; then
        rm ${1}/hyperchain
    fi

    cd ${1} && govendor build -ldflags "-X main.branch=${branch} -X main.commitID=${commitID} -X main.date=${date}"
}

f_check_local_env

if [ $# -eq 0 ]; then
    f_build_hyperchain ${GOPATH}/src/hyperchain
elif [ $# -eq 1 ]; then
    f_build_hyperchain ${1}
else
    echo "Please specify hyperchain directory to build hyperchain."
fi