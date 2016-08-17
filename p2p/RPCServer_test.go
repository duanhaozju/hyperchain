package p2p

import (
	"testing"
	"hyperchain.cn/app/jsonrpc/model"
)

func  TestStrarRPCServer(t *testing.T){
	rnodes := SaveNode("localhost:8081",model.Node{P2PAddr:"localhost",P2PPort:1001,HTTPPORT:1002})
	if len(rnodes) != 2 {
		t.Errorf("error %#v",rnodes)
	}
}