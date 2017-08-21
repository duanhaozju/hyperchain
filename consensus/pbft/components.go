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
	d string
}

//Batch state cache
type msgCert struct {
	resultHash   string
	vid	     uint64
	prePrepare   *PrePrepare
	sentPrepare  bool
	prepare      map[Prepare]bool
	sentValidate bool
	validated    bool
	sentCommit   bool
	commit       map[Commit]bool
	sentExecute  bool
	pStored      bool
	cStored      bool
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
	resultHash	string
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

type Xset map[uint64]string

type vidx struct {
	view uint64
	seqNo uint64
	vid uint64
}