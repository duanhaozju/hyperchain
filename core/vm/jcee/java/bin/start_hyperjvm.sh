#!/usr/bin/env bash

# TODO: Add memory control

java -cp $(for i in ../libs/*.jar ; do echo -n $i: ; done).  cn.hyperchain.jcee.JceeServer $1 $2