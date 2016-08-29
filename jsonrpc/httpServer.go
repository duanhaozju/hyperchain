package jsonrpc

import (
	"hyperchain/jsonrpc/routers"
	"strconv"
	"net/http"
	"log"

	"hyperchain/event"
)


func StartHttp(httpPort int,eventMux *event.TypeMux){
	//实例化路由
	eventMux=eventMux
	router := routers.NewRouter()
	// 指定静态文件目录
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./jsonrpc")))

	//启动http服务
		log.Println("启动http服务...")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(httpPort),router))
}

