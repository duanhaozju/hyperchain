package pbft

import (
	"testing"
	"time"

	"hyperchain/core/types"
)


func TestHash(t *testing.T) {

	var a, b []byte
	copy(a[:], "abc")
	copy(b[:], "def")

	req1 := &types.Transaction{
		Timestamp:	time.Now().UnixNano(),
		Value:		a,
		Id:		0,
		Signature:	a,
	}

	req2 := &types.Transaction{
		Timestamp:	time.Now().UnixNano(),
		Value:		b,
		Id:		1,
		Signature:	b,
	}

	reqBatch := &TransactionBatch{Batch: []*types.Transaction{req1, req2}}

	hashReq := hash(req1)
	hashReqBatch := hash(reqBatch)
	if hashReq == "" {
		t.Error("Hash Reqest error")
	} else {
		t.Logf("Hash Reqest pass, hashReq is: %s", hashReq)
	}
	if hashReqBatch == "" {
		t.Error("Hash ReqBatch error")
	} else {
		t.Logf("Hash ReqBatch pass, hashReqBatch is: %s", hashReqBatch)
	}
}
