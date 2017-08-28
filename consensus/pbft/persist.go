//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"fmt"
	"hyperchain/consensus/helper/persist"

	"encoding/base64"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func (pbft *pbftImpl) persistQSet(preprep *PrePrepare) {

	raw, err := proto.Marshal(preprep)
	if err != nil {
		pbft.logger.Warningf("Replica %d could not persist qset: %s", pbft.id, err)
		return
	}
	key := fmt.Sprintf("qset.%d.%d.%d.%s", preprep.View, preprep.SequenceNumber, preprep.Vid, preprep.BatchDigest)
	persist.StoreState(pbft.namespace, key, raw)
}

func (pbft *pbftImpl) persistPSet(v uint64, n uint64, d string) {

	cert := pbft.storeMgr.getCert(v, n, d)
	set := []*Prepare{}
	pset := &Pset{Set: set}
	for p := range cert.prepare {
		tmp := p
		pset.Set = append(pset.Set, &tmp)
	}

	raw, err := proto.Marshal(pset)
	if err != nil {
		pbft.logger.Warningf("Replica %d could not persist pset: %s", pbft.id, err)
		return
	}
	key := fmt.Sprintf("pset.%d.%d.%s", v, n, d)
	persist.StoreState(pbft.namespace, key, raw)
}

func (pbft *pbftImpl) persistCSet(v uint64, n uint64, d string) {

	cert := pbft.storeMgr.getCert(v, n, d)
	set := []*Commit{}
	cset := &Cset{Set: set}
	for c := range cert.commit {
		tmp := c
		cset.Set = append(cset.Set, &tmp)
	}

	raw, err := proto.Marshal(cset)
	if err != nil {
		pbft.logger.Warningf("Replica %d could not persist cset: %s", pbft.id, err)
		return
	}
	key := fmt.Sprintf("cset.%d.%d.%s", v, n, d)
	persist.StoreState(pbft.namespace, key, raw)
}

func (pbft *pbftImpl) persistDelQSet(v uint64, n uint64, vid uint64, d string) {
	qset := fmt.Sprintf("qset.%d.%d.%d.%s", v, n, vid, d)
	persist.DelState(pbft.namespace, qset)
}

func (pbft *pbftImpl) persistDelPSet(v uint64, n uint64, d string) {
	pset := fmt.Sprintf("pset.%d.%d.%s", v, n, d)
	persist.DelState(pbft.namespace, pset)
}

func (pbft *pbftImpl) persistDelCSet(v uint64, n uint64, d string) {
	cset := fmt.Sprintf("cset.%d.%d.%s", v, n, d)
	persist.DelState(pbft.namespace, cset)
}
func (pbft *pbftImpl) persistDelQPCSet(v uint64, n uint64, vid uint64, d string) {
	pbft.persistDelQSet(v, n, vid, d)
	pbft.persistDelPSet(v, n, d)
	pbft.persistDelCSet(v, n, d)
}

func (pbft *pbftImpl) restoreQSet() (map[msgID]*PrePrepare, error) {

	qset := make(map[msgID]*PrePrepare)

	payload, err := persist.ReadStateSet(pbft.namespace, "qset.")
	if err == nil {
		for key, set := range payload {
			var v, n, vid uint64
			var d string
			if _, err = fmt.Sscanf(key, "qset.%d.%d.%d.%s", &v, &n, &vid, &d); err != nil {
				pbft.logger.Warningf("Replica %d could not restore qset key %s", pbft.id, key)
			} else {
				preprep := &PrePrepare{}
				err := proto.Unmarshal(set, preprep)
				if err == nil {
					if n != uint64(0) {
						idx := msgID{v, n, d}
						qset[idx] = preprep
					} else {
						idx := msgID{v, vid, d}
						pbft.batchVdr.spNullRequest[idx] = preprep
					}
				} else {
					pbft.logger.Warningf("Replica %d could not restore pre-prepare key %v, err: %v", pbft.id, set, err)
				}
			}
		}
	} else {
		pbft.logger.Warningf("Replica %d could not restore qset: %s", pbft.id, err)
	}

	return qset, err
}

