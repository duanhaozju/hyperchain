// author: Xiaoyi Wang
// email: wangxiaoyi@hyperchain.cn
// date: 16/11/1
// last modified: 16/11/1
// last Modified Author: Xiaoyi Wang
// change log: new test for helper

package pbft

import (
	"testing"
	"time"
	"hyperchain/consensus/events"
	"strings"
	"hyperchain/core/types"
	"reflect"
)

func TestSortableUint64SliceFunctions(t *testing.T) {
	slice := sortableUint64Slice{1, 2, 3, 4, 5}
	if slice.Len() != 5{
		t.Errorf("error slice.len != 5")
	}
	if slice.Less(2, 3) != true{
		t.Errorf("error slice[2] >= slice[3]")
	}
	if slice.Swap(2, 3); !(slice[2] == 4 && slice[3] ==3){
		t.Error("error exchange slice[2], slice[3]")
	}
}


func TestPbftStateFunctions(t *testing.T)  {
	pp := new(pbftProtocal)
	pp.valid = false

	pp.validateState()
	if pp.valid == false {
		t.Errorf("pbftProtocal %s function not worked!", "validateState")
	}

	pp.valid = true;
	pp.invalidateState()

	if pp.valid == true {
		t.Errorf("pbftProtocal %s function not worked!", "invalidateState")
	}
}

func TestPbftTimeFunctions(t *testing.T)  {
	pp := new (pbftProtocal)
	pp.batchTimeout = 1 * time.Second
	pp.batchTimerActive = false;
	pp.batchTimer = events.NewTimerFactoryImpl(events.NewManagerImpl()).CreateTimer()

	pp.startBatchTimer()
	if(pp.batchTimeout != 1 * time.Second || pp.batchTimerActive == false){
		t.Errorf("pbftProtocal %s not work!", "startBatchTimer")
	}

	pp.stopBatchTimer()
	if pp.batchTimerActive == true {
		t.Errorf("pbftProtocal %s not work!", "stopBatchTimer")
	}

	pp.newViewTimer =  events.NewTimerFactoryImpl(events.NewManagerImpl()).CreateTimer()
	pp.startTimer(2 * time.Second, "test pbftProtocol viewTimer")
	if pp.timerActive == false {
		t.Errorf("pbftProtocal %s not work!", "startTimer")
	}

	pp.stopTimer()
	if pp.timerActive == true {
		t.Errorf("pbftProtocal %s not work!", "stopTimer")
	}
	rs := "test pbftProtocol softSDtartTimer"
	pp.softStartTimer(1 * time.Second, rs)
	if pp.timerActive == false || strings.Compare(pp.newViewTimerReason, rs) != 0 {
		t.Errorf(`pbftProtocal %s not work!`, "softStartTimer")
	}

	/*pp.nullRequestTimer = events.NewTimerFactoryImpl(events.NewManagerImpl()).CreateTimer()
	pp.nullRequestTimeout = 2 * time.Second
	pp.id = 1
	pp.view = 1
	pp.requestTimeout = 3


	pp.nullReqTimerReset()
	pp.nullRequestTimer*/
}

func TestPrimary(t *testing.T)  {
	pp := new (pbftProtocal)
	pp.replicaCount = 100
	x := pp.primary(3)
	if x != 4 {
		t.Errorf("primary(%d) == %d, actual: %d", 3, x, 4)
	}

	r1 := pp.primary(0)
	if r1 != 1 {
		t.Errorf("primary(%d) == %d, actual: %d", 0, r1, 1)
	}

	pp.h = 0
	pp.L = 100

	if !pp.inW(100) {
		t.Errorf("inw(%d) = false, actual true", 100)
	}
	if pp.inW(101) {
		t.Errorf("inw(%d) = true, actual false", 101)
	}

	pp.view = 9

	if !pp.inWV(9, 30) {
		t.Errorf("inWV(%d, %d) = false, actual true", 9, 30)
	}

	if pp.inWV(8, 30) {
		t.Errorf("inWV(%d, %d) = true, actual false", 9, 30)
	}

	if pp.inWV(9, 200) {
		t.Errorf("inWV(%d, %d) = true, actual false", 9, 30)
	}
}

