package rpc

import (
	"hyperchain/common"
	"reflect"
)

// callback is a method callback which was registered in the server
type callback struct {
	rcvr        reflect.Value  // receiver of method
	method      reflect.Method // callback
	argTypes    []reflect.Type // input argument types
	hasCtx      bool           // method's first argument is a context (not included in argTypes)
	errPos      int            // err return idx, of -1 when method cannot return error
	isSubscribe bool           // indication if the callback is a subscription
}

// service represents a registered object
type service struct {
	name          string        // name for service
	rcvr          reflect.Value // receiver of methods for the service
	typ           reflect.Type  // receiver type
	callbacks     callbacks     // registered handlers
	subscriptions subscriptions // available subscriptions/notifications
}

// serverRequest is an incoming request
type serverRequest struct {
	id            interface{}
	svcname       string
	rcvr          reflect.Value
	callb         *callback
	args          []reflect.Value
	isUnsubscribe bool
	err           common.RPCError
}

type serviceRegistry map[string]*service // collection of services
type callbacks map[string]*callback      // collection of RPC callbacks
type subscriptions map[string]*callback  // collection of subscription callbacks
//type subscriptionRegistry map[string]*callback // collection of subscription callbacks

// API describes the set of methods offered over the RPC interface
type API struct {
	Srvname string      // srvname under which the rpc methods of Service are exposed
	Version string      // api version for DApp's
	Service interface{} // receiver instance which holds the methods
	Public  bool        // indication if the methods must be considered safe for public use
}
