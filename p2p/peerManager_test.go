// author: chenquan
// date: 16-8-26
// last modified: 16-8-26 13:08
// last Modified Author: chenquan
// change log:
// comment: this file can not test easily, so just waiting the integration test
//
package p2p

import (
	"testing"
	"hyperchain/event"
)
func TestGrpcPeerManager_Start(t *testing.T){

	path := "./local_peerconfig.json"

	grpcPeerMgr := new(GrpcPeerManager)
	aliveChan := make(chan bool)
	eventMux := new(event.TypeMux)
	go grpcPeerMgr.Start(path,1, aliveChan,true,eventMux )

	grpcPeerMgr2 := new(GrpcPeerManager)

	go grpcPeerMgr2.Start(path,2, aliveChan,true,eventMux)

	grpcPeerMgr3 := new(GrpcPeerManager)

	go grpcPeerMgr3.Start(path,3, aliveChan,true,eventMux)

	grpcPeerMgr4 := new(GrpcPeerManager)

	go grpcPeerMgr4.Start(path,4, aliveChan,true,eventMux)
	// wait the sub thread done
	nodeCount := 0
	for flag := range aliveChan{
		if flag{
			log.Info("A peer has connected")
			nodeCount += 1
		}
		if nodeCount >=4{
			break
		}
	}

}



func TestGrpcPeerManager_GetPeerInfos(t *testing.T) {
	path := "/home/chenquan/Workspace/IdeaProjects/hyperchain-go/src/hyperchain/p2p/local_peerconfig.json"

	grpcPeerMgr := new(GrpcPeerManager)
	aliveChan := make(chan bool)
	eventMux := new(event.TypeMux)
	go grpcPeerMgr.Start(path,1, aliveChan,true,eventMux )

	grpcPeerMgr2 := new(GrpcPeerManager)

	go grpcPeerMgr2.Start(path,2, aliveChan,true,eventMux)

	grpcPeerMgr3 := new(GrpcPeerManager)

	go grpcPeerMgr3.Start(path,3, aliveChan,true,eventMux)

	grpcPeerMgr4 := new(GrpcPeerManager)

	go grpcPeerMgr4.Start(path,4, aliveChan,true,eventMux)
	// wait the sub thread done
	nodeCount := 0
	for flag := range aliveChan{
		if flag{
			log.Info("A peer has connected")
			nodeCount += 1
		}
		if nodeCount >=4{
			perinfos := grpcPeerMgr4.GetPeerInfos()
			t.Log(perinfos.GetNumber())
			break
		}
	}
}