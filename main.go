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
	"hyperchain/core/blockpool"
	"hyperchain/event"
	"hyperchain/manager"
	"hyperchain/p2p"
	"hyperchain/p2p/transport"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const HyperchainVersion = "Hyperchain Version:\nRelease1.2\n"

func GetOperationSystem() (string, error) {
	f, err := exec.Command("lsb_release", "-a").Output()
	if err != nil {
		return "", err
	}
	return string(f), nil
}

func GetHyperchainVersion() string {
	return HyperchainVersion
}

type argT struct {
	cli.Helper
	ConfigPath string `cli:"c,conf" usage:"config file path" dft:"./config/global.yaml"`
	//ConfigPath string `cli:"c,conf" usage:"config file path" dft:"./configuration/global.yaml"`
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

func initGloableConfig(argv *argT) *common.Config {
	//default path: ./configuration/global.yaml
	conf := common.NewConfig(argv.ConfigPath)
	return conf
}

func initConf(argv *argT) *common.Config {
	conf := common.NewConfig(argv.ConfigPath)
	// init peer configurations
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

		err, expiredTime := checkLicense(conf.GetString(common.LICENSE))
		if err != nil {
			return err
		}

		common.InitLog(conf)
		core.InitDB(conf)

		cm, cmerr := admittance.GetCaManager(conf)
		if cmerr != nil {
			panic("cannot initliazied the camanager")
		}

		//init peer manager to start grpc server and client
		grpcPeerMgr := p2p.NewGrpcManager(conf)

		//init genesis
		core.CreateInitBlock(conf)

		eventMux := new(event.TypeMux)
		//init pbft consensus
		cs := controller.NewConsenter(eventMux, conf)

		am := accounts.NewAccountManager(conf)
		am.UnlockAllAccount(conf.GetString(common.KEY_STORE_DIR))

		//init block pool to save block
		blockPool := blockpool.NewBlockPool(cs, conf, eventMux)
		if blockPool == nil {
			return errors.New("Initialize BlockPool failed")
		}
		blockPool.Initialize()
		exist := make(chan bool)
		//init manager
		pm := manager.New(eventMux, blockPool, grpcPeerMgr, cs, am, exist, expiredTime, cm, conf)

		go jsonrpc.Start(eventMux, pm, cm, conf)

		<-exist
		return nil
	})
}
