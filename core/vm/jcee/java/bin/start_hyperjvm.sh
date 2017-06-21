#!/usr/bin/env bash

# TODO: Add memory control

if [ "$#" -ne 2 ]; then
    echo "please specify localPort and the ledger port"
fi

java -cp $(for i in hyperjvm/libs/*.jar ; do echo -n $i: ; done).  cn.hyperchain.jcee.JceeServer $1 $2 &

