#!/bin/bash
if [ ! -d /home/satoshi/gopath ];then
    mkdir -p /home/satoshi/gopath
fi
sudo mount /dev/vdb /home/satoshi/gopath
cd /home/satoshi/gopath/src/hyperchain
git checkout develop
git pull origin develop
echo "already update"

