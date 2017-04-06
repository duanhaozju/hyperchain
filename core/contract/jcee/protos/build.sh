#!/usr/bin/env bash

protoc --plugin=protoc-gen-grpc-java=/Users/wangxiaoyi/codes/grpc-java/compiler/build/exe/java_plugin/protoc-gen-grpc-java ./contract.proto --grpc-java_out=../java/src/main/java/ --java_out=../java/src/main/java/
protoc ./contract.proto  --go_out=plugins=grpc:./