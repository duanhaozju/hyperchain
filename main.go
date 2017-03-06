//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"errors"
	"fmt"
	"github.com/mkideal/cli"
	"github.com/terasum/viper"
	"hyperchain/accounts"
	"hyperchain/admittance"
	"hyperchain/api/jsonrpc/core"
	"hyperchain/common"
	"hyperchain/consensus/controller"
	"hyperchain/core"
	"hyperchain/core/executor"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/manager"
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
	//NodeID     int    `cli:"o,id" usage:"node ID" dft:"1"`
	ConfigPath string `cli:"c,conf" usage:"config file path" dft:"./config/global.yaml"`
	//GRPCPort   int    `cli:"l,rpcport" usage:"inner grpc connect port" dft:"8001"`
	//HTTPPort   int    `cli:"t,httpport" useage:"jsonrpc open port" dft:"8081"`
	//RESTPort   int    `cli:"f,restport" useage:"restful api port" dft:"9000"`
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

	str1 := GetHyperchainVersion()
	str2, _ := GetOperationSystem()
	fmt.Println(str1)
	fmt.Println(str2)
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
	//conf.Set(common.HYPERCHAIN_ID, argv.NodeID)
	//conf.Set(common.HTTP_PORT, argv.HTTPPort)
	//conf.Set(common.REST_PORT, argv.RESTPort)
	//conf.Set(common.GRPC_PORT, argv.GRPCPort)
	// read the global peers path
	peerConfigPath := conf.GetString("global.configs.peers")
	peerViper := viper.New()
	peerViper.SetConfigFile(peerConfigPath)
	err := peerViper.ReadInConfig()
	if err != nil {
		panic("read in the peer config failed")
	}
	nodeID := peerViper.GetInt("self.node_id")
	grpcPort := peerViper.GetInt("self.grpc_port")
	jsonrpcPort := peerViper.GetInt("self.jsonrpc_port")
	restfulPort := peerViper.GetInt("self.restful_port")

	conf.Set(common.C_NODE_ID, nodeID)
	conf.Set(common.C_HTTP_PORT, jsonrpcPort)
	conf.Set(common.C_REST_PORT, restfulPort)
	conf.Set(common.C_GRPC_PORT, grpcPort)
	conf.Set(common.C_PEER_CONFIG_PATH, peerConfigPath)
	conf.Set(common.C_GLOBAL_CONFIG_PATH, argv.ConfigPath)

	return conf
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		conf := initConf(argv)

		//TODO:remove this config later
		config := newconfigsImpl(conf)
		//LOG
		common.InitLog(conf)
		//DB

		conf.MergeConfig(config.getDbConfig())//todo:refactor it

		core.InitDB(conf, config.getDbConfig(),conf.GetInt(common.C_NODE_ID))

		err, expiredTime := checkLicense(config.getLicense())
		if err != nil {
			return err
		}

		eventMux := new(event.TypeMux)

		/**
		 *传入true则开启所有验证，false则为取消ca以及签名的所有验证
		 */
		globalConfig := viper.New()
		globalConfig.SetConfigFile(conf.GetString(common.C_GLOBAL_CONFIG_PATH))
		err = globalConfig.ReadInConfig()
		if err != nil {
			panic(err)
		}
		cm, cmerr := admittance.GetCaManager(globalConfig)
		if cmerr != nil {
			panic("cannot initliazied the camanager")
		}

		//init peer manager to start grpc server and client
		//grpcPeerMgr := p2p.NewGrpcManager(config.getPeerConfigPath())
		grpcPeerMgr := p2p.NewGrpcManager(conf)

		//init genesis
		core.CreateInitBlock(conf)

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
		executor := executor.NewBlockExecutor(cs, conf, kec256Hash, encryption, eventMux)
		if executor == nil {
			return errors.New("Initialize BlockPool failed")
		}
		executor.Initialize()
		//init manager
		exist := make(chan bool)
		syncReplicaInterval, _ := config.getSyncReplicaInterval()
		syncReplicaEnable := config.getSyncReplicaEnable()
		pm := manager.New(eventMux,
			executor,
			grpcPeerMgr,
			cs,
			am,
			kec256Hash,
			syncReplicaInterval,
			syncReplicaEnable,
			exist,
			expiredTime, cm)
		go jsonrpc.Start(config.getHTTPPort(), config.getRESTPort(), config.getLogDumpFileDir(), eventMux, pm, cm, conf)

		//go func() {
		//	log.Println(http.ListenAndServe("localhost:6064", nil))
		//}()

		<-exist
		return nil
	})
}
