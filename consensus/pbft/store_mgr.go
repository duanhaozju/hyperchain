//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

/**
	This file provide a mechanism to manage the storage in PBFT
 */

//storeManager manage common store data structures for PBFT.
type storeManager struct {

	chkpts          map[uint64]string  // state checkpoints; map lastExec to global hash
	hChkpts         map[uint64]uint64  // map repilicaId to replica highest checkpoint num
	checkpointStore map[Checkpoint]bool                      // track checkpoints as set
	chkptCertStore  map[chkptID]*chkptCert                   // track quorum certificates for checkpoints

	certStore       map[msgID]*msgCert                       // track quorum certificates for requests
	committedCert   map[msgID]string                         // track the committed cert to help execute msgId - digest

	outstandingReqBatches map[string]*TransactionBatch       // track whether we are waiting for request batches to execute

	missingReqBatches map[string]bool                        // for all the assigned, non-checkpointed request batches we might be missing during view-change
	highStateTarget   *stateUpdateTarget                     // Set to the highest weak checkpoint cert we have observed
}

func newStoreMgr() *storeManager  {
	sm := &storeManager{}

	sm.chkpts = make(map[uint64]string)
	sm.chkpts[0] = "XXX GENESIS"
	sm.hChkpts = make(map[uint64]uint64)
	sm.checkpointStore = make(map[Checkpoint]bool)
	sm.chkptCertStore = make(map[chkptID]*chkptCert)

	sm.certStore = make(map[msgID]*msgCert)
	sm.committedCert = make(map[msgID]string)

	sm.outstandingReqBatches = make(map[string]*TransactionBatch)

	sm.missingReqBatches = make(map[string]bool)
	return sm
}


//removeLowPQSet remove set in chpts, pset, qset, cset, plist, qlist which index <= h
func (sm *storeManager) moveWatermarks(pbft *pbftImpl, h uint64) {

	for n := range sm.chkpts {
		if n < h {
			delete(sm.chkpts, n)
			pbft.persistDelCheckpoint(n)
		}
	}

	for idx := range pbft.vcMgr.qlist {
		if idx.n <= h {
			delete(pbft.vcMgr.qlist, idx)
		}
	}

	for n := range pbft.vcMgr.plist {
		if n <= h {
			delete(pbft.vcMgr.plist, n)
		}
	}
}

func (sm *storeManager) chkptsLen() int {
	return len(sm.chkpts)
}

//saveCheckpoint lastExec - global hash
func (sm *storeManager) saveCheckpoint(l uint64, gh string)  {
	sm.chkpts[l] = gh
}

func (sm *storeManager) getCheckpoint(l uint64) string  {
	return sm.chkpts[l]
}

// getCert given a digest/view/seq, is there an entry in the certLog?
// If so, return it. If not, create it.
func (sm *storeManager) getCert(v uint64, n uint64) (cert *msgCert) {

	idx := msgID{v, n}
	cert, ok := sm.certStore[idx]

	if ok {
		return
	}

	prepare := make(map[Prepare]bool)
	commit := make(map[Commit]bool)
	cert = &msgCert{
		prepare: prepare,
		commit:	 commit,
	}
	sm.certStore[idx] = cert
	return
}

// getChkptCert given a seqNo/id get the checkpoint Cert.
func (sm *storeManager) getChkptCert(n uint64, id string) (cert *chkptCert) {

	idx := chkptID{n, id}
	cert, ok := sm.chkptCertStore[idx]

	if ok {
		return
	}

	chkpts := make(map[Checkpoint]bool)
	cert = &chkptCert{
		chkpts: chkpts,
	}
	sm.chkptCertStore[idx] = cert

	return
}

// check for other PRE-PREPARE with same digest, but different seqNo
func (sm*storeManager) existedDigest(n uint64, view uint64, digest string) bool {
	for _, cert := range sm.certStore {
		if p := cert.prePrepare; p != nil {
			if p.View == view && p.SequenceNumber != n && p.BatchDigest == digest && digest != "" {
				// This will happen if primary receive same digest result of txs
				// It may result in DDos attack
				logger.Warningf("Other pre-prepare found with same digest but different seqNo: %d " +
					"instead of %d", p.SequenceNumber, n)
				return true
			}
		}
	}
	return false
}