func TestGetSert(t *testing.T)  {
	pp := new (pbftProtocal)
	pp.certStore = make(map[msgID]*msgCert)
	pp.getCert(1, 2)

	idx := msgID{1, 2}
	_, ok := pp.certStore[idx]
	if !ok {
		t.Error("getCert not worked")
	}

	pp.chkptCertStore = make(map[chkptID]*chkptCert)
	idx2 := chkptID{1, "123"}
	pp.getChkptCert(1, "123")

	_, ok = pp.chkptCertStore[idx2]
	if !ok {
		t.Error("getChkptCert error")
	}
}

func TestPrePrepared(t *testing.T)  {
	pbft := new(pbftProtocal)
	pbft.validatedBatchStore = make(map[string]*TransactionBatch)
	pbft.view = 2
	pbft.seqNo = 100

	pbft.validatedBatchStore["d1"] = &TransactionBatch{nil, int64(1001)}

	digest := ""
	v := uint64(2)
	seqNo := uint64(100)

	rs := pbft.prePrepared(digest, v, seqNo)

	if rs {
		t.Errorf("error prePrepared(%q, %d, %d) = %t, actual: %t", digest, v, seqNo, rs, false)
	}
	digest = "d1"
	pbft.certStore = make(map[msgID]*msgCert)
	prepare := make(map[Prepare]bool)
	prePrepare0 := &PrePrepare{
				View:v,
				SequenceNumber:seqNo,
				BatchDigest: digest,
	}

	commit := make(map[Commit]bool)
	cert0 := &msgCert{
		prepare:	prepare,
		commit:		commit,
		digest:     digest,
		prePrepare: prePrepare0,

	}

	idx := msgID{v, seqNo}
	pbft.certStore[idx] = cert0

	rs = pbft.prePrepared(digest, v, seqNo)

	if !rs {
		t.Errorf("error prePrepared(%q, %d, %d) = %t, actual: %t", digest, v, seqNo, rs, true)
	}
}

func TestPreparedReplicasQuorum(t *testing.T)  {
	pbft := new (pbftProtocal)
	pbft.f = 1
	k := pbft.preparedReplicasQuorum()
	if  k != 2 {
		t.Errorf("error preparedReplicasQuorum() = %d, expected: 2", k)
	}
}

func TestCommittedReplicasQuorum(t *testing.T)  {
	pbft := new (pbftProtocal)
	pbft.f = 1
	k := pbft.committedReplicasQuorum()
	if  k != 3 {
		t.Errorf("error committedReplicasQuorum() = %d, expected: 3", k)
	}
}

func TestIntersectionQuorum(t *testing.T)  {
	pbft := new (pbftProtocal)
	pbft.N = 1
	pbft.f = 1
	k := pbft.intersectionQuorum()
	if k != 2 {
		t.Errorf("error intersectionQuorum() = %d, expected: %d", k, 2)
	}
	pbft.N = 1
	pbft.f = 4
	k = pbft.intersectionQuorum()
	if k != 3 {
		t.Errorf("error intersectionQuorum() = %d, expected: %d", k, 3)
	}
}

func TestAllCorrectReplicasQuorum(t *testing.T)  {
	pbft := new (pbftProtocal)
	pbft.N = 100
	pbft.f = 33
	k := pbft.allCorrectReplicasQuorum()
	if k != 67 {
		t.Errorf("error allCorrectReplicasQuorum() = %d, expected: %d", k, 67)
	}
}

func TestPostRequestEvent(t *testing.T)  {
	pbft := new(pbftProtocal)

	queue := pbft.pbftManager.Queue()
	tx := &types.Transaction{Id:123}

	pbft.pbftManager = events.NewManagerImpl()
	pbft.postRequestEvent(tx)

	trx := <- queue
	if reflect.DeepEqual(trx, tx) != true {
		t.Errorf("error postRequestEvent(), not post successful!")
	}
}