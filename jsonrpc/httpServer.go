package jsonrpc

import (
	"hyperchain/jsonrpc/routers"
	"strconv"
	"net/http"
	"log"
	"github.com/ethereum/go-ethereum/common"
)

func StartHttp(httpPort int,from common.Hash){
	//实例化路由
	router := routers.NewRouter()
	// 指定静态文件目录
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../")))

	//启动http服务
		log.Println("启动http服务...")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(httpPort),router))
}
