package pbft

import (
	"fmt"
	"hyperchain/consensus/helper/persist"

	"github.com/golang/protobuf/proto"
)

func (instance *pbftCore) persistQSet() {
	var qset []*ViewChange_PQ

	for _, q := range instance.calcQSet() {
		qset = append(qset, q)
	}

	instance.persistPQSet("qset", qset)
}

func (instance *pbftCore) persistPSet() {
	var pset []*ViewChange_PQ

	for _, p := range instance.calcPSet() {
		pset = append(pset, p)
	}

	instance.persistPQSet("pset", pset)
}

func (instance *pbftCore) persistPQSet(key string, set []*ViewChange_PQ) {
	raw, err := proto.Marshal(&PQset{set})
	if err != nil {
		logger.Warningf("Replica %d could not persist pqset: %s", instance.id, err)
		return
	}
	persist.StoreState(key, raw)
}

func (instance *pbftCore) restorePQSet(key string) []*ViewChange_PQ {
	raw, err := persist.ReadState(key)
	if err != nil {
		logger.Debugf("Replica %d could not restore state %s: %s", instance.id, key, err)
		return nil
	}
	val := &PQset{}
	err = proto.Unmarshal(raw, val)
	if err != nil {
		logger.Errorf("Replica %d could not unmarshal %s - local state is damaged: %s", instance.id, key, err)
		return nil
	}
	return val.GetSet()
}

func (instance *pbftCore) persistRequestBatch(digest string) {
	reqBatch := instance.reqBatchStore[digest]
	reqBatchPacked, err := proto.Marshal(reqBatch)
	if err != nil {
		logger.Warningf("Replica %d could not persist request batch %s: %s", instance.id, digest, err)
		return
	}
	persist.StoreState("reqBatch."+digest, reqBatchPacked)
}

func (instance *pbftCore) persistDelRequestBatch(digest string) {
	persist.DelState("reqBatch."+digest)
}

//func (instance *pbftCore) persistDelAllRequestBatches() {
//	reqBatches, err := persist.ReadStateSet("reqBatch.")
//	if err != nil {
//		logger.Errorf("Read State Set Error %s", err)
//		return
//	} else {
//		for k := range reqBatches {
//			persist.DelState(k)
//		}
//	}
//}

func (instance *pbftCore) persistCheckpoint(seqNo uint64, id []byte) {
	key := fmt.Sprintf("chkpt.%d", seqNo)
	persist.StoreState(key, id)
}

func (instance *pbftCore) persistDelCheckpoint(seqNo uint64) {
	key := fmt.Sprintf("chkpt.%d", seqNo)
	persist.DelState(key)
}

func (instance *pbftCore) restoreState() {

}

func (instance *pbftCore) restoreLastSeqNo() {

}
