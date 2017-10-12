//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package rbft

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"

	"hyperchain/consensus/helper/persist"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"strconv"
)

// persistQSet persists marshaled pre-prepare message to database
func (rbft *rbftImpl) persistQSet(preprep *PrePrepare) {
	raw, err := proto.Marshal(preprep)
	if err != nil {
		rbft.logger.Warningf("Replica %d could not persist qset: %s", rbft.id, err)
		return
	}
	key := fmt.Sprintf("qset.%d.%d.%s", preprep.View, preprep.SequenceNumber, preprep.BatchDigest)
	persist.StoreState(rbft.namespace, key, raw)
}

// persistPSet persists marshaled prepare messages in the cert with the given msgID(v,n,d) to database
func (rbft *rbftImpl) persistPSet(v uint64, n uint64, d string) {
	cert := rbft.storeMgr.getCert(v, n, d)
	set := []*Prepare{}
	pset := &Pset{Set: set}
	for p := range cert.prepare {
		tmp := p
		pset.Set = append(pset.Set, &tmp)
	}

	raw, err := proto.Marshal(pset)
	if err != nil {
		rbft.logger.Warningf("Replica %d could not persist pset: %s", rbft.id, err)
		return
	}
	key := fmt.Sprintf("pset.%d.%d.%s", v, n, d)
	persist.StoreState(rbft.namespace, key, raw)
}

// persistCSet persists marshaled commit messages in the cert with the given msgID(v,n,d) to database
func (rbft *rbftImpl) persistCSet(v uint64, n uint64, d string) {
	cert := rbft.storeMgr.getCert(v, n, d)
	set := []*Commit{}
	cset := &Cset{Set: set}
	for c := range cert.commit {
		tmp := c
		cset.Set = append(cset.Set, &tmp)
	}

	raw, err := proto.Marshal(cset)
	if err != nil {
		rbft.logger.Warningf("Replica %d could not persist cset: %s", rbft.id, err)
		return
	}
	key := fmt.Sprintf("cset.%d.%d.%s", v, n, d)
	persist.StoreState(rbft.namespace, key, raw)
}

// persistDelQSet deletes marshaled pre-prepare message with the given key from database
func (rbft *rbftImpl) persistDelQSet(v uint64, n uint64, d string) {
	qset := fmt.Sprintf("qset.%d.%d.%s", v, n, d)
	persist.DelState(rbft.namespace, qset)
}

// persistDelPSet deletes marshaled prepare messages with the given key from database
func (rbft *rbftImpl) persistDelPSet(v uint64, n uint64, d string) {
	pset := fmt.Sprintf("pset.%d.%d.%s", v, n, d)
	persist.DelState(rbft.namespace, pset)
}

// persistDelCSet deletes marshaled commit messages with the given key from database
func (rbft *rbftImpl) persistDelCSet(v uint64, n uint64, d string) {
	cset := fmt.Sprintf("cset.%d.%d.%s", v, n, d)
	persist.DelState(rbft.namespace, cset)
}

// persistDelQPCSet deletes marshaled pre-prepare,prepare,commit messages with the given key from database
func (rbft *rbftImpl) persistDelQPCSet(v uint64, n uint64, d string) {
	rbft.persistDelQSet(v, n, d)
	rbft.persistDelPSet(v, n, d)
	rbft.persistDelCSet(v, n, d)
}

// restoreQSet restores pre-prepare messages from database, which, keyed by msgID
func (rbft *rbftImpl) restoreQSet() (map[msgID]*PrePrepare, error) {
	qset := make(map[msgID]*PrePrepare)

	payload, err := persist.ReadStateSet(rbft.namespace, "qset.")
	if err == nil {
		for key, set := range payload {
			var v, n uint64
			var d string
			if _, err = fmt.Sscanf(key, "qset.%d.%d.%s", &v, &n, &d); err != nil {
				rbft.logger.Warningf("Replica %d could not restore qset key %s", rbft.id, key)
			} else {
				preprep := &PrePrepare{}
				err := proto.Unmarshal(set, preprep)
				if err == nil {
					idx := msgID{v, n, d}
					qset[idx] = preprep
				} else {
					rbft.logger.Warningf("Replica %d could not restore pre-prepare key %v, err: %v", rbft.id, set, err)
				}
			}
		}
	} else {
		rbft.logger.Warningf("Replica %d could not restore qset: %s", rbft.id, err)
	}

	return qset, err
}

