#!/usr/bin/env bash

java -cp $(for i in ../libs/*.jar ; do echo -n $i: ; done).  cn.hyperchain.jcee.LocalJceeServer