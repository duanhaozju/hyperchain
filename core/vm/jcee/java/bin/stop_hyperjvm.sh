#!/usr/bin/env bash

# stop hyperjvm
if [ $# -eq 2 ]; then
    echo "kill hyperjvm with port $1 $2"
    jps -ml | grep "cn.hyperchain.jcee.JceeServer $1 $2" | awk '{ print $1 }' | xargs kill -9
else
    echo "kill all hyperjvm process"
    jps -l | grep cn.hyperchain.jcee.JceeServer | awk '{ print $1 }' | xargs kill -9
fi