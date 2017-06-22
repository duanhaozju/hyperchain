package pbft

//
//import (
//	"testing"
//	"hyperchain/event"
//	"hyperchain/core"
//	"hyperchain/consensus/helper"
//	"reflect"
//)
//
//func TestInitRecovery(t *testing.T) {
//	core.InitDB("./build/keystore", 8001)
//
//	id := 1
//	pbftConfigPath := getPbftConfigPath()
//	config := loadConfig(pbftConfigPath)
//	eventMux := new(event.TypeMux)
//	h := helper.NewHelper(eventMux)
//	pbft := newPBFT(uint64(id), config, h)
//	defer pbft.Close()
//
//	pbft.initRecovery()
//
//}
//
//func TestRecvRecoveryRsp(t *testing.T) {
//	core.InitDB("./build/keystore", 8088)
//
//	id := 1
//	pbftConfigPath := getPbftConfigPath()
//	config := loadConfig(pbftConfigPath)
//	eventMux := new(event.TypeMux)
//	h := helper.NewHelper(eventMux)
//	pbft := newPBFT(uint64(id), config, h)
//	defer pbft.Close()
//
//	pbft.rcRspStore = make(map[uint64]*RecoveryResponse)
//
//	// pbft not in recovery, should just return
//	pbft.inRecovery = false
//	chkpts := make(map[uint64]string)
//	for n, d := range pbft.chkpts {
//		chkpts[n] = d
//	}
//	fromId := uint64(2)
//	rc := &RecoveryResponse{
//		ReplicaId:	fromId,
//		Chkpts: 	chkpts,
//	}
//	pbft.recvRecoveryRsp(rc)
//	if _, ok := pbft.rcRspStore[fromId]; ok {
//		t.Errorf("Pbft not in recovery, start processing recovery response, expect not")
//	}
//
//	// pbft in recovery
//	// already same 'from' response cached
//	rspPlaceholder := &RecoveryResponse{}
//	pbft.rcRspStore[fromId] = rspPlaceholder
//	pbft.recvRecoveryRsp(rc)
//	rsp := pbft.rcRspStore[fromId]
//	if !reflect.DeepEqual(rspPlaceholder, rsp) {
//		t.Errorf("Pbft process rsp from same source, expect just return")
//	}
//
//}
//
//func TestFindHighestChkptQuorum(t *testing.T) {
//	core.InitDB("./build/keystore", 8088)
//	id := 1
//	pbftConfigPath := getPbftConfigPath()
//	config := loadConfig(pbftConfigPath)
//	eventMux := new(event.TypeMux)
//	h := helper.NewHelper(eventMux)
//	pbft := newPBFT(uint64(id), config, h)
//	defer pbft.Close()
//
//	pbft.rcRspStore = make(map[uint64]*RecoveryResponse)
//
//}
