//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"

	"hyperchain/consensus/helper/persist"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func (pbft *pbftProtocal) persistQSet(preprep *PrePrepare) {

	raw, err := proto.Marshal(preprep)
	if err != nil {
		logger.Warningf("Replica %d could not persist qset: %s", pbft.id, err)
		return
	}
	key := fmt.Sprintf("qset.%d.%d", preprep.View, preprep.SequenceNumber)
	persist.StoreState(pbft.namespace, key, raw)
}

func (pbft *pbftProtocal) persistPSet(v uint64, n uint64) {

	cert := pbft.getCert(v, n)
	set := []*Prepare{}
	pset := &Pset{Set: set}
	for p := range cert.prepare {
		tmp := p
		pset.Set = append(pset.Set, &tmp)
	}

	raw, err := proto.Marshal(pset)
	if err != nil {
		logger.Warningf("Replica %d could not persist pset: %s", pbft.id, err)
		return
	}
	key := fmt.Sprintf("pset.%d.%d", v, n)
	persist.StoreState(pbft.namespace, key, raw)
}

func (pbft *pbftProtocal) persistCSet(v uint64, n uint64) {

	cert := pbft.getCert(v, n)
	set := []*Commit{}
	cset := &Cset{Set: set}
	for c := range cert.commit {
		tmp := c
		cset.Set = append(cset.Set, &tmp)
	}

	raw, err := proto.Marshal(cset)
	if err != nil {
		logger.Warningf("Replica %d could not persist cset: %s", pbft.id, err)
		return
	}
	key := fmt.Sprintf("cset.%d.%d", v, n)
	persist.StoreState(pbft.namespace, key, raw)
}

func (pbft *pbftProtocal) persistDelQPCSet(v uint64, n uint64) {
	qset := fmt.Sprintf("qset.%d.%d", v, n)
	persist.DelState(pbft.namespace, qset)
	pset := fmt.Sprintf("pset.%d.%d", v, n)
	persist.DelState(pbft.namespace, pset)
	cset := fmt.Sprintf("cset.%d.%d", v, n)
	persist.DelState(pbft.namespace, cset)
}

func (pbft *pbftProtocal) restoreQSet() (map[msgID]*PrePrepare, error) {

	qset := make(map[msgID]*PrePrepare)

	payload, err := persist.ReadStateSet(pbft.namespace, "qset.")
	if err == nil {
		for key, set := range payload {
			var v, n uint64
			if _, err = fmt.Sscanf(key, "qset.%d.%d", &v, &n); err != nil {
				logger.Warningf("Replica %d could not restore qset key %s", pbft.id, key)
			} else {
				preprep := &PrePrepare{}
				err := proto.Unmarshal(set, preprep)
				if err == nil {
					idx := msgID{v, n}
					qset[idx] = preprep
				} else {
					logger.Warningf("Replica %d could not restore pre-prepare key %v, err: %v", pbft.id, set, err)
				}
			}
		}
	} else {
		logger.Warningf("Replica %d could not restore qset: %s", pbft.id, err)
	}

	return qset, err
}

func (pbft *pbftProtocal) restorePSet() (map[msgID]*Pset, error) {

	pset := make(map[msgID]*Pset)

	payload, err := persist.ReadStateSet(pbft.namespace, "pset.")
	if err == nil {
		for key, set := range payload {
			var v, n uint64
			if _, err = fmt.Sscanf(key, "pset.%d.%d", &v, &n); err != nil {
				logger.Warningf("Replica %d could not restore pset key %s", pbft.id, key)
			} else {
				prepares := &Pset{}
				err := proto.Unmarshal(set, prepares)
				if err == nil {
					idx := msgID{v, n}
					pset[idx] = prepares
				} else {
					logger.Warningf("Replica %d could not restore prepares key %v", pbft.id, set)
				}
			}
		}
	} else {
		logger.Warningf("Replica %d could not restore pset: %s", pbft.id, err)
	}

	return pset, err
}

func (pbft *pbftProtocal) restoreCSet() (map[msgID]*Cset, error) {

	cset := make(map[msgID]*Cset)

	payload, err := persist.ReadStateSet(pbft.namespace, "cset.")
	if err == nil {
		for key, set := range payload {
			var v, n uint64
			if _, err = fmt.Sscanf(key, "cset.%d.%d", &v, &n); err != nil {
				logger.Warningf("Replica %d could not restore pset key %s", pbft.id, key)
			} else {
				commits := &Cset{}
				err := proto.Unmarshal(set, commits)
				if err == nil {
					idx := msgID{v, n}
					cset[idx] = commits
				} else {
					logger.Warningf("Replica %d could not restore commits key %v", pbft.id, set)
				}
			}
		}
	} else {
		logger.Warningf("Replica %d could not restore cset: %s", pbft.id, err)
	}

	return cset, err
}

