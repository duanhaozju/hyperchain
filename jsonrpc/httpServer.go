package jsonrpc

import (
	"hyperchain/jsonrpc/routers"
	"strconv"
	"net/http"
	"log"
	"github.com/ethereum/go-ethereum/common"
)

func StartHttp(httpPort int,from common.Hash,to common.Hash ){
	//实例化路由
	router := routers.NewRouter()
	// 指定静态文件目录
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../")))

	//启动http服务
		log.Println("启动http服务...")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(httpPort),router))
}

// TODO NewTransaction
// TODO GetBalance
// TODO 抛事件 POST
// eventmux:=new(event.TypeMux)
// eventmux.Post(event.NewTxEvent{[]byte{0x00, 0x00, 0x03, 0xe8}})

// TODO GetAllTransactions

// TODO GetAllBalances
