//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"hyperchain/event"
	"testing"
	"hyperchain/membersrvc"
)
func init(){
	membersrvc.Start("../config/test/local_membersrvc2.yaml",1)
}

func TestGrpcPeerManager_Start(t *testing.T) {

	path := "../config/local_peerconfig.json"

	grpcPeerMgr := NewGrpcManager(path, 1)
	aliveChan := make(chan bool)
	eventMux := new(event.TypeMux)
	go grpcPeerMgr.Start(aliveChan, eventMux,false)

	grpcPeerMgr2 := NewGrpcManager(path, 2)

	go grpcPeerMgr2.Start(aliveChan, eventMux,false)

	grpcPeerMgr3 := NewGrpcManager(path, 3)

	go grpcPeerMgr3.Start(aliveChan, eventMux,false)

	grpcPeerMgr4 := NewGrpcManager(path, 4)

	go grpcPeerMgr4.Start(aliveChan, eventMux,false)
	// wait the sub thread done
	nodeCount := 0
	for flag := range aliveChan {
		if flag {
			log.Info("A peer has connected")
			nodeCount += 1
		}
		if nodeCount >= 4 {
			break
		}
	}

}

//func TestGrpcPeerManager_GetPeerInfos(t *testing.T) {
//	path := "./peerconfig.json"
//
//	grpcPeerMgr := new(GrpcPeerManager)
//	aliveChan := make(chan bool)
//	eventMux := new(event.TypeMux)
//	go grpcPeerMgr.Start(path,1, aliveChan,true,eventMux )
//
//	grpcPeerMgr2 := new(GrpcPeerManager)
//
//	go grpcPeerMgr2.Start(path,2, aliveChan,true,eventMux)
//
//	grpcPeerMgr3 := new(GrpcPeerManager)
//
//	go grpcPeerMgr3.Start(path,3, aliveChan,true,eventMux)
//
//	grpcPeerMgr4 := new(GrpcPeerManager)
//
//	go grpcPeerMgr4.Start(path,4, aliveChan,true,eventMux)
//	// wait the sub thread done
//	nodeCount := 0
//	for flag := range aliveChan{
//		if flag{
//			log.Info("A peer has connected")
//			nodeCount += 1
//		}
//		if nodeCount >=4{
//			perinfos := grpcPeerMgr4.GetPeerInfos()
//			t.Log(perinfos.GetNumber())
//			break
//		}
//	}
//}
