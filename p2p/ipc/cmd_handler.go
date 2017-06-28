package ipc

import (
	"fmt"
	"github.com/abiosoft/ishell"
	"strings"
)

type CMDHandler	struct {
	ipcEndpoint string
	callbacks map[string]func(args []string)error
}



func NewIPCHandler(endpoint string)*CMDHandler{
	return &CMDHandler{
		ipcEndpoint:endpoint,
	}
}

func(handler *CMDHandler)handle(c *ishell.Context){
	client,err := newIPCConnection(handler.ipcEndpoint)
	if err != nil {
		fmt.Printf("can't create IPC pipe, error: %s\n", err)
		return
	}
	args := Args{
		Cmd:c.Cmd.Name,
		Argv:c.Args,
	}
	ret := new(Ret)
	err = client.Call("RemoteCall.Call",args,ret)
	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Println(strings.Join(ret.Rets," "))
	client.Close()
}
