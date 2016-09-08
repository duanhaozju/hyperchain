package jsonrpc

import (
	"hyperchain/jsonrpc/routers"
	"strconv"
	"net/http"

	"hyperchain/event"
	"hyperchain/jsonrpc/controller"
	"github.com/op/go-logging"
)
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("jsonrpc")
}
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
	log.Info("启动http服务...")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(httpPort),router))
}

func start() {

}

