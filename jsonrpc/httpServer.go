package jsonrpc

import (
	"hyperchain/jsonrpc/routers"
	"strconv"
	"net/http"
	"log"

	"hyperchain/event"
	"hyperchain/jsonrpc/controller"
)

func StartHttp(httpPort int,eventMux *event.TypeMux){

	eventMux=eventMux

	// Create new router and configurate it
	router := routers.NewRouter()

	// Specify static files directory
	if controller.Testing {
		router.PathPrefix("/").Handler(http.FileServer(http.Dir(".")))
	} else {
		router.PathPrefix("/").Handler(http.FileServer(http.Dir("./jsonrpc")))
	}


	// Start http server
	log.Println("启动http服务...")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(httpPort),router))
}

