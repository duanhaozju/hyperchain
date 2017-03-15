//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

// API describes the set of methods offered over the RPC interface
type API struct {
	Srvname string      // srvname under which the rpc methods of Service are exposed
	Version string      // api version
	Service interface{} // receiver instance which holds the methods
	Public  bool        // indication if the methods must be considered safe for public use
}

var Apis []API

func GetApiObjectByNamespace(name string) API {
	for _, api := range Apis { //TODO: change this into name <-> service map
		if api.Srvname == name {
			return api
		}
	}
	return API{}
}
