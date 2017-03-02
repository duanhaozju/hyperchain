#!/bin/bash - 
#===============================================================================
#
#          FILE: new.sh
# 
#         USAGE: ./new.sh 
# 
#   DESCRIPTION: 
# 
#       OPTIONS: ---
#  REQUIREMENTS: ---
#          BUGS: ---
#         NOTES: ---
#        AUTHOR: YOUR NAME (), 
#  ORGANIZATION: 
#       CREATED: 2017/02/28 11:07
#      REVISION:  ---
#===============================================================================

set -o nounset                              # Treat unset variables as an
DUMP_PATH="./"
MAXPEERNUM=4
# 运行命令

f_all_in_one_cmd(){
    cd $DUMP_PATH/node${1} && ./hyperchain  &
}

f_x_in_linux_cmd(){
    gnome-terminal -x bash -c "cd ${DUMP_PATH}/node${1} && ./hyperchain"
}

f_x_in_mac_cmd(){
    osascript -e 'tell app "Terminal" to do script "cd '$DUMP_PATH/node${1}' && ./hyperchain "'
}


# 组织运行命令函数
f_run_process(){
    for((j=1;j<=$MAXPEERNUM;j++))
    do
        case "$_SYSTYPE" in
          MAC*)
             f_x_in_mac_cmd $j
          ;;
          LINUX*)
             f_x_in_linux_cmd $j
          ;;
          ALLINONE*)
            f_all_in_one_cmd $j
          ;;
          *)
            echo "unknow OS TYPE $OSTYPE"
            exit -1
          ;;
        esac
    done
}

f_sleep(){
    sleep ${1}s
}

