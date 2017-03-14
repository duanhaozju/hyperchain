//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"fmt"
	"github.com/mkideal/cli"
	"hyperchain/common"
	"hyperchain/namespace"
	//"hyperchain/api/jsonrpc/core"
	"hyperchain/api/jsonrpc/core"
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



		gns := nsMgr.GetNamespaceByName("global").(*namespace.NamespaceImpl)
		jsonrpc.Start(gns.CaMgr, gns.Conf, nsMgr)


		<-stopHyperchain
		return nil
	})
}

func Stop() {
	stopHyperchain <- true
}
