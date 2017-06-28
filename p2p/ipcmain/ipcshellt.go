package main

import (
	"hyperchain/p2p/ipc"
	"hyperchain/p2p/utils"
	"hyperchain/p2p/network"
	"github.com/spf13/viper"
	"time"
)

func main() {
	vip := viper.New()
	vip.Set("global.p2p.hosts",utils.GetProjectPath()+"/p2p/test/hosts.yaml")
	vip.Set("global.p2p.port",50015)
	vip.Set("global.p2p.retrytime","30s")

	hypernet,err := network.NewHyperNet(vip)
	if err != nil{
		panic(err)
	}
	err = hypernet.InitServer()

	if err !=nil{
		panic(err)
	}
	<- time.After(3*time.Second)
	err = hypernet.InitClients()

	if err != nil{
		panic(err)
	}
	rc := ipc.NewRemoteCall()
	ipc.RegisterFunc(rc,"network",hypernet.Command)
	ipcserver := ipc.NEWIPCServer("./hpc.ipc")
	err = ipcserver.Start(rc)
	if err != nil{
		panic(err)
	}
	ipc.IPCShell("./hpc.ipc")
}