func (pbft *pbftProtocal) restoreCert() {

	qset, _ := pbft.restoreQSet()
	for idx, q := range qset {
		cert := pbft.getCert(idx.v, idx.n)
		cert.prePrepare = q
		cert.digest = q.BatchDigest

		pbft.validatedBatchStore[cert.digest] = q.GetTransactionBatch()
	}

	pset, _ := pbft.restorePSet()
	for idx, prepares := range pset {
		cert := pbft.getCert(idx.v, idx.n)
		for _, p := range prepares.Set {
			cert.prepare[*p] = true
			if p.ReplicaId == pbft.id {
				cert.sentPrepare = true
			}
		}
		cert.prepareCount = len(cert.prepare)
	}

	cset, _ := pbft.restoreCSet()
	for idx, commits := range cset {
		cert := pbft.getCert(idx.v, idx.n)
		for _, c := range commits.Set {
			cert.commit[*c] = true
			if c.ReplicaId == pbft.id {
				cert.sentValidate = true
				cert.validated = true
				cert.sentCommit = true
			}
		}
		cert.commitCount = len(cert.commit)
	}
}

func (pbft *pbftProtocal) persistRequestBatch(digest string) {
	reqBatch := pbft.validatedBatchStore[digest]
	reqBatchPacked, err := proto.Marshal(reqBatch)
	if err != nil {
		logger.Warningf("Replica %d could not persist request batch %s: %s", pbft.id, digest, err)
		return
	}
	persist.StoreState(pbft.namespace, "reqBatch."+digest, reqBatchPacked)
}

func (pbft *pbftProtocal) persistDelRequestBatch(digest string) {
	persist.DelState(pbft.namespace, "reqBatch." + digest)
}

func (pbft *pbftProtocal) persistDelAllRequestBatches() {
	reqBatches, err := persist.ReadStateSet(pbft.namespace, "reqBatch.")
	if err != nil {
		logger.Errorf("Read State Set Error %s", err)
		return
	} else {
		for k := range reqBatches {
			persist.DelState(pbft.namespace, k)
		}
	}
}

func (pbft *pbftProtocal) persistCheckpoint(seqNo uint64, id []byte) {
	key := fmt.Sprintf("chkpt.%d", seqNo)
	persist.StoreState(pbft.namespace, key, id)
}

func (pbft *pbftProtocal) persistDelCheckpoint(seqNo uint64) {
	key := fmt.Sprintf("chkpt.%d", seqNo)
	persist.DelState(pbft.namespace, key)
}

func (pbft *pbftProtocal) persistView(view uint64) {
	key := fmt.Sprint("view")
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, view)
	persist.StoreState(pbft.namespace, key, b)
}

func (pbft *pbftProtocal) persistDelView() {
	key := fmt.Sprint("view")
	persist.DelState(pbft.namespace, key)
}

func (pbft *pbftProtocal) persistN(n int) {
	key := fmt.Sprint("nodes")
	res := make([]byte, 8)
	binary.LittleEndian.PutUint64(res, uint64(n))
	persist.StoreState(pbft.namespace, key, res)
}

func (pbft *pbftProtocal) persistNewNode(new uint64) {
	key := fmt.Sprint("new")
	res := make([]byte, 8)
	binary.LittleEndian.PutUint64(res, new)
	persist.StoreState(pbft.namespace, key, res)
}

func (pbft *pbftProtocal) persistLocalKey(hash []byte) {
	key := fmt.Sprint("localkey")
	persist.StoreState(pbft.namespace, key, hash)
}

func (pbft *pbftProtocal) persistDellLocalKey() {
	key := fmt.Sprint("localkey")
	persist.DelState(pbft.namespace, key)
}

func (pbft *pbftProtocal) restoreState() {

	pbft.restoreCert()

	chkpts, err := persist.ReadStateSet(pbft.namespace, "chkpt.")
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

	b, err := persist.ReadState(pbft.namespace, "view")
	if err == nil {
		view := binary.LittleEndian.Uint64(b)
		pbft.view = view
		logger.Noticef("========= restore view %d =======", view)
	} else {
		logger.Noticef("Replica %d could not restore view: %s", pbft.id, err)
	}

	n, err := persist.ReadState(pbft.namespace, "nodes")
	if err == nil {
		nodes := binary.LittleEndian.Uint64(n)
		pbft.N = int(nodes)
		pbft.f = (pbft.N - 1) / 3
	}
	logger.Noticef("========= restore N=%d, f=%d =======", pbft.N, pbft.f)

	new, err := persist.ReadState(pbft.namespace, "new")
	if err == nil {
		newNode := binary.LittleEndian.Uint64(new)
		if newNode == 1 {
			pbft.isNewNode = true
		}
	}

	localKey, err := persist.ReadState(pbft.namespace, "localkey")
	if err == nil {
		pbft.localKey = string(localKey)
	}

	pbft.restoreLastSeqNo() // assign value to lastExec
	if pbft.seqNo < pbft.lastExec {
		pbft.seqNo = pbft.lastExec
	}
	pbft.vid = pbft.seqNo
	pbft.lastVid = pbft.seqNo
	logger.Infof("Replica %d restored state: view: %d, seqNo: %d, reqBatches: %d, chkpts: %d",
		pbft.id, pbft.view, pbft.seqNo, len(pbft.validatedBatchStore), len(pbft.chkpts))
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
	h := persist.GetHeightofChain(pbft.namespace)
	if h == 0 {
		err = errors.Errorf("Height of chain is 0")
		return h, err
	}

	return h, nil
}
