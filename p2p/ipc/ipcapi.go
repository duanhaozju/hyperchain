package ipc

import "fmt"

type IPCAPI struct {
	apis map[string]func(args []string)error
}

func NewIPCApi()*IPCAPI{
	return &IPCAPI{
		apis:make(map[string]func(args []string)error),
	}
}

func (api *IPCAPI)Register(cmd string,callback func(args[]string)error){
	if _,ok :=api.apis[cmd];!ok{
		api.apis[cmd] = callback
	}else{
		fmt.Printf("omit register of cmd: %s \n",cmd)
	}
}

func (api *IPCAPI)Process(cmd string,args []string){
	if f,ok:=api.apis[cmd];ok{
		err := f(args)
		if err != nil{
			fmt.Println("Failed: %s",err.Error())
		}
	}else{
		fmt.Println("Command \"%s\" not found!",cmd)
	}
}
