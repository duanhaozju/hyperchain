package ipc

import (
	"os"
	"net"
	"path/filepath"
	"time"
	"fmt"
	"net/http"
	"net/rpc"
)
var (
	tcpKeepAliveInterval = 30 * time.Second
)

type IPCServer	struct {
	endpoint string
}

func NEWIPCServer(endpoint string) *IPCServer{
	return &IPCServer{
		endpoint:endpoint,
	}
}


func (server *IPCServer)Start(rcrecv interface{})error{
	var (
		listener net.Listener
		err error
	)

	rpc.Register(rcrecv)
	rpc.HandleHTTP()

	if listener,err = server.listener();err != nil {
		fmt.Println("some error occured",err.Error())
		return err
	}
	if err != nil{
		fmt.Println(err)
		return err
	}

	go func(){
		err =  http.Serve(listener,nil)
		if err != nil{
			panic(err)
		}
	}()
	return nil
}


func (server *IPCServer)listener()(net.Listener, error){
	if err := os.MkdirAll(filepath.Dir(server.endpoint),0751);err != nil{
		return nil,err
	}

	os.Remove(server.endpoint)
	fmt.Println("start unix ipc server, ",server.endpoint)
	l,err := net.Listen("unix",server.endpoint)
	if err != nil{
		return nil,err
	}
	os.Chmod(server.endpoint,0600)
	return l,nil
}

