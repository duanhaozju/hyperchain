//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

/**
This file defines the structs uesd in PBFT
*/

//Preprepare index
type qidx struct {
	d string
	n uint64
}

//Checkpoint index
type cidx struct {
	n uint64
	d string
}

type msgID struct { // our index through certStore
	v uint64
	n uint64
}

//Batch state cache
type msgCert struct {
	digest       string
	prePrepare   *PrePrepare
	sentPrepare  bool
	prepare      map[Prepare]bool
	prepareCount int
	sentValidate bool
	validated    bool
	sentCommit   bool
	commit       map[Commit]bool
	commitCount  int
	sentExecute  bool
}

//Checkpoint id
type chkptID struct {
	n  uint64 //seq num
	id string //digest of checkpoint
}

type chkptCert struct {
	chkpts     map[Checkpoint]bool
	chkptCount int
}

//viewchange index
type vcidx struct {
	v  uint64 //view
	id uint64 //replica id
}

type cacheBatch struct {
	batch *TransactionBatch
	vid   uint64
}

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

type updateCert struct {
	digest      string
	sentPrepare bool
	validated   bool
	sentCommit  bool
	sentExecute bool
}

type Xset map[uint64]string
