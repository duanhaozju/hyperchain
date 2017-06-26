package main

import (
	"hyperchain/p2p/ipc"
)

func main() {
	ipcserver := ipc.NEWIPCServer("./hpc.ipc")
	err := ipcserver.Start()
	if err != nil{
		panic(err)
	}
	ipc.IPCShell()
}