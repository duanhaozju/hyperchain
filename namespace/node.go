//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package namespace

import (
	"github.com/op/go-logging"
	//"github.com/pkg/errors"
	"sync"
)

// Node contains the basic node info exposed in namespace.
type Node struct {
	Id           int
	Addr         string
	ExternalAddr string
	GrpcPort     int
	RpcPort      int
}

func NewNode(id, grpc, rpc int, addr, extAddr string) *Node {
	return &Node{
		Id:           id,
		Addr:         addr,
		ExternalAddr: extAddr,
		GrpcPort:     grpc,
		RpcPort:      rpc,
	}
}

// NamespaceInfo contains the basic information of this namespace.
type NamespaceInfo struct {
	name   string
	desc   string
	logger *logging.Logger

	nodes map[int]*Node
	lock *sync.RWMutex
}

// NewNamespaceInfo returns namespace info by peerconfig file.
func NewNamespaceInfo(peerConfigPath, namespace string, logger *logging.Logger) (*NamespaceInfo, error) {
	ni := &NamespaceInfo{
		name:   namespace,
		logger: logger,
		lock:   &sync.RWMutex{},
	}
	return ni, ni.init(peerConfigPath, namespace)
}

// During add node or delete node, this method must be invoked.
func (ni *NamespaceInfo) init(peerConfigPath, namespace string) error {
	//TODO: complement namespace info gathering
	return nil
}

func (ni *NamespaceInfo) MaxNodeNum() int {
	ni.lock.RLock()
	defer ni.lock.RUnlock()
	return len(ni.nodes)
}

func (ni *NamespaceInfo) Nodes() map[int]*Node {
	return ni.nodes
}

func (ni *NamespaceInfo) Name() string {
	return ni.name
}

func (ni *NamespaceInfo) Desc() string {
	return ni.name
}

// Node returns node by node id.
func (ni *NamespaceInfo) Node(id int) (*Node, error) {
	ni.lock.RLock()
	defer ni.lock.RUnlock()
	if node, ok := ni.nodes[id]; ok {
		return node, nil
	}
	ni.logger.Errorf("Node with id %d is not found", id)
	return nil, ErrNodeNotFound
}

func (ni *NamespaceInfo) AddNode(node *Node) error {
	ni.lock.Lock()
	defer ni.lock.Unlock()
	if !isNodeConfigLegal(node) {
		return ErrIllegalNodeConfig
	}
	if _, ok := ni.nodes[node.Id]; ok {
		ni.logger.Warningf("node with id %d is existed, update it.", node.Id)
		ni.nodes[node.Id] = node
	}
	ni.logger.Info("Add newed node with id %d ", node.Id)
	ni.logger.Debugf("Add newed node: %v ", node)
	ni.nodes[node.Id] = node
	return nil
}

func (ni *NamespaceInfo) PrintInfo() {
	ni.lock.RLock()
	defer ni.lock.RUnlock()
	for _, node := range ni.nodes {
		ni.logger.Debugf("node info %v ", node)
	}
}

// isNodeConfigLegal checks the legality of node config.
func isNodeConfigLegal(node *Node) bool {
	//TODO: more limits?
	if node == nil ||
		len(node.Addr) == 0 ||
		len(node.ExternalAddr) == 0 ||
		node.Id < 0 ||
		node.GrpcPort < 0 ||
		node.RpcPort < 0 {
		return false
	}
	return true
}
