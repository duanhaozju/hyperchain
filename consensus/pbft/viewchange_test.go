package pbft

import(
	"testing"

	"hyperchain/event"
	"hyperchain/consensus/helper"
	"hyperchain/core"
	"reflect"
	"hyperchain/core/types"

	"encoding/base64"
	"github.com/golang/protobuf/proto"
	"hyperchain/protos"
)


func getTestViewChange() (vc *ViewChange){
	vc = &ViewChange{
		View: 1,

		Cset: []*ViewChange_C{
			{
				SequenceNumber:60,
				Id:"chkponit---digest60",
			},
			{
				SequenceNumber:70,
				Id:"chkpoint---digest70",
			},
		},

		Pset: []*ViewChange_PQ{
			{
				SequenceNumber:51,
				BatchDigest:"batch---P1",
				View:0,
			},
			{
				SequenceNumber:52,
				BatchDigest:"batch---P2",
				View:0,
			},
			{
				SequenceNumber:53,
				BatchDigest:"batch---P3",
				View:0,
			},
		},

		Qset: []*ViewChange_PQ{
			{
				SequenceNumber:54,
				BatchDigest:"batch---Q4",
				View:0,
			},
			{
				SequenceNumber:55,
				BatchDigest:"batch---Q5",
				View:0,
			},
			{
				SequenceNumber:56,
				BatchDigest:"batch---Q6",
				View:0,
			},
		},

		H: 50,

		ReplicaId: 2,
	}

	return
}

func TestCorrectViewChange(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := "../../config/pbft.yaml"
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPBFT(uint64(id), config, h)

	vc := getTestViewChange()

	if !(pbft.correctViewChange(vc) == true){
		t.Error("Should be a correct view change.")
	}

	vc.View = 0
	if pbft.correctViewChange(vc) == true{
		t.Error("Should not be a correct view change: vc.View is lower than view in PQset.")
	}
	vc.View = 2

	vc.H = 80
	if pbft.correctViewChange(vc) == true{
		t.Error("Should not be a correct view change: vc.H is too high.")
	}
	vc.H = 60

	vc.H = 10
	if pbft.correctViewChange(vc) == true{
		t.Error("Should not be a correct view change: vc.H is too low.")
	}
}

func TestCalcPSet(t *testing.T){
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 2
	pbftConfigPath := "../../config/pbft.yaml"
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPBFT(uint64(id), config, h)

	pbft.certStore = map[msgID]*msgCert{}
	calc_pset := pbft.calcPSet()
	if !(len(calc_pset) == 0){
		t.Errorf("pset should be nil, but it is :%v", calc_pset)
	}

	pbft.certStore = map[msgID]*msgCert{
		msgID{v:1,n:2}: {digest:"batch1-2",prePrepare:&PrePrepare{View:1,SequenceNumber:2,BatchDigest:"batch1-2"}},
	}
	calc_pset = pbft.calcPSet()
	if !(len(calc_pset) == 0){
		t.Errorf("pset should be nil, but it is :%v", calc_pset)
	}

	pbft.certStore[msgID{1,2}].sentPrepare = true
	pbft.validatedBatchStore["batch1-2"] = &TransactionBatch{Batch:[]*types.Transaction{}, Timestamp:1}
	calc_pset = pbft.calcPSet()
	if !(len(calc_pset) == 0){
		t.Errorf("pset should be nil, but it is :%v", calc_pset)
	}

	pbft.certStore[msgID{v:1,n:2}].prepare = map[Prepare]bool{Prepare{View:1,SequenceNumber:2,BatchDigest:"batch1-2"}:true}
	pbft.certStore[msgID{v:1,n:2}].prepareCount = 2*(pbft.f) + 1
	calc_pset = pbft.calcPSet()
	if (len(calc_pset) == 0){
		t.Errorf("pset should not be nil, but it is :%v", calc_pset)
	}
}

