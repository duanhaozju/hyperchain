//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"hyperchain/api/jsonrpc/core"
	"hyperchain/common"
)

//Namespace represent the namespace instance
type Namespace interface {
	//Start start services under this namespace.
	Start() error
	//Stop stop services under this namespace.
	Stop() error
	//Status return current namespace status.
	Status() *Status
	//Info return basic information of this namespace.
	Info() *NamespaceInfo
	//ProcessRequest process request under this namespace.
	ProcessRequest(request interface{}) interface{}

	//Name of current namespace.
	Name() string
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
	nsInfo *NamespaceInfo
	status *Status
}

func newNamespaceImpl(name string, conf *common.Config) (*namespaceImpl, error) {

	ninfo := &NamespaceInfo{
		name: name,
		//more info to record
	}
	//TODO: more to init
	status := &Status{
		state: STARTTING,
		desc:  "start ting",
	}

	ns := &namespaceImpl{
		nsInfo: ninfo,
		status: status,
	}

	return ns, nil
}

func GetNamespace(name string, conf *common.Config) (Namespace, error) {
	ns, err := newNamespaceImpl(name, conf)
	if err != nil {
		logger.Errorf("namespace %s init error", name)
		return ns, err
	}
	err = ns.init()
	return ns, err
}

func (ns *namespaceImpl) init() error {
	//TODO: init the namespace by configuration
	logger.Criticalf("Init namespace %s", ns.Name())

	return nil
}

//Start start services under this namespace.
func (ns *namespaceImpl) Start() error {
	//TODO:

	ns.status.state = STARTTED
	return nil
}

//Stop stop services under this namespace.
func (ns *namespaceImpl) Stop() error {
	//TODO: stop a namespace service

	ns.status.state = CLOSED
}

//Status return current namespace status.
func (ns *namespaceImpl) Status() *Status {
	return ns.status
}

//Info return basic information of this namespace.
func (ns *namespaceImpl) Info() *NamespaceInfo {
	return ns.nsInfo
}

func (ns *namespaceImpl) Name() string {
	return ns.nsInfo.name
}

//ProcessRequest process request under this namespace
func (ns *namespaceImpl) ProcessRequest(request interface{}) interface{} {
	switch r := request.(type) {
	case *jsonrpc.JSONRequest:
		return ns.handleJsonRequest(r)
	default:
		logger.Errorf("event not supportted %v", r)
	}
	return nil
}
