#!/usr/bin/env bash

#java -cp $(for i in ../libs/*.jar ; do echo -n $i: ; done).  $1 $2 $3
if [ "$#" -ne 2 ]; then
    echo "invalid params length"
    echo "usage: contract_compile.sh contractDir contractClassDir"
fi
find $1 -name '*.java' | xargs javac -cp hyperjvm/sdk/hyperjvm-sdk-1.0.jar -d $2
#javac -cp ../sdk/hyperjvm-sdk-1.0.jar -d