func TestCalcQSet(t *testing.T) {
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 2
	pbftConfigPath := "../../config/pbft.yaml"
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPBFT(uint64(id), config, h)

	pbft.certStore = map[msgID]*msgCert{}
	calc_qset := pbft.calcQSet()
	if !(len(calc_qset) == 0){
		t.Errorf("qset should be nil, but it is :%v", calc_qset)
	}

	pbft.certStore = map[msgID]*msgCert{
		msgID{v:1,n:2}: {digest:"batch1-2",prePrepare:&PrePrepare{View:1,SequenceNumber:2,BatchDigest:"batch1-2"}},
	}
	calc_qset = pbft.calcQSet()
	if !(len(calc_qset) == 0){
		t.Errorf("qset should be nil, but it is :%v", calc_qset)
	}

	pbft.certStore[msgID{v:1,n:2}].sentPrepare = true
	pbft.validatedBatchStore["batch1-2"] = &TransactionBatch{Batch:[]*types.Transaction{}, Timestamp:1}
	_, mInLog := pbft.validatedBatchStore["batch1-2"]
	calc_qset = pbft.calcQSet()
	if (len(calc_qset) == 0){
		t.Errorf("qset should not be nil, but it is :%v and transcation batch is %v and msg digest is %v", calc_qset, mInLog, pbft.certStore[msgID{1,2}].digest)
	}

}


func TestRecvAndSendViewChange(t *testing.T){
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := "../../config/pbft.yaml"
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPBFT(uint64(id), config, h)

	pbft.sendViewChange()

	vc := getTestViewChange()

	pbft.inNegoView = true
	if !(pbft.recvViewChange(vc) == nil){
		t.Errorf("Replica %d is in nego-view, so it should not receive this view-change message", pbft.id)
	}
	pbft.inNegoView = false

	pbft.inRecovery = true
	if !(pbft.recvViewChange(vc) == nil){
		t.Errorf("Replica %d is in recovery, so it should not receive this view-change message", pbft.id)
	}
	pbft.inRecovery = false

	pbft.view = 3
	if !(pbft.recvViewChange(vc) == nil){
		t.Errorf("Replica %d found view-change message for old view, so it should not receive this view-change message", pbft.id)
	}
	pbft.view = 0

	vc.Pset[0].View = 5
	if !(pbft.recvViewChange(vc) == nil){
		t.Errorf("Replica %d found view-change message incorrect, so it should not receive this view-change message", pbft.id)
	}
	vc.Pset[0].View = 0

	pbft.recvViewChange(vc)
	if !(reflect.DeepEqual(pbft.viewChangeStore[vcidx{vc.View, vc.ReplicaId}], vc)){
		t.Error("Error in receving view-change meaasge")
	}

	pbft.recvViewChange(vc)

	vc1 := getTestViewChange()
	vc1.ReplicaId = 3
	pbft.recvViewChange(vc1)
	if !(pbft.view == 1){
		t.Errorf("Error:Replica%d receive f+1 view-change messages with larger view but didn't go into viewchange", pbft.id)
	}

	vc2 := getTestViewChange()
	vc2.ReplicaId = 4
	pbft.recvViewChange(vc2)
	if !(16000000000 == pbft.lastNewViewTimeout){
		t.Errorf("Now, lastNewViewTimeout should be 16 but it is %d.", pbft.lastNewViewTimeout)
	}

}

