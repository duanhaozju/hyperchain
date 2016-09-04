#!/usr/bin/env bash
git pull origin master
govendor build
./hyperchain -o $1 -l 8081