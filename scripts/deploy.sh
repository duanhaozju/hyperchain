#!/bin/bash
# Author: ChenQuan
# Description: this script is used for deploy code env in the server side.
# Date: 2016-09-15
# Happy mid-moon day!

set -e
set -x
sudo apt-update --yyq
sudo apt-upgrade
sudo apt-get -y install git build-essential wget

GOROOT=/usr/local/go
cd /tmp
mkdir gopkg
wget --no-check-certificate https://api.google.com/golang/go1.7.1-amd64.tar.gz
sudo tar zxf go1.7.1-amd64.tar.gz -c /usr/local
rm -rf /tmp/gopkg/
export PATH=$PATH:$GOROOT/bin
export GOPATH=$HOME/gopath
export PATH=$PATH:$GOPATH/bin
echo "export PATH=$PATH:$GOROOT/bin" >> $HOME/.bashrc
echo "export GOPATH=$HOME/gopath" >> $HOME/.bashrc
echo "export PATH=$PATH:$GOPATH/bin" >> $HOME/.bashrc
source $HOME/.bashrc 


mkdir -p $HOME/gopath
sudo mount /dev/vdb $HOME/gopath
mkdir -p $HOME/gopath/src

cd $HOME/gopath/src
git clone git@git.hyperchain.cn/hyperchain/hyperchain.git
cd hyperchain
git checkout develop
git pull origin develop
echo "all code were upgreded"