func (pbft *pbftImpl) restorePSet() (map[msgID]*Pset, error) {

	pset := make(map[msgID]*Pset)

	payload, err := persist.ReadStateSet(pbft.namespace, "pset.")
	if err == nil {
		for key, set := range payload {
			var v, n uint64
			var d string
			if _, err = fmt.Sscanf(key, "pset.%d.%d.%s", &v, &n, &d); err != nil {
				pbft.logger.Warningf("Replica %d could not restore pset key %s", pbft.id, key)
			} else {
				prepares := &Pset{}
				err := proto.Unmarshal(set, prepares)
				if err == nil {
					idx := msgID{v, n, d}
					pset[idx] = prepares
				} else {
					pbft.logger.Warningf("Replica %d could not restore prepares key %v", pbft.id, set)
				}
			}
		}
	} else {
		pbft.logger.Warningf("Replica %d could not restore pset: %s", pbft.id, err)
	}

	return pset, err
}

func (pbft *pbftImpl) restoreCSet() (map[msgID]*Cset, error) {

	cset := make(map[msgID]*Cset)

	payload, err := persist.ReadStateSet(pbft.namespace, "cset.")
	if err == nil {
		for key, set := range payload {
			var v, n uint64
			var d string
			if _, err = fmt.Sscanf(key, "cset.%d.%d.%s", &v, &n, &d); err != nil {
				pbft.logger.Warningf("Replica %d could not restore pset key %s", pbft.id, key)
			} else {
				commits := &Cset{}
				err := proto.Unmarshal(set, commits)
				if err == nil {
					idx := msgID{v, n, d}
					cset[idx] = commits
				} else {
					pbft.logger.Warningf("Replica %d could not restore commits key %v", pbft.id, set)
				}
			}
		}
	} else {
		pbft.logger.Warningf("Replica %d could not restore cset: %s", pbft.id, err)
	}

	return cset, err
}

func (pbft *pbftImpl) restoreCert() {

	qset, _ := pbft.restoreQSet()
	for idx, q := range qset {
		cert := pbft.storeMgr.getCert(idx.v, idx.n, idx.d)
		if idx.n > pbft.exec.lastExec {
			pbft.persistDelQSet(idx.v, idx.n, cert.vid, idx.d)
			continue
		}
		cert.prePrepare = q
		cert.vid = q.Vid
		cert.resultHash = q.ResultHash
	}

	pset, _ := pbft.restorePSet()
	for idx, prepares := range pset {
		if idx.n > pbft.exec.lastExec {
			pbft.persistDelPSet(idx.v, idx.n, idx.d)
			continue
		}
		cert := pbft.storeMgr.getCert(idx.v, idx.n, idx.d)
		for _, p := range prepares.Set {
			cert.prepare[*p] = true
			if p.ReplicaId == pbft.id {
				cert.sentPrepare = true
			}
		}
	}

	cset, _ := pbft.restoreCSet()
	for idx, commits := range cset {
		if idx.n > pbft.exec.lastExec {
			pbft.persistDelCSet(idx.v, idx.n, idx.d)
			continue
		}
		cert := pbft.storeMgr.getCert(idx.v, idx.n, idx.d)
		for _, c := range commits.Set {
			cert.commit[*c] = true
			if c.ReplicaId == pbft.id {
				cert.sentValidate = true
				cert.validated = true
				cert.sentCommit = true
			}
		}
	}
	for idx, cert := range pbft.storeMgr.certStore {
		if idx.n <= pbft.exec.lastExec {
			cert.sentExecute = true
		}
	}

}