func TestSendAndRecvNewView(t *testing.T){
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := "../../config/pbft.yaml"
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPBFT(uint64(id), config, h)

	id = 2
	pbft2 := newPBFT(uint64(id), config, h)

	chkpt50 := &protos.BlockchainInfo{
		Height: 50,
		CurrentBlockHash: []byte("current block is 50."),
		PreviousBlockHash: []byte("previous block is 49."),
	}
	chkpt60 := &protos.BlockchainInfo{
		Height: 60,
		CurrentBlockHash: []byte("current block is 60."),
		PreviousBlockHash: []byte("previous block is 59."),
	}
	chkpt70 := &protos.BlockchainInfo{
		Height: 70,
		CurrentBlockHash: []byte("current block is 70."),
		PreviousBlockHash: []byte("previous block is 69."),
	}

	digest50,_ := proto.Marshal(chkpt50)
	digest60,_ := proto.Marshal(chkpt60)
	digest70,_ := proto.Marshal(chkpt70)

	vc1 := &ViewChange{
		View: 1,
		H: 40,
		ReplicaId: 1,

		Cset: []*ViewChange_C{
			{
				SequenceNumber:50,
				Id:base64.StdEncoding.EncodeToString(digest50),
			},
			{
				SequenceNumber:60,
				Id:base64.StdEncoding.EncodeToString(digest60),
			},
		},

		Pset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
		},

		Qset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
		},

	}


	vc2 := &ViewChange{
		View: 2,
		H: 40,
		ReplicaId: 2,

		Cset: []*ViewChange_C{
			{
				SequenceNumber:50,
				Id:base64.StdEncoding.EncodeToString(digest50),
			},
			{
				SequenceNumber:60,
				Id:base64.StdEncoding.EncodeToString(digest60),
			},
		},

		Pset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
		},

		Qset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:67,
				BatchDigest:"batch---67",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
		},

	}

	vc3 := &ViewChange{
		View: 2,
		H: 50,
		ReplicaId: 3,

		Cset: []*ViewChange_C{
			{
				SequenceNumber:60,
				Id:base64.StdEncoding.EncodeToString(digest60),
			},
			{
				SequenceNumber:70,
				Id:base64.StdEncoding.EncodeToString(digest70),
			},
		},

		Pset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
			{
				SequenceNumber:80,
				BatchDigest:"batch---80",
				View:0,
			},
		},

		Qset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:71,
				BatchDigest:"batch---71",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
		},

	}

	vc4 := &ViewChange{
		View: 2,
		H: 50,
		ReplicaId: 4,

		Cset: []*ViewChange_C{
			{
				SequenceNumber:60,
				Id:base64.StdEncoding.EncodeToString(digest60),
			},
		},

		Pset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
			{
				SequenceNumber:87,
				BatchDigest:"batch---87",
				View:0,
			},
		},

		Qset: []*ViewChange_PQ{
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
			{
				SequenceNumber:79,
				BatchDigest:"batch---79",
				View:0,
			},
		},

	}

	bool1 := pbft.correctViewChange(vc1)
	bool2 := pbft.correctViewChange(vc2)
	bool3 := pbft.correctViewChange(vc3)
	bool4 := pbft.correctViewChange(vc4)

	if !(bool1 == true && bool2 == true &&bool3 == true &&bool4 == true){
		t.Error("Should be 4 correct View-Change.")
	}

	pbft.viewChangeStore = map[vcidx]*ViewChange{
		vcidx{v:vc1.View, id:vc1.ReplicaId}: vc1,
		vcidx{v:vc2.View, id:vc2.ReplicaId}: vc2,
		vcidx{v:vc3.View, id:vc3.ReplicaId}: vc3,
		vcidx{v:vc4.View, id:vc4.ReplicaId}: vc4,
	}

	pbft.inNegoView = true
	if !(pbft.sendNewView() == nil){
		t.Errorf("Replica %d try to sendNewView, but it's in nego-view", pbft.id)
	}
	pbft.inNegoView = false

	vset := []*ViewChange{vc1,vc2,vc3,vc4}

	cp,ok,replicas := pbft.selectInitialCheckpoint(vset)
	if !(ok == true && reflect.DeepEqual(cp, ViewChange_C{SequenceNumber:60, Id:base64.StdEncoding.EncodeToString(digest60)}) && reflect.DeepEqual(replicas, []uint64{1,2,3,4})){
		t.Error("should get initial checkpoint!")
	}

	msglist := pbft.assignSequenceNumbers(vset, cp.SequenceNumber)
	if !(msglist[66] == vc1.Pset[0].BatchDigest){
		t.Error("sequence number 66 should be assigned but not assigned actually")
	}

	nv := &NewView{
		View:      2,
		Vset:      vset,
		Xset:      msglist,
		ReplicaId: 3,
	}
	pbft2.view = 2
	pbft2.activeView = false
	pbft2.recvNewView(nv)

	pbft.sendNewView()
}

