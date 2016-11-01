// init ProtcolManager
// author: Lizhong kuang
// date: 2016-08-23
// last modified:2016-08-29
package main

import (
	"errors"
	"github.com/mkideal/cli"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/consensus/controller"
	"hyperchain/core"
	"hyperchain/core/blockpool"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/jsonrpc"
	"hyperchain/manager"
	"hyperchain/membersrvc"
	"hyperchain/p2p"
	"hyperchain/p2p/transport"
	"strconv"
	"strings"
	"time"
)

type argT struct {
	cli.Helper
	NodeID     int    `cli:"o,id" usage:"node ID" dft:"1"`
	ConfigPath string `cli:"c,conf" usage:"配置文件所在路径" dft:"./config/global.yaml"`
	GRPCPort   int    `cli:"l,rpcport" usage:"远程连接端口" dft:"8001"`
	HTTPPort   int    `cli:"t,httpport" useage:"jsonrpc开放端口" dft:"8081"`
}

func checkLicense(licensePath string) error {
	dateChecker := func(start, now, expire time.Time) bool {
		return now.After(start) && now.Before(expire)
	}
	privateKey := string("TnrEP|N.*lAgy<Q&@lBPd@J/")
	identificationSuffix := string("Copyright 2016 The Hyperchain. All rights reserved.")
	encryptedInfo := []byte{206, 21, 208, 92, 82, 46, 155, 119, 43, 186, 40, 10, 39, 7, 36, 157, 147, 211, 14, 232, 175, 32, 72, 130, 237, 209, 85, 160, 30, 121, 121, 153, 124, 59, 41, 0, 189, 155, 236, 254, 129, 166, 164, 51, 98, 191, 196, 46, 136, 95, 191, 89, 139, 130, 175, 240, 0, 7, 30, 95, 50, 115, 140, 73, 226, 73, 160, 217, 189, 170, 221, 245}
	identification, err := transport.TripleDesDecrypt(encryptedInfo, []byte(privateKey))
	if err != nil {
		return errors.New("Invalid License")
	}
	plainText := string(identification)
	suffix := plainText[len(plainText)-len(identificationSuffix):]
	if strings.Compare(suffix, identificationSuffix) != 0 {
		return errors.New("Invalid Identification")
	}
	timestamp, err := strconv.ParseInt(plainText[:len(plainText)-len(identificationSuffix)], 16, 64)
	if err != nil {
		return errors.New("Invalid License Timestamp")
	}
	startTime := time.Unix(timestamp, 0)
	expiredTime := startTime.AddDate(3, 0, 0)
	currentTime := time.Now()
	if validation := dateChecker(startTime, currentTime, expiredTime); !validation {
		return errors.New("License Expired")
	}
	return nil
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		config := newconfigsImpl(argv.ConfigPath, argv.NodeID, argv.GRPCPort, argv.HTTPPort)
		if err := checkLicense(config.getLicense()); err != nil {
			return err
		}

		membersrvc.Start(config.getMemberSRVCConfigPath(), config.getNodeID())

		//init log
		common.InitLog(config.getLogLevel(), config.getLogDumpFileDir(), config.getGRPCPort(), config.getLogDumpFileFlag())

		eventMux := new(event.TypeMux)

		//init peer manager to start grpc server and client
		grpcPeerMgr := p2p.NewGrpcManager(config.getPeerConfigPath(), config.getNodeID())

		//init db
		core.InitDB(config.getDatabaseDir(), config.getGRPCPort())

		//init genesis
		core.CreateInitBlock(config.getGenesisConfigPath())

		//init pbft consensus
		cs := controller.NewConsenter(uint64(config.getNodeID()), eventMux, config.getPBFTConfigPath())

		//init encryption object

		encryption := crypto.NewEcdsaEncrypto("ecdsa")
		encryption.GenerateNodeKey(strconv.Itoa(config.getNodeID()), config.getKeyNodeDir())
		//
		am := accounts.NewAccountManager(config.getKeystoreDir(), encryption)
		am.UnlockAllAccount(config.getKeystoreDir())

		//init hash object
		kec256Hash := crypto.NewKeccak256Hash("keccak256")

		//init block pool to save block
		blockPool := blockpool.NewBlockPool(eventMux, cs)

		//init manager
		exist := make(chan bool)
		syncReplicaInterval, _ := config.getSyncReplicaInterval()
		syncReplicaEnable := config.getSyncReplicaEnable()
		pm := manager.New(eventMux, blockPool, grpcPeerMgr, cs, am, kec256Hash,
			config.getNodeID(), syncReplicaInterval, syncReplicaEnable)

		go jsonrpc.Start(config.getHTTPPort(), eventMux, pm)

		<-exist

		return nil
	})
}
