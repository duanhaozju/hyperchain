//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
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
	"hyperchain/api/jsonrpc/core"
	"hyperchain/manager"
	"hyperchain/membersrvc"
	"hyperchain/p2p"
	"hyperchain/p2p/transport"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type argT struct {
	cli.Helper
	NodeID     int    `cli:"o,id" usage:"node ID" dft:"1"`
	ConfigPath string `cli:"c,conf" usage:"config file path" dft:"./config/global.yaml"`
	GRPCPort   int    `cli:"l,rpcport" usage:"inner grpc connect port" dft:"8001"`
	HTTPPort   int    `cli:"t,httpport" useage:"jsonrpc open port" dft:"8081"`
	RESTPort   int	  `cli:"f,restport" useage:"restful api port" dft:"9000"`
	//IsReconnect bool  `cli:"e,isReconnect" usage:"是否重新链接" dft:"false"`
}



func checkLicense(licensePath string) (err error, expiredTime time.Time) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Invalid License Cause a Panic")
		}
	}()
	dateChecker := func(now, expire time.Time) bool {
		return now.Before(expire)
	}
	privateKey := string("TnrEP|N.*lAgy<Q&@lBPd@J/")
	identificationSuffix := string("Hyperchain")

	license, err := ioutil.ReadFile(licensePath)
	if err != nil {
		err = errors.New("No License Found")
		return
	}
	pattern, _ := regexp.Compile("Identification: (.*)")
	identification := pattern.FindString(string(license))[16:]

	ctx, err := transport.TripleDesDecrypt(common.Hex2Bytes(identification), []byte(privateKey))
	if err != nil {
		err = errors.New("Invalid License")
		return
	}
	plainText := string(ctx)
	suffix := plainText[len(plainText)-len(identificationSuffix):]
	if strings.Compare(suffix, identificationSuffix) != 0 {
		err = errors.New("Invalid Identification")
		return
	}
	timestamp, err := strconv.ParseInt(plainText[:len(plainText)-len(identificationSuffix)], 10, 64)
	if err != nil {
		err = errors.New("Invalid License Timestamp")
		return
	}
	expiredTime = time.Unix(timestamp, 0)
	currentTime := time.Now()
	if validation := dateChecker(currentTime, expiredTime); !validation {
		err = errors.New("License Expired")
		return
	}
	return
}

func initConf(argv *argT) *common.Config {
	conf := common.NewConfig(argv.ConfigPath)
	conf.Set(common.HYPERCHAIN_ID, argv.NodeID)
	conf.Set(common.HTTP_PORT, argv.HTTPPort)
	conf.Set(common.REST_PORT, argv.RESTPort)
	conf.Set(common.GRPC_PORT, argv.GRPCPort)
	return conf
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		//TODO:remove this config later
		config := newconfigsImpl(argv.ConfigPath, argv.NodeID, argv.GRPCPort, argv.HTTPPort, argv.RESTPort)

		conf := initConf(argv)
		common.InitLog(conf)

		err, expiredTime := checkLicense(config.getLicense())
		if err != nil {
			return err
		}

		membersrvc.Start(config.getMemberSRVCConfigPath(), config.getNodeID())

		eventMux := new(event.TypeMux)

		//init memversrvc CAManager
		// rca.ca 应该改为 eca.ca
		cm,cmerr := membersrvc.NewCAManager("./config/cert/eca.ca","./config/cert/ecert.cert","./config/cert/rcert.cert","./config/cert/rca.ca","./config/cert/ecert.priv","./config/cert/server/tca.ca")
		if cmerr != nil{
			panic("cannot initliazied the camanager")
		}

		//init peer manager to start grpc server and client
		//grpcPeerMgr := p2p.NewGrpcManager(config.getPeerConfigPath())
		grpcPeerMgr := p2p.NewGrpcManager(conf)


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
		blockPoolConf := blockpool.BlockPoolConf{
			BlockVersion: config.getBlockVersion(),
			TransactionVersion: config.getTransactionVersion(),
		}
		blockPool := blockpool.NewBlockPool(eventMux, cs, blockPoolConf)
		if blockPool == nil {
			return errors.New("Initialize BlockPool failed")
		}

		//init manager
		exist := make(chan bool)
		syncReplicaInterval, _ := config.getSyncReplicaInterval()
		syncReplicaEnable := config.getSyncReplicaEnable()
		pm := manager.New(eventMux,
				blockPool,
				grpcPeerMgr,
				cs,
				am,
				kec256Hash,
				syncReplicaInterval,
				syncReplicaEnable,
				exist,
				expiredTime,cm)
		rateLimitCfg := config.getRateLimitConfig()
		go jsonrpc.Start(config.getHTTPPort(), config.getRESTPort(),config.getLogDumpFileDir(),eventMux, pm, rateLimitCfg,cm, config.getPaillerPublickey())

		//go func() {
		//	log.Println(http.ListenAndServe("localhost:6064", nil))
		//}()

		<-exist
		return nil
	})
}
