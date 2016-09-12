package pbft

import (
	"fmt"
	"hyperchain/consensus/helper/persist"

	"github.com/golang/protobuf/proto"
	"encoding/base64"
	"github.com/pkg/errors"
	//"reflect"
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

//func (instance *pbftCore) persistDelPSet(n uint64) {
//	raw, err := persist.ReadStateSet("pset")
//
//	if err != nil {
//		logger.Errorf("Read State Set Error %s", err)
//		return
//	} else {
//		pqset := &PQset{}
//		if umErr := proto.Unmarshal(raw, pqset); umErr != nil {
//			logger.Error(umErr)
//			return
//		} else {
//			pset := pqset.GetSet()
//			for key := range pset {
//				if key == n {
//					delete(pset, key)
//				}
//			}
//			instance.persistPQSet("pset", pset)
//		}
//	}
//}
//
//func (instance *pbftCore) persistDelQSet(idx qidx) {
//	raw, err := persist.ReadStateSet("qset")
//
//	if err!= nil {
//		logger.Errorf("Read State Set Error %s", err)
//		return
//	} else {
//		pqset := &PQset{}
//		if umErr := proto.Unmarshal(raw, pqset); umErr != nil {
//			logger.Error(umErr)
//			return
//		} else {
//			qset := pqset.GetSet()
//			for key := range qset {
//				if reflect.DeepEqual(key, idx) {
//					delete(qset, key)
//				}
//			}
//			instance.persistPQSet("qset", qset)
//		}
//	}
//}

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

func (instance *pbftCore) persistDelAllRequestBatches() {
	reqBatches, err := persist.ReadStateSet("reqBatch.")
	if err != nil {
		logger.Errorf("Read State Set Error %s", err)
		return
	} else {
		for k := range reqBatches {
			persist.DelState(k)
		}
	}
}

func (instance *pbftCore) persistCheckpoint(seqNo uint64, id []byte) {
	key := fmt.Sprintf("chkpt.%d", seqNo)
	persist.StoreState(key, id)
}

func (instance *pbftCore) persistDelCheckpoint(seqNo uint64) {
	key := fmt.Sprintf("chkpt.%d", seqNo)
	persist.DelState(key)
}

func (instance *pbftCore) restoreState() {
	updateSeqView := func(set []*ViewChange_PQ) {
		for _, e := range set {
			if instance.view < e.View {
				instance.view = e.View
			}
			if instance.seqNo < e.SequenceNumber {
				instance.seqNo = e.SequenceNumber
			}
		}
	}

	set := instance.restorePQSet("pset")
	for _, e := range set {
		instance.pset[e.SequenceNumber] = e
	}
	updateSeqView(set)

	set = instance.restorePQSet("qset")
	for _, e := range set {
		instance.qset[qidx{e.BatchDigest, e.SequenceNumber}] = e
	}
	updateSeqView(set)

	reqBatchesPacked, err := persist.ReadStateSet("reqBatch.")
	if err == nil {
		for k, v := range reqBatchesPacked {
			reqBatch := &RequestBatch{}
			err = proto.Unmarshal(v, reqBatch)
			if err != nil {
				logger.Warningf("Replica %d could not restore request batch %s", instance.id, k)
			} else {
				instance.reqBatchStore[hash(reqBatch)] = reqBatch
			}
		}
	} else {
		logger.Warningf("Replica %d could not restore reqBatchStore: %s", instance.id, err)
	}

	chkpts, err := persist.ReadStateSet("chkpt.")
	if err == nil {
		highSeq := uint64(0)
		for key, id := range chkpts {
			var seqNo uint64
			if _, err = fmt.Sscanf(key, "chkpt.%d", &seqNo); err != nil {
				logger.Warningf("Replica %d could not restore checkpoint key %s", instance.id, key)
			} else {
				idAsString := base64.StdEncoding.EncodeToString(id)
				logger.Debugf("Replica %d found checkpoint %s for seqNo %d", instance.id, idAsString, seqNo)
				instance.chkpts[seqNo] = idAsString
				if seqNo > highSeq {
					highSeq = seqNo
				}
			}
		}
		instance.moveWatermarks(highSeq)
	} else {
		logger.Warningf("Replica %d could not restore checkpoints: %s", instance.id, err)
	}

	instance.restoreLastSeqNo() // assign value to lastExec
	if instance.seqNo < instance.lastExec {
		instance.seqNo = instance.lastExec
	}
	logger.Infof("Replica %d restored state: view: %d, seqNo: %d, pset: %d, qset: %d, reqBatches: %d, chkpts: %d",
		instance.id, instance.view, instance.seqNo, len(instance.pset), len(instance.qset), len(instance.reqBatchStore), len(instance.chkpts))
}

func (instance *pbftCore) restoreLastSeqNo() {
	var err error
	if instance.lastExec, err = instance.getLastSeqNo(); err != nil {
		logger.Warningf("Replica %d could not restore lastExec: %s", instance.id, err)
		instance.lastExec = 0
	}
	logger.Infof("Replica %d restored lastExec: %d", instance.id, instance.lastExec)
}

func (instance *pbftCore) getLastSeqNo() (uint64, error) {

	var err error
	h := persist.GetHeightofChain()
	if h == 0 {
		err = errors.Errorf("Height of chain is 0")
		return h, err
	}

	return h, nil
}

