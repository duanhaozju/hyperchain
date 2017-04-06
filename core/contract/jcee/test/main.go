//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"github.com/op/go-logging"
	"hyperchain/core/contract/jcee/go"
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
	x := make(chan bool)
	<- x
}