func TestGetViewChanges(t *testing.T){
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := "../../config/pbft.yaml"
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPBFT(uint64(id), config, h)

	vc1 := getTestViewChange()
	vc1.ReplicaId = 1

	vc2 := getTestViewChange()
	vc2.ReplicaId = 2

	vc3 := getTestViewChange()
	vc3.ReplicaId = 3

	vc4 := getTestViewChange()
	vc4.ReplicaId = 4

	pbft.viewChangeStore = map[vcidx]*ViewChange{
		vcidx{v:vc1.View, id:vc1.ReplicaId}: vc1,
		vcidx{v:vc2.View, id:vc2.ReplicaId}: vc2,
		vcidx{v:vc3.View, id:vc3.ReplicaId}: vc3,
		vcidx{v:vc4.View, id:vc4.ReplicaId}: vc4,
	}

	vset := pbft.getViewChanges()
	count := 0
	for _,value := range vset{
		if value.ReplicaId == 1{
			if(reflect.DeepEqual(value, vc1)){
				count = count + 1
			}else{
				t.Errorf("Error while storing View-change1 to VSet.\nActual: %v\nExpect: %v", value, vc1)
			}
		}else if value.ReplicaId == 2{
			if(reflect.DeepEqual(value, vc2)){
				count = count + 1
			}else{
				t.Errorf("Error while storing View-change2 to VSet.\nActual: %v\nExpect: %v", value, vc2)
			}
		}else if value.ReplicaId == 3 {
			if(reflect.DeepEqual(value, vc3)) {
				count = count + 1
			} else {
				t.Errorf("Error while storing View-change3 to VSet.\nActual: %v\nExpect: %v", value, vc3)
			}
		}else if value.ReplicaId == 4{
			if(reflect.DeepEqual(value, vc4)){
				count = count + 1
			}else{
				t.Errorf("Error while storing View-change4 to VSet.\nActual: %v\nExpect: %v", value, vc4)
			}
		}else {
			t.Error("Error while storing View-change to VSet.")
		}
	}

	if !(count == 4){
		t.Error("Error while storing View-change to VSet: lack of View-change in VSet")
	}
}


func TestSelectInitialCheckpoint(t *testing.T){
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := "../../config/pbft.yaml"
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft1 := newPBFT(uint64(id), config, h)
	id = 2
	pbft2 := newPBFT(uint64(id), config, h)


	vc1 := getTestViewChange()
	vc1.ReplicaId = 1

	vc2 := getTestViewChange()
	vc2.ReplicaId = 2

	vc3 := getTestViewChange()
	vc3.ReplicaId = 3

	vc4 := getTestViewChange()
	vc4.ReplicaId = 4

	vset := []*ViewChange{}

	_,ok,_ := pbft1.selectInitialCheckpoint(vset)
	if !(ok == false){
		t.Error("VSet is nil. so we should not be able to select initial checkpoint.")
	}

	vc1.H = 20
	vc1.Cset = []*ViewChange_C{
		{
			SequenceNumber:30,
			Id:"chkpoint---digest30",
		},
		{
			SequenceNumber:40,
			Id:"chkpoint---digest40",
		},
	}

	vc2.H = 30
	vc2.Cset = []*ViewChange_C{
		{
			SequenceNumber:40,
			Id:"chkpoint---digest40",
		},

		{
			SequenceNumber:50,
			Id:"chkpoint---digest50",
		},
	}

	vc3.H = 50
	vc3.Cset = []*ViewChange_C{
		{
			SequenceNumber:60,
			Id:"chkpoint---digest60",
		},
		{
			SequenceNumber:70,
			Id:"chkpoint---digest70",
		},
	}

	vc4.H = 50
	vc4.Cset = []*ViewChange_C{
		{
			SequenceNumber:60,
			Id:"chkpoint---digest60",
		},
		{
			SequenceNumber:70,
			Id:"chkpoint---digest70",
		},
	}


	bool1 := pbft1.correctViewChange(vc1)
	bool2 := pbft1.correctViewChange(vc2)
	bool3 := pbft1.correctViewChange(vc3)
	bool4 := pbft1.correctViewChange(vc4)
	if !(bool1 == true && bool2 == true &&bool3 == true &&bool4 == true){
		t.Error("Should be 4 correct View-Change.")
	}

	vset = []*ViewChange{vc1,vc2,vc3,vc4}

	cp,ok,replicas := pbft2.selectInitialCheckpoint(vset)
	if !(ok == true && reflect.DeepEqual(cp, ViewChange_C{SequenceNumber:70, Id:"chkpoint---digest70"}) && reflect.DeepEqual(replicas, []uint64{3,4})){
		t.Error("should get initial checkpoint!")
	}

}