func (pbft *pbftImpl) parseSpecifyCertStore() {
	for midx, mcert := range pbft.storeMgr.certStore {
		idx := midx
		cert := mcert
		for nidx, ncert := range pbft.storeMgr.certStore {
			if midx.n == nidx.n {
				if midx.v <= nidx.v {
					idx = nidx
					cert = ncert
				}
				delete(pbft.storeMgr.certStore, nidx)
				pbft.persistDelQPCSet(nidx.v, nidx.n, ncert.vid, nidx.d)
			}
		}
		if cert.prePrepare != nil {
			cert.prePrepare.View = pbft.view
			primary := pbft.primary(pbft.view)
			cert.prePrepare.ReplicaId = primary
			pbft.persistQSet(cert.prePrepare)
		}
		preps := make(map[Prepare]bool)
		for prep := range cert.prepare {
			prep.View = pbft.view
			preps[prep] = true
		}
		cert.prepare = preps
		cmts := make(map[Commit]bool)
		for cmt := range cert.commit {
			cmt.View = pbft.view
			cmts[cmt] = true
		}
		cert.commit = cmts
		idx.v = pbft.view
		pbft.storeMgr.certStore[idx] = cert
		pbft.persistPSet(idx.v, idx.n, idx.d)
		pbft.persistCSet(idx.v, idx.n, idx.d)

	}
}

func (pbft *pbftImpl) persistTxBatch(digest string) {
	txBatch := pbft.storeMgr.txBatchStore[digest]
	txBatchPacked, err := proto.Marshal(txBatch)
	if err != nil {
		pbft.logger.Warningf("Replica %d could not persist request batch %s: %s", pbft.id, digest, err)
		return
	}
	persist.StoreState(pbft.namespace, "txBatch."+digest, txBatchPacked)
}

func (pbft *pbftImpl) persistDelTxBatch(digest string) {
	persist.DelState(pbft.namespace, "txBatch."+digest)
}

func (pbft *pbftImpl) persistDelAllTxBatches() {
	reqBatches, err := persist.ReadStateSet(pbft.namespace, "txBatch.")
	if err != nil {
		pbft.logger.Errorf("Read State Set Error %s", err)
		return
	} else {
		for k := range reqBatches {
			persist.DelState(pbft.namespace, k)
		}
	}
}

func (pbft *pbftImpl) persistCheckpoint(seqNo uint64, id []byte) {
	key := fmt.Sprintf("chkpt.%d", seqNo)
	persist.StoreState(pbft.namespace, key, id)
}

func (pbft *pbftImpl) persistDelCheckpoint(seqNo uint64) {
	key := fmt.Sprintf("chkpt.%d", seqNo)
	persist.DelState(pbft.namespace, key)
}

func (pbft *pbftImpl) persistView(view uint64) {
	key := fmt.Sprint("view")
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, view)
	persist.StoreState(pbft.namespace, key, b)
}

func (pbft *pbftImpl) persistDelView() {
	key := fmt.Sprint("view")
	persist.DelState(pbft.namespace, key)
}

func (pbft *pbftImpl) persistN(n int) {
	key := fmt.Sprint("nodes")
	res := make([]byte, 8)
	binary.LittleEndian.PutUint64(res, uint64(n))
	persist.StoreState(pbft.namespace, key, res)
}

func (pbft *pbftImpl) persistNewNode(new uint64) {
	key := fmt.Sprint("new")
	res := make([]byte, 8)
	binary.LittleEndian.PutUint64(res, new)
	persist.StoreState(pbft.namespace, key, res)
}

func (pbft *pbftImpl) persistLocalKey(hash []byte) {
	key := fmt.Sprint("localkey")
	persist.StoreState(pbft.namespace, key, hash)
}

func (pbft *pbftImpl) persistDellLocalKey() {
	key := fmt.Sprint("localkey")
	persist.DelState(pbft.namespace, key)
}

func (pbft *pbftImpl) restoreView() {

	v, err := persist.ReadState(pbft.namespace, "view")
	if err == nil {
		view := binary.LittleEndian.Uint64(v)
		pbft.view = view
		pbft.parseSpecifyCertStore()
		pbft.logger.Noticef("========= restore view %d =======", pbft.view)
	} else {
		pbft.logger.Noticef("Replica %d could not restore view: %s", pbft.id, err)
	}
}