// restorePSet restores prepare messages from database, which, keyed by msgID
func (rbft *rbftImpl) restorePSet() (map[msgID]*Pset, error) {
	pset := make(map[msgID]*Pset)

	payload, err := persist.ReadStateSet(rbft.namespace, "pset.")
	if err == nil {
		for key, set := range payload {
			var v, n uint64
			var d string
			if _, err = fmt.Sscanf(key, "pset.%d.%d.%s", &v, &n, &d); err != nil {
				rbft.logger.Warningf("Replica %d could not restore pset key %s", rbft.id, key)
			} else {
				prepares := &Pset{}
				err := proto.Unmarshal(set, prepares)
				if err == nil {
					idx := msgID{v, n, d}
					pset[idx] = prepares
				} else {
					rbft.logger.Warningf("Replica %d could not restore prepares key %v", rbft.id, set)
				}
			}
		}
	} else {
		rbft.logger.Warningf("Replica %d could not restore pset: %s", rbft.id, err)
	}

	return pset, err
}

// restoreCSet restores commit messages from database, which, keyed by msgID
func (rbft *rbftImpl) restoreCSet() (map[msgID]*Cset, error) {
	cset := make(map[msgID]*Cset)

	payload, err := persist.ReadStateSet(rbft.namespace, "cset.")
	if err == nil {
		for key, set := range payload {
			var v, n uint64
			var d string
			if _, err = fmt.Sscanf(key, "cset.%d.%d.%s", &v, &n, &d); err != nil {
				rbft.logger.Warningf("Replica %d could not restore pset key %s", rbft.id, key)
			} else {
				commits := &Cset{}
				err := proto.Unmarshal(set, commits)
				if err == nil {
					idx := msgID{v, n, d}
					cset[idx] = commits
				} else {
					rbft.logger.Warningf("Replica %d could not restore commits key %v", rbft.id, set)
				}
			}
		}
	} else {
		rbft.logger.Warningf("Replica %d could not restore cset: %s", rbft.id, err)
	}

	return cset, err
}

// restoreCert restores pre-prepares,prepares,commits from database and remove the messages with seqNo>lastExec
func (rbft *rbftImpl) restoreCert() {
	qset, _ := rbft.restoreQSet()
	for idx, q := range qset {
		cert := rbft.storeMgr.getCert(idx.v, idx.n, idx.d)
		if idx.n > rbft.exec.lastExec {
			rbft.persistDelQSet(idx.v, idx.n, idx.d)
			continue
		}
		cert.prePrepare = q
		cert.resultHash = q.ResultHash
	}

	pset, _ := rbft.restorePSet()
	for idx, prepares := range pset {
		if idx.n > rbft.exec.lastExec {
			rbft.persistDelPSet(idx.v, idx.n, idx.d)
			continue
		}
		cert := rbft.storeMgr.getCert(idx.v, idx.n, idx.d)
		for _, p := range prepares.Set {
			cert.prepare[*p] = true
			if p.ReplicaId == rbft.id {
				cert.sentPrepare = true
			}
		}
	}

	cset, _ := rbft.restoreCSet()
	for idx, commits := range cset {
		if idx.n > rbft.exec.lastExec {
			rbft.persistDelCSet(idx.v, idx.n, idx.d)
			continue
		}
		cert := rbft.storeMgr.getCert(idx.v, idx.n, idx.d)
		for _, c := range commits.Set {
			cert.commit[*c] = true
			if c.ReplicaId == rbft.id {
				cert.sentValidate = true
				cert.validated = true
				cert.sentCommit = true
			}
		}
	}
	for idx, cert := range rbft.storeMgr.certStore {
		if idx.n <= rbft.exec.lastExec {
			cert.sentExecute = true
		}
	}

}

