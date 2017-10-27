//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import "github.com/op/go-logging"

/**
This file provide a mechanism to manage the storage in RBFT
*/

// storeManager manages common store data structures for RBFT.
type storeManager struct {
	logger *logging.Logger

	// track quorum certificates for messages
	certStore map[msgID]*msgCert

	// track the committed cert to help execute
	committedCert map[msgID]string

	// track whether we are waiting for transaction batches to execute
	outstandingReqBatches map[string]*TransactionBatch

	// track L cached transaction batches produced from txPool
	txBatchStore map[string]*TransactionBatch

	// for all the assigned, non-checkpointed request batches we might miss
	// some transactions in some batches, record batch id
	missingReqBatches map[string]bool

	// Set to the highest weak checkpoint cert we have observed
	highStateTarget *stateUpdateTarget

	// ---------------checkpoint related--------------------
	// checkpoints that we reached by ourselves after commit a block with a
	// block number == integer multiple of K; map lastExec to a base64
	// encoded BlockchainInfo
	chkpts map[uint64]string

	// checkpoint numbers received from others which are bigger than our
	// H(=h+L); map replicaID to the last checkpoint number received from
	// that replica bigger than H
	hChkpts map[uint64]uint64

	// track all non-repeating checkpoints received from others
	checkpointStore map[Checkpoint]bool

	// track quorum certificates for checkpoints with the same chkptID; map
	// chkptID(seqNo and id) to chkptCert(all checkpoints with that chkptID)
	chkptCertStore map[chkptID]*chkptCert
}

// newStoreMgr news an instance of storeManager
func newStoreMgr(logger *logging.Logger) *storeManager {
	sm := &storeManager{}

	sm.chkpts = make(map[uint64]string)
	sm.chkpts[0] = "XXX GENESIS"
	sm.hChkpts = make(map[uint64]uint64)
	sm.checkpointStore = make(map[Checkpoint]bool)
	sm.chkptCertStore = make(map[chkptID]*chkptCert)
	sm.certStore = make(map[msgID]*msgCert)
	sm.committedCert = make(map[msgID]string)
	sm.outstandingReqBatches = make(map[string]*TransactionBatch)
	sm.txBatchStore = make(map[string]*TransactionBatch)
	sm.missingReqBatches = make(map[string]bool)

	sm.logger = logger
	return sm
}

// moveWatermarks removes useless set in chkpts, plist, qlist whose index <= h
func (sm *storeManager) moveWatermarks(rbft *rbftImpl, h uint64) {
	for n := range sm.chkpts {
		if n < h {
			delete(sm.chkpts, n)
			rbft.persistDelCheckpoint(n)
		}
	}

	for idx := range rbft.vcMgr.qlist {
		if idx.n <= h {
			delete(rbft.vcMgr.qlist, idx)
		}
	}

	for n := range rbft.vcMgr.plist {
		if n <= h {
			delete(rbft.vcMgr.plist, n)
		}
	}
}

// saveCheckpoint saves checkpoint information to chkpts, whose key is lastExec, value is the global hash of current
// BlockchainInfo
func (sm *storeManager) saveCheckpoint(l uint64, gh string) {
	sm.chkpts[l] = gh
}

// Given a digest/view/seq, is there an entry in the certStore?
// If so, return it else, create a new entry
func (sm *storeManager) getCert(v uint64, n uint64, d string) (cert *msgCert) {
	idx := msgID{v, n, d}
	cert, ok := sm.certStore[idx]

	if ok {
		return
	}

	prepare := make(map[Prepare]bool)
	commit := make(map[Commit]bool)
	cert = &msgCert{
		prepare: prepare,
		commit:  commit,
	}
	sm.certStore[idx] = cert
	return
}

// Given a seqNo/id, is there an entry in the chkptCertStore?
// If so, return it, else, create a new entry
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

// existedDigest checks if there exists another PRE-PREPARE message in certStore which has the same digest, same view,
// but different seqNo with the given one
func (sm *storeManager) existedDigest(n uint64, view uint64, digest string) bool {
	for _, cert := range sm.certStore {
		if p := cert.prePrepare; p != nil {
			if p.View == view && p.SequenceNumber != n && p.BatchDigest == digest && digest != "" {
				// This will happen if primary receive same digest result of txs
				// It may result in DDos attack
				sm.logger.Warningf("Other prePrepare found with same digest but different seqNo: %d "+
					"instead of %d", p.SequenceNumber, n)
				return true
			}
		}
	}
	return false
}
