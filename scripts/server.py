#! /usr/bin/python
# -*- coding=utf8 -*-
#
# this script is used to auto deploy the compiled binary code.
# and auto run the predefined command.
# Author: Chen Quan
# Update Date: 2016-10-16
# Features:
# 1. auto add the ssh key into the primary sever
# 2. auto add the primary's ssh key charo the non-primary server
# 3. accelerate the distributes speed
# 4. auto read the server list file and auto run the suit command
import subprocess

def add_ssh_key_into_primary:
    print "add your local ssh public key into primary node"
    print "将你的本地ssh公钥添加到primary中"
    subpricess.call(['./sub_scripts/add_ssh_to_primary.sh'])

def add_ssh_key_form_primary_to_others:
    print "primary add its ssh key into others nodes"
    print "primary 将它的ssh 公钥加入到其它节点中"

def build:
    print "编译并生成配置文件"

def upload_binary_to_primary:
    print "将本地生成的文件上传到primary中"

def distribute_the_binary:
    print "由primary 分发二进制文件和配置文件"


def clean:
    print "清除本地生成的文件"

def read_server_list:
    print "读取server_list"

def read_inner_server_lis:
    print "读取innerserverlist,并生成相应的文件供primary读取操作"

def auto_run:
    print "自动运行相应命令，启动全节点"

def interaction_mode_hint:
    print "请输入相应编号"
    print "\t1. 全节点压力测试"
    print "\t2. 主节点压力测试"
    print "\t3. 从节点压力测试"
    print "\t4. 数据查询"

def interaction_mode:
    print "进入交互测试模式"
    interaction_mode_hint()

