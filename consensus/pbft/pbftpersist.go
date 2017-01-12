//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"fmt"
	"encoding/base64"
	"encoding/binary"

	"hyperchain/consensus/helper/persist"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func (pbft *pbftProtocal) persistQSet(v uint64, n uint64) {

	cert := pbft.getCert(v, n)
	if pbft.qset == nil {
		qset := []*PrePrepare{}
		pbft.qset = &Qset{Set: qset}
	}
	pbft.qset.Set = append(pbft.qset.Set, cert.prePrepare)
	raw, err := proto.Marshal(pbft.qset)
	if err != nil {
		logger.Warningf("Replica %d could not persist qset: %s", pbft.id, err)
		return
	}
	persist.StoreState("qset", raw)
}

func (pbft *pbftProtocal) persistPSet(v uint64, n uint64) {

	cert := pbft.getCert(v, n)
	if pbft.pset == nil {
		pset := []*Prepare{}
		pbft.pset = &Pset{Set: pset}
	}
	for p := range cert.prepare {
		pbft.pset.Set = append(pbft.pset.Set, &p)
	}

	raw, err := proto.Marshal(pbft.pset)
	if err != nil {
		logger.Warningf("Replica %d could not persist pset: %s", pbft.id, err)
		return
	}
	persist.StoreState("pset", raw)
}

func (pbft *pbftProtocal) persistCSet(v uint64, n uint64) {

	cert := pbft.getCert(v, n)
	if pbft.cset == nil {
		cset := []*Commit{}
		pbft.cset = &Cset{Set: cset}
	}
	for c := range cert.commit {
		pbft.cset.Set = append(pbft.cset.Set, &c)
	}

	raw, err := proto.Marshal(pbft.cset)
	if err != nil {
		logger.Warningf("Replica %d could not persist cset: %s", pbft.id, err)
		return
	}
	persist.StoreState("cset", raw)
}

func (pbft *pbftProtocal) restoreQSet() *Qset {
	raw, err := persist.ReadState("qset")
	if err != nil {
		logger.Debugf("Replica %d could not restore state qset: %s", pbft.id, err)
		return nil
	}
	qset := &Qset{}
	err = proto.Unmarshal(raw, qset)
	if err != nil {
		logger.Errorf("Replica %d could not unmarshal qset - local state is damaged: %s", pbft.id, err)
		return nil
	}
	return qset
}

func (pbft *pbftProtocal) restorePSet() *Pset {
	raw, err := persist.ReadState("pset")
	if err != nil {
		logger.Debugf("Replica %d could not restore state pset: %s", pbft.id, err)
		return nil
	}
	pset := &Pset{}
	err = proto.Unmarshal(raw, pset)
	if err != nil {
		logger.Errorf("Replica %d could not unmarshal pset - local state is damaged: %s", pbft.id, err)
		return nil
	}
	return pset
}

func (pbft *pbftProtocal) restoreCSet() *Cset {
	raw, err := persist.ReadState("pset")
	if err != nil {
		logger.Debugf("Replica %d could not restore state cset: %s", pbft.id, err)
		return nil
	}
	cset := &Cset{}
	err = proto.Unmarshal(raw, cset)
	if err != nil {
		logger.Errorf("Replica %d could not unmarshal cset - local state is damaged: %s", pbft.id, err)
		return nil
	}
	return cset
}

func (pbft *pbftProtocal) restoreCert() {

	qset := []*PrePrepare{}
	pset := []*Prepare{}
	cset := []*Commit{}

	if pbft.qset != nil {
		qset = pbft.qset.Set
	}
	if pbft.pset != nil {
		pset = pbft.pset.Set
	}
	if pbft.cset != nil {
		cset = pbft.cset.Set
	}

	if len(qset) == 0 {
		return
	}

	for _, q := range qset {
		cert := pbft.getCert(q.View, q.SequenceNumber)
		cert.prePrepare = q
		cert.digest = q.BatchDigest

		pbft.validatedBatchStore[cert.digest] = q.GetTransactionBatch()

		if len(pset) == 0 {
			continue
		}

		pcount := 0
		ptmp := []*Prepare{}
		for _, p := range pset {
			if q.View == p.View && q.SequenceNumber == p.SequenceNumber {
				cert.prepare[*p] = true
				pcount++
			} else {
				ptmp = append(ptmp, p)
			}
		}
		cert.prepareCount = pcount
		pset = ptmp

		if len(cset) == 0 {
			continue
		}

		ccount := 0
		ctmp := []*Commit{}
		for _, c := range cset {
			if q.View == c.View && q.SequenceNumber == c.SequenceNumber {
				cert.commit[*c] = true
				ccount++
			} else {
				ctmp = append(ctmp, c)
			}
		}
		cert.commitCount = ccount
		cset = ctmp

	}

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

func (pbft *pbftProtocal) persistView(view uint64) {
	key := fmt.Sprintf("view")
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, view)
	persist.StoreState(key, b)
}

func (pbft *pbftProtocal) persistDelView() {
	key := fmt.Sprintf("view")
	persist.DelState(key)
}

func (pbft *pbftProtocal) restoreState() {

	pbft.qset = pbft.restoreQSet()
	pbft.pset = pbft.restorePSet()
	pbft.cset = pbft.restoreCSet()

	pbft.restoreCert()

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

	b, err := persist.ReadState("view")
	if err == nil {
		view := binary.LittleEndian.Uint64(b)
		pbft.view = view
		logger.Noticef("=========restore view %d=======", view)
	} else {
		logger.Noticef("Replica %d could not restore view: %s", pbft.id, err)
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
	h := persist.GetHeightofChain()
	if h == 0 {
		err = errors.Errorf("Height of chain is 0")
		return h, err
	}

	return h, nil
}
