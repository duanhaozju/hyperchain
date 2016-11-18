package routers

import (
	"github.com/gorilla/mux"
	"net/http"
	//"hyperchain/jsonrpc/restful/logger"
	"fmt"
)
func NewRouter() *mux.Router {
	fmt.Println("======================== enter NewRouter ======================")
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		//handler = logger.Logger(handler, route.Name)

		router.
		Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}