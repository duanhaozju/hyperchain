#!/usr/bin/env bash

# build hyperjvm scripts

if [ -d hyperjvm ] ; then
    echo "hyperjvm dir existed, remove it"
    rm -rf hyperjvm
fi

echo "1. make dir hyperjvm"
mkdir hyperjvm
mkdir hyperjvm/libs
mkdir hyperjvm/bin
mkdir hyperjvm/config
mkdir hyperjvm/contracts
mkdir hyperjvm/meta
touch  hyperjvm/meta/meta.yaml

echo "2. build the hyperjvm"
mvn clean package -Dmaven.test.skip=true
cp target/lib/*  hyperjvm/libs/

echo "3. build the hyperjvm client"
#./build_sdk.sh

echo "4. clean target package"
rm -rf target

echo "5. copy the control scripts and configurations"
cp -rf bin/* hyperjvm/bin/

cp -rf ./src/main/resources/* hyperjvm/config/


echo "6. finish build hyperjvm"
