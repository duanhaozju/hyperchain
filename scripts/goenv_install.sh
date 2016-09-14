#!/bin/bash
# this script is used for auto install go env on the aliyun server
# author:ChenQuan
# date: 2016-09-05
# usage: bash goenv_install.sh


# Stop on first error
set -e
set -x
# Update the entire system to the latest releases
apt-get update -qq
apt-get dist-upgrade -qqy

# auto install the golang env
apt-get install  --yes build-essential git

mkdir -p /usr/local/go

# Set Go environment variables needed by other scripts
export GOPATH="$HOME/gopath"
mkdir -p $GOPATH

MACHINE=`uname -m`
if [ x$MACHINE = xx86_64 ]
then
   export GOROOT="/usr/local/go"

   #ARCH=`uname -m | sed 's|i686|386|' | sed 's|x86_64|amd64|'`
   ARCH=amd64
   GO_VER=1.7.1

   cd /tmp
   wget --quiet --no-check-certificate https://storage.googleapis.com/golang/go$GO_VER.linux-amd64.tar.gz
   tar -xvf go$GO_VER.linux-${ARCH}.tar.gz
   mv go $GOROOT
   chmod 775 $GOROOT
   rm go$GO_VER.linux-${ARCH}.tar.gz
else
    echo "NOT X86_64"
    exit 1
fi

echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.profile
echo 'export GOPATH=$HOME/gopath' >> $HOME/.profile
echo 'export PATH=$PATH:$GOPATH/bin' >> $HOME/.profile
source $HOME/.profile
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/gopath
export PATH=$PATH:$GOPATH/bin
go env