func TestAssignSequenceNumbers(t *testing.T){
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := "../../config/pbft.yaml"
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft1 := newPBFT(uint64(id), config, h)

	vc1 := &ViewChange{
		View: 1,
		H: 40,
		ReplicaId: 1,

		Cset: []*ViewChange_C{
			{
				SequenceNumber:50,
				Id:"chkpoint---digest50",
			},
			{
				SequenceNumber:60,
				Id:"chkpoint---digest60",
			},
		},

		Pset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
			{
				SequenceNumber:77,
				BatchDigest:"batch---77",
				View:0,
			},
		},

		Qset: []*ViewChange_PQ{
			{
				SequenceNumber:61,
				BatchDigest:"batch---61",
				View:0,
			},
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
		},

	}


	vc2 := &ViewChange{
		View: 2,
		H: 40,
		ReplicaId: 2,

		Cset: []*ViewChange_C{
			{
				SequenceNumber:60,
				Id:"chkpoint---digest60",
			},
		},

		Pset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
			{
				SequenceNumber:77,
				BatchDigest:"batch---77",
				View:0,
			},
			{
				SequenceNumber:79,
				BatchDigest:"batch---79",
				View:0,
			},
		},

		Qset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:67,
				BatchDigest:"batch---67",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
		},

	}

	vc3 := &ViewChange{
		View: 2,
		H: 50,
		ReplicaId: 3,

		Cset: []*ViewChange_C{
			{
				SequenceNumber:60,
				Id:"chkpoint---digest60",
			},
			{
				SequenceNumber:70,
				Id:"chkpoint---digest70",
			},
		},

		Pset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
			{
				SequenceNumber:80,
				BatchDigest:"batch---80",
				View:0,
			},
		},

		Qset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:71,
				BatchDigest:"batch---71",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
			{
				SequenceNumber:77,
				BatchDigest:"batch---77",
				View:0,
			},
		},

	}

	vc4 := &ViewChange{
		View: 2,
		H: 50,
		ReplicaId: 4,

		Cset: []*ViewChange_C{
			{
				SequenceNumber:60,
				Id:"chkpoint---digest60",
			},
		},

		Pset: []*ViewChange_PQ{
			{
				SequenceNumber:66,
				BatchDigest:"batch---66",
				View:0,
			},
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
			{
				SequenceNumber:85,
				BatchDigest:"batch---85",
				View:0,
			},
			{
				SequenceNumber:87,
				BatchDigest:"batch---87",
				View:0,
			},
		},

		Qset: []*ViewChange_PQ{
			{
				SequenceNumber:75,
				BatchDigest:"batch---75",
				View:0,
			},
			{
				SequenceNumber:77,
				BatchDigest:"batch---77",
				View:0,
			},
			{
				SequenceNumber:79,
				BatchDigest:"batch---79",
				View:0,
			},
		},

	}

	bool1 := pbft1.correctViewChange(vc1)
	bool2 := pbft1.correctViewChange(vc2)
	bool3 := pbft1.correctViewChange(vc3)
	bool4 := pbft1.correctViewChange(vc4)

	if !(bool1 == true && bool2 == true &&bool3 == true &&bool4 == true){
		t.Errorf("Should be 4 correct View-Change.and bllo1 is %v, bool2 is %v, bool3 is %v, bool4 is %v", bool1,bool2,bool3,bool4)
	}

	vset := []*ViewChange{vc1,vc2,vc3,vc4}

	cp,ok,replicas := pbft1.selectInitialCheckpoint(vset)
	if !(ok == true && reflect.DeepEqual(cp, ViewChange_C{SequenceNumber:60, Id:"chkpoint---digest60"}) && reflect.DeepEqual(replicas, []uint64{1,2,3,4})){
		t.Error("should get initial checkpoint!")
	}

	msglist := pbft1.assignSequenceNumbers(vset, cp.SequenceNumber)
	if !(msglist[66] == vc1.Pset[0].BatchDigest){
		t.Error("sequence number 66 should be assigned but not assigned actually")
	}

	//TODO: sequence number 77 should not be assigned!!!

}

