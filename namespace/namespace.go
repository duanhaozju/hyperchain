//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"hyperchain/common"
	"hyperchain/api/jsonrpc/core"
)

//Namespace represent the namespace instance
type Namespace interface {
	//Start start services under this namespace.
	Start()
	//Stop stop services under this namespace.
	Stop()
	//Status return current namespace status.
	Status() *Status
	//Info return basic information of this namespace.
	Info() *NamespaceInfo
	//ProcessRequest process request under this namespace
	ProcessRequest(request interface{}) interface{}
}

const (
	STARTTING = iota
	STARTTED
	RUNNING
	CLOSING
	CLOSED
)

//Status dynamic state of current namespace.
type Status struct {
	state int
	desc  string
}

//NamespaceInfo basic information of this namespace.
type NamespaceInfo struct {
	name    string
	members []string //member ips list
	desc    string
}

//namespaceImpl implementation of Namespace
type namespaceImpl struct {
	//TODO:
}

func newNamespaceImpl(name string, conf *common.Config) Namespace {
	//TODO:
	return nil
}

func GetNamespace(name string, conf *common.Config) Namespace {
	return newNamespaceImpl(name, conf)
}

//Start start services under this namespace.
func (ns *namespaceImpl) Start() {
	//TODO:
}

//Stop stop services under this namespace.
func (ns *namespaceImpl) Stop() {

}

//Status return current namespace status.
func (ns *namespaceImpl) Status() *Status {
	//TODO:
	return nil
}

//Info return basic information of this namespace.
func (ns *namespaceImpl) Info() *NamespaceInfo {
	//TODO:
	return nil
}

//ProcessRequest process request under this namespace
func (ns *namespaceImpl) ProcessRequest(request interface{}) interface{}{

	switch r := request.(type) {
	case *jsonrpc.JSONRequest:
		return ns.handleJsonRequest(r)
	default:
		logger.Errorf("event not supportted %v", r)
	}
	return nil
}