// parseSpecifyCertStore re-constructs certStore:
// 1. for messages with the same seqNo but different view in certStore, save only the cert with the largest seqNo and
// remove all the certs from memory and database
// 2. replace all view in certStore with rbft.view and persist the new constructed certStore
func (rbft *rbftImpl) parseSpecifyCertStore() {
	for midx, mcert := range rbft.storeMgr.certStore {
		idx := midx
		cert := mcert
		for nidx, ncert := range rbft.storeMgr.certStore {
			if midx.n == nidx.n {
				if midx.v <= nidx.v {
					idx = nidx
					cert = ncert
				}
				delete(rbft.storeMgr.certStore, nidx)
				rbft.persistDelQPCSet(nidx.v, nidx.n, nidx.d)
			}
		}
		if cert.prePrepare != nil {
			cert.prePrepare.View = rbft.view
			primary := rbft.primary(rbft.view)
			cert.prePrepare.ReplicaId = primary
			rbft.persistQSet(cert.prePrepare)
		}
		preps := make(map[Prepare]bool)
		for prep := range cert.prepare {
			prep.View = rbft.view
			preps[prep] = true
		}
		cert.prepare = preps
		cmts := make(map[Commit]bool)
		for cmt := range cert.commit {
			cmt.View = rbft.view
			cmts[cmt] = true
		}
		cert.commit = cmts
		idx.v = rbft.view
		rbft.storeMgr.certStore[idx] = cert
		rbft.persistPSet(idx.v, idx.n, idx.d)
		rbft.persistCSet(idx.v, idx.n, idx.d)
	}
}

// persistTxBatch persists one marshaled transaction batch with the given digest to database
func (rbft *rbftImpl) persistTxBatch(digest string) {
	txBatch := rbft.storeMgr.txBatchStore[digest]
	txBatchPacked, err := proto.Marshal(txBatch)
	if err != nil {
		rbft.logger.Warningf("Replica %d could not persist request batch %s: %s", rbft.id, digest, err)
		return
	}
	persist.StoreState(rbft.namespace, "txBatch."+digest, txBatchPacked)
}

// persistDelTxBatch removes one marshaled transaction batch with the given digest from database
func (rbft *rbftImpl) persistDelTxBatch(digest string) {
	persist.DelState(rbft.namespace, "txBatch."+digest)
}

// persistDelAllTxBatches removes all marshaled transaction batches from database
func (rbft *rbftImpl) persistDelAllTxBatches() {
	reqBatches, err := persist.ReadStateSet(rbft.namespace, "txBatch.")
	if err != nil {
		rbft.logger.Errorf("Read State Set Error %s", err)
		return
	} else {
		for k := range reqBatches {
			persist.DelState(rbft.namespace, k)
		}
	}
}

// persistCheckpoint persists checkpoint to database, which, key contains the seqNo of checkpoint, value is the
// checkpoint ID
func (rbft *rbftImpl) persistCheckpoint(seqNo uint64, id []byte) {
	key := fmt.Sprintf("chkpt.%d", seqNo)
	persist.StoreState(rbft.namespace, key, id)
}

// persistDelCheckpoint deletes checkpoint with the given seqNo from database
func (rbft *rbftImpl) persistDelCheckpoint(seqNo uint64) {
	key := fmt.Sprintf("chkpt.%d", seqNo)
	persist.DelState(rbft.namespace, key)
}

// persistView persists current view to database
func (rbft *rbftImpl) persistView(view uint64) {
	key := fmt.Sprint("view")
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, view)
	persist.StoreState(rbft.namespace, key, b)
}

// persistDelView deletes the view entries from database
func (rbft *rbftImpl) persistDelView() {
	key := fmt.Sprint("view")
	persist.DelState(rbft.namespace, key)
}

// persistN persists current N to database
func (rbft *rbftImpl) persistN(n int) {
	key := fmt.Sprint("nodes")
	res := make([]byte, 8)
	binary.LittleEndian.PutUint64(res, uint64(n))
	persist.StoreState(rbft.namespace, key, res)
}

// persistNewNode persists new node message to database
func (rbft *rbftImpl) persistNewNode(new uint64) {
	key := fmt.Sprint("new")
	res := make([]byte, 8)
	binary.LittleEndian.PutUint64(res, new)
	persist.StoreState(rbft.namespace, key, res)
}

// persistLocalKey persists hash of local key to database
func (rbft *rbftImpl) persistLocalKey(hash []byte) {
	key := fmt.Sprint("localkey")
	persist.StoreState(rbft.namespace, key, hash)
}

// persistDelLocal key deletes local key info from database
func (rbft *rbftImpl) persistDellLocalKey() {
	key := fmt.Sprint("localkey")
	persist.DelState(rbft.namespace, key)
}

// restoreView restores current view from database and then re-construct certStore
func (rbft *rbftImpl) restoreView() {
	v, err := persist.ReadState(rbft.namespace, "view")
	if err == nil {
		view := binary.LittleEndian.Uint64(v)
		rbft.view = view
		rbft.parseSpecifyCertStore()
		rbft.logger.Noticef("========= restore view %d =======", rbft.view)
	} else {
		rbft.logger.Noticef("Replica %d could not restore view: %s", rbft.id, err)
	}
}

