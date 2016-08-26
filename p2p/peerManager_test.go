// author: chenquan
// date: 16-8-26
// last modified: 16-8-26 13:08
// last Modified Author: chenquan
// change log:
//
package p2p

import (
	"testing"
	"sync"
)

func TestGrpcPeerManager_Start(t *testing.T){

	path := "/home/chenquan/Workspace/IdeaProjects/hyperchain-go/src/hyperchain-alpha/p2p/peerconfig.json"
	var sc sync.WaitGroup
	grpcPeerMgr := new(GrpcPeerManager)
	aliveChan := make(chan bool)
	grpcPeerMgr.Start(path, sc,3, aliveChan)
	<-aliveChan

	sc.Add(4)
	sc.Wait()
}
