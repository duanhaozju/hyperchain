#!/usr/bin/env bash

# TODO: Add memory control

which java
if [ "$?" -ne 0 ]; then
    echo "no java found ,please install jdk first, version >= 1.8"
    exit 1
fi

JAVA_VERSION=`java -version 2>&1 |awk 'NR==1{ gsub(/"/,""); print $3 }'`
MIN_JAVA_VERSION="1.8"
if [[ "$JAVA_VERSION" < "$MIN_JAVA_VERSION" ]]; then
    echo "machine java version is $JAVA_VERSION which is older than min required java version $MIN_JAVA_VERSION"
    echo "please install jdk1.8 instead"
    exit 1
fi

if [ "$#" -ne 2 ]; then
    echo "please specify localPort and the ledger port"
fi

java -cp $(for i in hyperjvm/libs/*.jar ; do echo -n $i: ; done).  cn.hyperchain.jcee.JceeServer $1 $2 &

