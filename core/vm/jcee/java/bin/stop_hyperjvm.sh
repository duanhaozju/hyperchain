#!/usr/bin/env bash

jps -l | grep cn.hyperchain.jcee.JceeServer | awk '{ print $1 }' | xargs kill -9