func TestCanExecuteToTarget(t *testing.T){
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := "../../config/pbft.yaml"
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft1 := newPBFT(uint64(id), config, h)

	lastexec := uint64(25)
	initialcp := ViewChange_C{SequenceNumber: 30, Id: "vc---30"}

	canexe := pbft1.canExecuteToTarget(lastexec, initialcp)
	if !(canexe == false){
		t.Error("26 should not be executed because there is less than 2f+1 replicas committesd this request. ")
	}

	pbft1.certStore = map[msgID]*msgCert{
		msgID{v:0, n:26}: {
			commit: map[Commit]bool{
				Commit{View:0, SequenceNumber:26, ReplicaId:1}: true,
				Commit{View:0, SequenceNumber:26, ReplicaId:2}: true,
				Commit{View:0, SequenceNumber:26, ReplicaId:3}: true,
				Commit{View:0, SequenceNumber:26, ReplicaId:4}: true,
			},
		},
	}

	canexe = pbft1.canExecuteToTarget(lastexec, initialcp)
	if !(canexe == false){
		t.Error("27 should not be executed because there is less than 2f+1 replicas committesd this request. ")
	}

	pbft1.certStore[msgID{v:0, n:27}] = &msgCert{
		commit: map[Commit]bool{
			Commit{View:0, SequenceNumber:27, ReplicaId:1}: true,
			Commit{View:0, SequenceNumber:27, ReplicaId:2}: true,
			Commit{View:0, SequenceNumber:27, ReplicaId:3}: true,
		},
	}

	canexe = pbft1.canExecuteToTarget(lastexec, initialcp)
	if !(canexe == false){
		t.Error("28 should not be executed because there is less than 2f+1 replicas committesd this request. ")
	}

	pbft1.certStore[msgID{v:0, n:28}] = &msgCert{
		commit: map[Commit]bool{
			Commit{View:0, SequenceNumber:28, ReplicaId:1}: true,
			Commit{View:0, SequenceNumber:28, ReplicaId:2}: true,
			Commit{View:0, SequenceNumber:28, ReplicaId:3}: true,
		},
	}

	canexe = pbft1.canExecuteToTarget(lastexec, initialcp)
	if !(canexe == false){
		t.Error("29 should not be executed because there is less than 2f+1 replicas committesd this request. ")
	}

	pbft1.certStore[msgID{v:0, n:29}] = &msgCert{
		commit: map[Commit]bool{
			Commit{View:0, SequenceNumber:29, ReplicaId:1}: true,
			Commit{View:0, SequenceNumber:29, ReplicaId:2}: true,
			Commit{View:0, SequenceNumber:29, ReplicaId:3}: true,
		},
	}

	canexe = pbft1.canExecuteToTarget(lastexec, initialcp)
	if !(canexe == false){
		t.Error("30 should not be executed because there is less than 2f+1 replicas committesd this request. ")
	}

	pbft1.certStore[msgID{v:0, n:30}] = &msgCert{
		commit: map[Commit]bool{
			Commit{View:0, SequenceNumber:30, ReplicaId:1}: true,
			Commit{View:0, SequenceNumber:30, ReplicaId:2}: true,
			Commit{View:0, SequenceNumber:30, ReplicaId:3}: true,
		},
	}

	canexe = pbft1.canExecuteToTarget(lastexec, initialcp)
	if !(canexe == true && pbft1.nvInitialSeqNo == 30){
		t.Error("New-View sequence number should be 30 because we can execute to 30.")
	}
}

func TestProcessReqInNewView(t *testing.T){
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := "../../config/pbft.yaml"
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPBFT(uint64(id), config, h)

	vc1 := getTestViewChange()
	vc1.ReplicaId = 1
	vc2 := getTestViewChange()
	vc2.ReplicaId = 2
	vc3 := getTestViewChange()
	vc3.ReplicaId = 3
	vc4 := getTestViewChange()
	vc4.ReplicaId = 4

	nv := &NewView{
		View: 1,
		Vset: []*ViewChange{vc1,vc2,vc3,vc4},
		Xset: map[uint64]string{
			61: "chkponit---digest60",
			62: "chkponit---digest70",
		},
		ReplicaId: 2,
	}

	pbft.h = 60
	pbft.validatedBatchStore["chkponit---digest60"] = &TransactionBatch{}
	pbft.validatedBatchStore["chkponit---digest70"] = &TransactionBatch{}
	pbft.processReqInNewView(nv)

}