// restoreTxBatchStore restores transaction batches from database
func (rbft *rbftImpl) restoreTxBatchStore() {
	payload, err := persist.ReadStateSet(rbft.namespace, "txBatch.")
	if err == nil {
		for key, set := range payload {
			var digest string
			if _, err = fmt.Sscanf(key, "txBatch.%s", &digest); err != nil {
				rbft.logger.Warningf("Replica %d could not restore pset key %s", rbft.id, key)
			} else {
				batch := &TransactionBatch{}
				err := proto.Unmarshal(set, batch)
				if err == nil {
					rbft.storeMgr.txBatchStore[digest] = batch
				} else {
					rbft.logger.Warningf("Replica %d could not unmarshal batch key %s for error: %v", rbft.id, key, err)
				}
			}
		}
	} else {
		rbft.logger.Warningf("Replica %d could not restore txBatch for error: %v", rbft.id, err)
	}
}

// restoreState restores lastExec, certStore, view, transaction batches, checkpoints, h and other add/del node related
// params from database
func (rbft *rbftImpl) restoreState() {
	rbft.restoreLastSeqNo()
	if rbft.seqNo < rbft.exec.lastExec {
		rbft.seqNo = rbft.exec.lastExec
	}
	rbft.batchVdr.setLastVid(rbft.seqNo)

	rbft.restoreCert()
	rbft.restoreView()

	rbft.restoreTxBatchStore()

	chkpts, err := persist.ReadStateSet(rbft.namespace, "chkpt.")
	if err == nil {
		for key, id := range chkpts {
			var seqNo uint64
			if _, err = fmt.Sscanf(key, "chkpt.%d", &seqNo); err != nil {
				rbft.logger.Warningf("Replica %d could not restore checkpoint key %s", rbft.id, key)
			} else {
				idAsString := base64.StdEncoding.EncodeToString(id)
				rbft.logger.Debugf("Replica %d found checkpoint %s for seqNo %d", rbft.id, idAsString, seqNo)
				rbft.storeMgr.saveCheckpoint(seqNo, idAsString)
			}
		}
	} else {
		rbft.logger.Warningf("Replica %d could not restore checkpoints: %s", rbft.id, err)
	}
	hstr, err := persist.ReadState(rbft.namespace, "rbft.h")
	if err != nil {
		rbft.logger.Warningf("Replica %d could not restore h: %s", rbft.id, err)
	} else {
		h, err := strconv.ParseUint(string(hstr), 10, 64)
		if err != nil {
			panic("transfer rbft.h from string to uint64 failed with err: " + err.Error())
		}
		rbft.moveWatermarks(h)
	}
	n, err := persist.ReadState(rbft.namespace, "nodes")
	if err == nil {
		nodes := binary.LittleEndian.Uint64(n)
		rbft.N = int(nodes)
		rbft.f = (rbft.N - 1) / 3
	}
	rbft.logger.Noticef("========= restore N=%d, f=%d =======", rbft.N, rbft.f)

	new, err := persist.ReadState(rbft.namespace, "new")
	if err == nil {
		newNode := binary.LittleEndian.Uint64(new)
		if newNode == 1 {
			rbft.status.activeState(&rbft.status.isNewNode)
		}
	}

	localKey, err := persist.ReadState(rbft.namespace, "localkey")
	if err == nil {
		rbft.nodeMgr.localKey = string(localKey)
	}

	rbft.logger.Infof("Replica %d restored state: view: %d, seqNo: %d, reqBatches: %d, chkpts: %d",
		rbft.id, rbft.view, rbft.seqNo, len(rbft.storeMgr.txBatchStore), len(rbft.storeMgr.chkpts))
}

// restoreLastSeqNo restores lastExec from database
func (rbft *rbftImpl) restoreLastSeqNo() {
	var err error
	if rbft.exec.lastExec, err = rbft.getLastSeqNo(); err != nil {
		rbft.logger.Warningf("Replica %d could not restore lastExec: %s", rbft.id, err)
		rbft.exec.lastExec = 0
	}
	rbft.logger.Infof("Replica %d restored lastExec: %d", rbft.id, rbft.exec.lastExec)
}

// getLastSeqNo retrieves database and returns the last block number
func (rbft *rbftImpl) getLastSeqNo() (uint64, error) {
	var err error
	h := persist.GetHeightOfChain(rbft.namespace)
	if h == 0 {
		err = errors.Errorf("Height of chain is 0")
		return h, err
	}

	return h, nil
}
