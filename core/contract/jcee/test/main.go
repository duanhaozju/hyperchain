//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"github.com/op/go-logging"
	"hyperchain/core/contract/jcee/go"
	pb "hyperchain/core/contract/jcee/protos"
)

var logger *logging.Logger
const (
	address     = "localhost:50051"
	defaultName = "world"
)

func init() {
	logger = logging.MustGetLogger("test")
}

func main() {
	exe := jcee.NewContractExecutor()
	exe.Start()

	request := &pb.Request{
		Cid:"msc001",
		Method:"invoke",
		Args:[][]byte{[]byte("test"), []byte("wangxiaoyi")},
	}

	response, err := exe.Execute(request)

	if err!= nil {
		logger.Error(err)
	}
	logger.Info(response)

	x := make(chan bool)
	<- x
}
