
#!/bin/bash
#if [ ! -f "/usr/bin/expect" ];then
#  echo "hasn't install expect,please install expect mannualy: 'apt-get install expect'"
##  sudo apt-get install expect
#exit 1
#fi

if [ ! -d /home/satoshi/gopath ];then
   mkdir -p /home/satoshi/gopath
fi
sudo mount /dev/vdb /home/satoshi/gopath/
ls /home/satoshi/gopath/
cd /home/satoshi/gopath/src/hyperchain
git checkout develop
#expect <<EOF
#        set timeout 60
#        spawn git pull origin develop
#        expect {
#          "yes/no" {send "yes\r";exp_continue }
#          eof
#        }
#EOF
echo "already update"
