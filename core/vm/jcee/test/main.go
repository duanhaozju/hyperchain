//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	//"fmt"
	"github.com/op/go-logging"
	//"google.golang.org/grpc"
	//"hyperchain/core/vm/jcee/go/client"
	//lg "hyperchain/core/vm/jcee/go/ledger"
	//pb "hyperchain/core/vm/jcee/protos"
	//"net"
	//"strconv"
	//"time"
	"hyperchain/common"
	"hyperchain/core/state"
	"hyperchain/core/vm"
	"hyperchain/hyperdb/mdb"
	"os"
	"path"
	//"github.com/op/go-logging"
	"fmt"
)

var logger *logging.Logger

const (
	address     = "localhost:50051"
	defaultName = "world"
	configPath  = "./namespaces/global/config/global.yaml"
)

func init() {
	logger = logging.MustGetLogger("test")
}

func initConfig() *common.Config {
	switchToExeLoc()
	return initLog()
}

func initDb() vm.Database {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	stateDb, _ := state.New(common.Hash{}, db, db, initConfig(), 0, "global")
	stateDb.CreateAccount(common.HexToAddress("e81e714395549ba939403c7634172de21367f8b5"))
	stateDb.Commit()
	return stateDb
}

//func startServer() {
//	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50052))
//	if err != nil {
//		fmt.Printf("failed to listen: %v\n", err)
//	}
//	grpcServer := grpc.NewServer()
//	//ledger := lg.NewLedgerProxy()
//	ledger.Register("global", initDb())
//	pb.RegisterLedgerServer(grpcServer, ledger)
//	grpcServer.Serve(lis)
//}

func main() {
	fmt.Println([]byte("100001"))
	//go startServer()
	//exe := jcee.NewContractExecutor()
	//exe.Start()
	//testNum := 10 * 10000
	//t1 := time.Now()
	//for i := 0; i < testNum; i++ {
	//	//time.Sleep(3 * time.Second)
	//	request := &pb.Request{
	//		Context: &pb.RequestContext{
	//			Txid: "tx000000" + strconv.Itoa(i),
	//			Namespace: "global",
	//			Cid:  "e81e714395549ba939403c7634172de21367f8b5",
	//		},
	//		Method: "invoke",
	//		Args:   [][]byte{[]byte("issue"), []byte("user1"), []byte("1000")},
	//	}
	//	response, err := exe.Execute(request)
	//	//_, err := exe.Execute(request)
	//
	//	if err != nil {
	//		logger.Error(err)
	//	}
	//	logger.Info(response)
	//}
	//t2 := time.Now()
	//
	////logger.Critical((testNum * 1.0) / t2.Sub(t1).Seconds())
	//a := (float64(1.0 * testNum)) / t2.Sub(t1).Seconds()
	//logger.Critical(a)
	//
	//x := make(chan bool)
	//<-x
}

func switchToExeLoc() string {
	owd, _ := os.Getwd()
	os.Chdir(path.Join(common.GetGoPath(), "src/hyperchain/configuration"))
	return owd
}

func initLog() *common.Config {
	globalConfig := common.NewConfig(configPath)
	common.InitHyperLoggerManager(globalConfig)
	globalConfig.Set(common.NAMESPACE, "global")
	common.InitHyperLogger(globalConfig)
	return globalConfig
}
