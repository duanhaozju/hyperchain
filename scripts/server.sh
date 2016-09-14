#!/usr/bin/env bash
# this script is used for auto test on the server
# author:ChenQuan
# date: 2016-09-05
# usage: bash server.sh 1

set -e
cd ../
rm -rf /tmp/hyperchain/*
git pull origin develop
govendor build
./hyperchain -o $1 -l 8081