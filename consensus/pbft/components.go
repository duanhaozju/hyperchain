//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

/**
This file defines the structs uesd in PBFT
*/

// -----------certStore related structs-----------------
// Preprepare index
type qidx struct {
	d string // digest
	n uint64 // seqNo
}

// certStore index
type msgID struct {
	v uint64 // view
	n uint64 // seqNo
	d string // digest
}

// cached consensus msgs related to batch
type msgCert struct {
	resultHash   string           // validated result hash of primary
	vid          uint64           // validate ID
	prePrepare   *PrePrepare      // pre-prepare msg
	sentPrepare  bool             // track whether broadcast prepare for this batch before or not
	prepare      map[Prepare]bool // prepare msgs received from other nodes
	sentValidate bool             // track whether sent validate event to executor module before or not
	validated    bool             // track whether received the validated result of this batch or not
	sentCommit   bool             // track whether broadcast commit for this batch before or not
	commit       map[Commit]bool  // commit msgs received from other nodes
	sentExecute  bool             // track whether sent execute event to executor module before or not
}

// -----------checkpoint related structs-----------------
// Checkpoint index
type cidx struct {
	n uint64 // checkpoint number
	d string // checkpoint digest
}

// chkptCertStore index
type chkptID struct {
	n  uint64 // seq num
	id string // digest of checkpoint
}

// checkpoints received from other nodes with same chkptID
type chkptCert struct {
	chkpts     map[Checkpoint]bool
	chkptCount int
}

// -----------validate related structs-----------------
// cached batch related params stored in cacheValidatedBatch
type cacheBatch struct {
	batch      *TransactionBatch
	vid        uint64
	resultHash string
}

// preparedCert index
type vidx struct {
	view  uint64
	seqNo uint64
	vid   uint64
}

// -----------viewchange related structs-----------------
// viewchange index
type vcidx struct {
	v  uint64 // view
	id uint64 // replica id
}

type Xset map[uint64]string

// -----------node addition/deletion related structs-----------------
type addNodeCert struct {
	addNodes  map[AddNode]bool
	addCount  int
	finishAdd bool
}

type delNodeCert struct {
	newId      uint64
	delId      uint64
	routerHash string
	delNodes   map[DelNode]bool
	delCount   int
	finishDel  bool
}

type aidx struct {
	v    uint64
	n    int64
	id   uint64
	flag bool
}

type uidx struct {
	v    uint64
	n    int64
	flag bool
	key  string
}

// -----------state update related structs-----------------
type checkpointMessage struct {
	seqNo uint64
	id    []byte
}

type replicaInfo struct {
	id      uint64
	height  uint64
	genesis uint64
}

type stateUpdateTarget struct {
	checkpointMessage
	replicas []replicaInfo
}
