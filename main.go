//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"fmt"
	"github.com/mkideal/cli"
	"hyperchain/common"
	"hyperchain/namespace"
)

type argT struct {
	cli.Helper
	ConfigPath string `cli:"c,conf" usage:"config file path" dft:"./global.yaml"`
}

func initGloableConfig(argv *argT) *common.Config {
	conf := common.NewConfig(argv.ConfigPath)
	return conf
}

var stopHyperchain chan bool

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		stopHyperchain = make(chan bool)

		globalConfig := initGloableConfig(argv)
		common.InitLog(globalConfig)

		nsMgr := namespace.GetNamespaceManager(globalConfig)

		namespaces := nsMgr.List()
		for _, n := range namespaces {
			fmt.Printf("namespace: %s\n", n)
		}
		nsMgr.Start()

		<-stopHyperchain
		return nil
	})
}

func Stop() {
	stopHyperchain <- true
}
