#!/bin/bash
# this script is used for auto install go env on the aliyun server
# author:ChenQuan
# date: 2016-09-05
# usage: bash goenv_install.sh


# Stop on first error
set -e
#set -x

# Update the entire system to the latest releases
sudo apt-get update -qq
sudo apt-get dist-upgrade -qqy
sudo apt-get -y install git build-essential wget

GOROOT=/usr/local/go
GOINSTALLPATH=/usr/local


cd /tmp
mkdir gopkg
wget --no-check-certificate https://api.google.com/golang/go1.7.1-amd64.tar.gz
sudo tar zxf go1.7.1-amd64.tar.gz -c $GOINSTALLPATH
chmod 775 $GOROOT
rm -rf /tmp/gopkg/

# mount the ssd
mkdir -p $HOME/gopath
sudo mount /dev/vdb $HOME/gopath
mkdir -p $HOME/gopath/src

# setup the env
export PATH=$PATH:$GOROOT/bin
export GOPATH=$HOME/gopath
export PATH=$PATH:$GOPATH/bin
echo "export PATH=$PATH:$GOROOT/bin" >> $HOME/.bashrc
echo "export GOPATH=$HOME/gopath" >> $HOME/.bashrc
echo "export PATH=$PATH:$GOPATH/bin" >> $HOME/.bashrc
source $HOME/.bashrc
go env

# deploy the code
cd $HOME/gopath/src
git clone git@git.hyperchain.cn/hyperchain/hyperchain.git
cd hyperchain
git checkout develop
git pull origin develop
echo "all code were deployed"

# install the govendor
go get -u github.com/kardianos/govendor
govendor sync

echo "done"