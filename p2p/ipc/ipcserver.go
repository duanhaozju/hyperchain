package ipc

import (
	"os"
	"net"
	"path/filepath"
	"time"
	"fmt"
	"encoding/json"
)
var (
	tcpKeepAliveInterval = 30 * time.Second
)

type IPCServer	struct {
	api *IPCAPI
	endpoint string
}

func NEWIPCServer(endpoint string) *IPCServer{
	return &IPCServer{
		endpoint:endpoint,
		api:NewIPCApi(),
	}
}


func (server *IPCServer)Start()error{
	var (
		listener net.Listener
		err error
	)
	if listener,err = server.Listener();err != nil {
		return err
	}
	go func() {
		for{
			conn,err := listener.Accept()
			if err !=nil{
				fmt.Errorf("error occured: %s",err.Error())
				continue
			}
			output := make([]byte,500);
			conn.Read(output)
			result := make([]byte,0)
			for _,b :=range output{
				if b != byte(0x0){
					result = append(result,b)
				}else{
					break
				}
			}
			c := new(IPCCmd)
			err = json.Unmarshal(result, c)
			if err != nil{
				fmt.Errorf("error occured: %s",err.Error())
			}
			fmt.Println("got a new command,",result)
			server.api.Process(c.CMD, c.Args)
		}
	}()
	return nil
}


func (server *IPCServer)Listener()(net.Listener, error){
	if err := os.MkdirAll(filepath.Dir(server.endpoint),0751);err != nil{
		return nil,err
	}

	os.Remove(server.endpoint)
	l,err := net.Listen("unix",server.endpoint)
	if err != nil{
		return nil,err
	}
	os.Chmod(server.endpoint,0600)
	return l,nil
}

