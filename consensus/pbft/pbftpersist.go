package pbft

import (
	"fmt"
	"hyperchain/consensus/helper/persist"

	"github.com/golang/protobuf/proto"
	"encoding/base64"
	"github.com/pkg/errors"
	//"reflect"
)

func (pbft *pbftProtocal) persistQSet() {
	var qset []*ViewChange_PQ

	for _, q := range pbft.calcQSet() {
		qset = append(qset, q)
	}

	pbft.persistPQSet("qset", qset)
}

func (pbft *pbftProtocal) persistPSet() {
	var pset []*ViewChange_PQ

	for _, p := range pbft.calcPSet() {
		pset = append(pset, p)
	}

	pbft.persistPQSet("pset", pset)
}

func (pbft *pbftProtocal) persistPQSet(key string, set []*ViewChange_PQ) {
	raw, err := proto.Marshal(&PQset{set})
	if err != nil {
		logger.Warningf("Replica %d could not persist pqset: %s", pbft.id, err)
		return
	}
	persist.StoreState(key, raw)
}

//func (pbft *pbftProtocal) persistDelPSet(n uint64) {
//	raw, err := persist.ReadState("pset")
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
//			var newPset []*ViewChange_PQ
//			for _, p := range pset {
//				if p.SequenceNumber == n {
//					continue
//				}
//				newPset = append(newPset, p)
//			}
//			pbft.persistPQSet("pset", newPset)
//		}
//	}
//}
//
//func (pbft *pbftProtocal) persistDelQSet(idx qidx) {
//	raw, err := persist.ReadState("qset")
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
//			var newQset []*ViewChange_PQ
//			for _, q := range qset {
//				if idx.d == q.BatchDigest && idx.n == q.SequenceNumber {
//					continue
//				}
//				newQset = append(newQset, q)
//			}
//			pbft.persistPQSet("qset", newQset)
//		}
//	}
//}

func (pbft *pbftProtocal) restorePQSet(key string) []*ViewChange_PQ {
	raw, err := persist.ReadState(key)
	if err != nil {
		logger.Debugf("Replica %d could not restore state %s: %s", pbft.id, key, err)
		return nil
	}
	val := &PQset{}
	err = proto.Unmarshal(raw, val)
	if err != nil {
		logger.Errorf("Replica %d could not unmarshal %s - local state is damaged: %s", pbft.id, key, err)
		return nil
	}
	return val.GetSet()
}

func (pbft *pbftProtocal) persistRequestBatch(digest string) {
	reqBatch := pbft.validatedBatchStore[digest]
	reqBatchPacked, err := proto.Marshal(reqBatch)
	if err != nil {
		logger.Warningf("Replica %d could not persist request batch %s: %s", pbft.id, digest, err)
		return
	}
	persist.StoreState("reqBatch."+digest, reqBatchPacked)
}

func (pbft *pbftProtocal) persistDelRequestBatch(digest string) {
	persist.DelState("reqBatch."+digest)
}

func (pbft *pbftProtocal) persistDelAllRequestBatches() {
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

func (pbft *pbftProtocal) persistCheckpoint(seqNo uint64, id []byte) {
	key := fmt.Sprintf("chkpt.%d", seqNo)
	persist.StoreState(key, id)
}

func (pbft *pbftProtocal) persistDelCheckpoint(seqNo uint64) {
	key := fmt.Sprintf("chkpt.%d", seqNo)
	persist.DelState(key)
}

func (pbft *pbftProtocal) restoreState() {
	updateSeqView := func(set []*ViewChange_PQ) {
		for _, e := range set {
			if pbft.view < e.View {
				pbft.view = e.View
			}
			if pbft.seqNo < e.SequenceNumber {
				pbft.seqNo = e.SequenceNumber
			}
		}
	}

	set := pbft.restorePQSet("pset")
	for _, e := range set {
		pbft.pset[e.SequenceNumber] = e
	}
	updateSeqView(set)

	set = pbft.restorePQSet("qset")
	for _, e := range set {
		pbft.qset[qidx{e.BatchDigest, e.SequenceNumber}] = e
	}
	updateSeqView(set)

	reqBatchesPacked, err := persist.ReadStateSet("reqBatch.")
	if err == nil {
		for k, v := range reqBatchesPacked {
			reqBatch := &TransactionBatch{}
			err = proto.Unmarshal(v, reqBatch)
			if err != nil {
				logger.Warningf("Replica %d could not restore request batch %s", pbft.id, k)
			} else {
				pbft.validatedBatchStore[hash(reqBatch)] = reqBatch
			}
		}
	} else {
		logger.Warningf("Replica %d could not restore reqBatchStore: %s", pbft.id, err)
	}

	chkpts, err := persist.ReadStateSet("chkpt.")
	if err == nil {
		highSeq := uint64(0)
		for key, id := range chkpts {
			var seqNo uint64
			if _, err = fmt.Sscanf(key, "chkpt.%d", &seqNo); err != nil {
				logger.Warningf("Replica %d could not restore checkpoint key %s", pbft.id, key)
			} else {
				idAsString := base64.StdEncoding.EncodeToString(id)
				logger.Debugf("Replica %d found checkpoint %s for seqNo %d", pbft.id, idAsString, seqNo)
				pbft.chkpts[seqNo] = idAsString
				if seqNo > highSeq {
					highSeq = seqNo
				}
			}
		}
		pbft.moveWatermarks(highSeq)
	} else {
		logger.Warningf("Replica %d could not restore checkpoints: %s", pbft.id, err)
	}

	pbft.restoreLastSeqNo() // assign value to lastExec
	if pbft.seqNo < pbft.lastExec {
		pbft.seqNo = pbft.lastExec
	}
	pbft.vid, pbft.lastVid = pbft.seqNo
	logger.Infof("Replica %d restored state: view: %d, seqNo: %d, pset: %d, qset: %d, reqBatches: %d, chkpts: %d",
		pbft.id, pbft.view, pbft.seqNo, len(pbft.pset), len(pbft.qset), len(pbft.validatedBatchStore), len(pbft.chkpts))
}

func (pbft *pbftProtocal) restoreLastSeqNo() {
	var err error
	if pbft.lastExec, err = pbft.getLastSeqNo(); err != nil {
		logger.Warningf("Replica %d could not restore lastExec: %s", pbft.id, err)
		pbft.lastExec = 0
	}
	logger.Infof("Replica %d restored lastExec: %d", pbft.id, pbft.lastExec)
}

func (pbft *pbftProtocal) getLastSeqNo() (uint64, error) {

	var err error
	h := persist.GetHeightofChain()
	if h == 0 {
		err = errors.Errorf("Height of chain is 0")
		return h, err
	}

	return h, nil
}

