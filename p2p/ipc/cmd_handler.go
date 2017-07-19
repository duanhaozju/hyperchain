package ipc

import (
	"github.com/abiosoft/ishell"
	"strings"
	"fmt"
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
		logger.Errorf("can't create IPC pipe, error: %s\n", err)
		return
	}
	args := Args{
		Cmd:c.Cmd.Name,
		Argv:c.Args,
	}
	ret := new(Ret)
	err = client.Call("RemoteCall.Call",args,ret)
	if err != nil{
		logger.Errorf("error occures: %s",err.Error())
		return
	}
	fmt.Println(strings.Join(ret.Rets," "))
	client.Close()
}
