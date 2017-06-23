#!/usr/bin/env bash

#java -cp $(for i in ../libs/*.jar ; do echo -n $i: ; done).  $1 $2 $3
if [ "$#" -ne 3 ]; then
    echo "invalid params length"
    echo "usage: contract_compile.sh jarDir contractDir contractClassDir"
fi
find $2 -name '*.java' | xargs javac -cp $1/hyperjvm-sdk-1.0.jar -d $3