func (pbft *pbftImpl) restoreTxBatchStore() {

	payload, err := persist.ReadStateSet(pbft.namespace, "txBatch.")
	if err == nil {
		for key, set := range payload {
			var digest string
			if _, err = fmt.Sscanf(key, "txBatch.%s", &digest); err != nil {
				pbft.logger.Warningf("Replica %d could not restore pset key %s", pbft.id, key)
			} else {
				batch := &TransactionBatch{}
				err := proto.Unmarshal(set, batch)
				if err == nil {
					pbft.storeMgr.txBatchStore[digest] = batch
				} else {
					pbft.logger.Warningf("Replica %d could not unmarshal batch key %s for error: %v", pbft.id, key, err)
				}
			}
		}
	} else {
		pbft.logger.Warningf("Replica %d could not restore txBatch for error: %v", pbft.id, err)
	}
}

func (pbft *pbftImpl) restoreState() {

	pbft.restoreLastSeqNo() // assign value to lastExec
	if pbft.seqNo < pbft.exec.lastExec {
		pbft.seqNo = pbft.exec.lastExec
	}
	pbft.batchVdr.setVid(pbft.seqNo)
	pbft.batchVdr.setLastVid(pbft.seqNo)

	pbft.restoreCert()
	pbft.restoreView()

	pbft.restoreTxBatchStore()

	chkpts, err := persist.ReadStateSet(pbft.namespace, "chkpt.")
	if err == nil {
		highSeq := uint64(0)
		for key, id := range chkpts {
			var seqNo uint64
			if _, err = fmt.Sscanf(key, "chkpt.%d", &seqNo); err != nil {
				pbft.logger.Warningf("Replica %d could not restore checkpoint key %s", pbft.id, key)
			} else {
				idAsString := base64.StdEncoding.EncodeToString(id)
				pbft.logger.Debugf("Replica %d found checkpoint %s for seqNo %d", pbft.id, idAsString, seqNo)
				pbft.storeMgr.saveCheckpoint(seqNo, idAsString)
				if seqNo > highSeq {
					highSeq = seqNo
				}
			}
		}
		pbft.moveWatermarks(highSeq)
	} else {
		pbft.logger.Warningf("Replica %d could not restore checkpoints: %s", pbft.id, err)
	}

	n, err := persist.ReadState(pbft.namespace, "nodes")
	if err == nil {
		nodes := binary.LittleEndian.Uint64(n)
		pbft.N = int(nodes)
		pbft.f = (pbft.N - 1) / 3
	}
	pbft.logger.Noticef("========= restore N=%d, f=%d =======", pbft.N, pbft.f)

	new, err := persist.ReadState(pbft.namespace, "new")
	if err == nil {
		newNode := binary.LittleEndian.Uint64(new)
		if newNode == 1 {
			pbft.status.activeState(&pbft.status.isNewNode)
		}
	}

	localKey, err := persist.ReadState(pbft.namespace, "localkey")
	if err == nil {
		pbft.nodeMgr.localKey = string(localKey)
	}


	pbft.logger.Infof("Replica %d restored state: view: %d, seqNo: %d, reqBatches: %d, chkpts: %d",
		pbft.id, pbft.view, pbft.seqNo, len(pbft.storeMgr.txBatchStore), len(pbft.storeMgr.chkpts))
}

func (pbft *pbftImpl) restoreLastSeqNo() {
	var err error
	if pbft.exec.lastExec, err = pbft.getLastSeqNo(); err != nil {
		pbft.logger.Warningf("Replica %d could not restore lastExec: %s", pbft.id, err)
		pbft.exec.lastExec = 0
	}
	pbft.logger.Infof("Replica %d restored lastExec: %d", pbft.id, pbft.exec.lastExec)
}

func (pbft *pbftImpl) getLastSeqNo() (uint64, error) {

	var err error
	h := persist.GetHeightOfChain(pbft.namespace)
	if h == 0 {
		err = errors.Errorf("Height of chain is 0")
		return h, err
	}

	return h, nil
}
