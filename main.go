//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"github.com/mkideal/cli"
	"hyperchain/common"
	"hyperchain/namespace"
)

type argT struct {
	cli.Helper
	ConfigPath string `cli:"c,conf" usage:"config file path" dft:"./global.yaml"`
}

func init()  {
	//globalConfig := common.NewConfig("./global.yaml")
	//common.InitHyperLogger(globalConfig)
}
var stopHyperchain chan bool
var nsMgr namespace.NamespaceManager

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		stopHyperchain = make(chan bool)

		globalConfig := common.NewConfig(argv.ConfigPath)
		common.InitHyperLogger(globalConfig)

		nsMgr = namespace.GetNamespaceManager(globalConfig)
		nsMgr.Start()

		<-stopHyperchain
		return nil
	})
}

func Stop() {
	nsMgr.Stop()
	stopHyperchain <- true
}
