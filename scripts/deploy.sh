#!/bin/bash
# this script is used for auto deploy the hyperchain project into gopath
# author:ChenQuan
# date: 2016-09-05
# usage: bash deploy.sh

# Stop on first error
set -e

# auto install the golang env
version=`go version | awk '{print $3}'`
if [ x"$version" != x"go1.7" ]; then
	echo "go hasn't installed!,try to export the env path..."
	export PATH=$PATH:/usr/local/go/bin
fi

if [ x"$version" != x"go1.7" ]; then
	echo "go hasn't installed!,please re install go"
else
	export GOPATH=$HOME/gopath
	go env
	go get -u github.com/kardianos/govendor
	mkdir -p $GOPATH/src
	cd $GOPATH/src
	git clone git@git.hyperchain.cn:hyperchain/hyperchain.git
	ls $GOPATH
fi