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
		"TransactionCreate",
		"POST",
		"/trans",
		controller.TransactionCreate,
	},
	Route{
		"TransactionGet",
		"GET",
		"/trans",
		controller.TransactionGet,
	},
	Route{
		"BalancesGet",
		"GET",
		"/balances",
		controller.BalancesGet,
	},
	Route{
		"BlocksGet",
		"GET",
		"/blocks",
		controller.BlocksGet,
	},
}