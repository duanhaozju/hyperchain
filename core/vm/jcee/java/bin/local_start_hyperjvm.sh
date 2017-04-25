#!/usr/bin/env bash

# build and start the hyperjvm
#./build.sh

#cd ./hyperjvm/bin/ && ./stop_hyperjvm.sh

ledgerPorts=(50051 50052 50053 50054)
localPorts=(50081 50082 50083 50084)

for i in 0 1 2 3
do
    ./start_hyperjvm.sh ${localPorts[i]} ${ledgerPorts[i]} &
done
