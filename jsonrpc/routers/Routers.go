package routers

import (
	"net/http"
	"hyperchain/jsonrpc/controller"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		controller.Index,
	},
	Route{
		"TransactionIndex",
		"GET",
		"/trans",
		controller.TransactionIndex,
	},
	Route{
		"TransactionShow",
		"GET",
		"/trans/{transId}",
		controller.TransacionShow,
	},
	Route{
		"TransactionCreate",
		"POST",
		"/trans",
		controller.TransactionCreate,
	},
}