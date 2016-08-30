package pbft

import (
	"testing"
	"time"
)


func TestHash(t *testing.T) {

	var a, b []byte
	copy(a[:], "abc")
	copy(b[:], "def")

	req1 := &Request{
		Timestamp:	time.Now().Unix(),
		Payload:	a,
		ReplicaId:	0,
		Signature:	a,
	}

	req2 := &Request{
		Timestamp:	time.Now().Unix(),
		Payload:	b,
		ReplicaId:	1,
		Signature:	b,
	}

	reqBatch := &RequestBatch{Batch: []*Request{req1, req2}}

	hashReq := hash(req1)
	hashReqBatch := hash(reqBatch)
	if hashReq == "" {
		t.Errorf("Hash Reqest error")
	} else {
		t.Logf("Hash Reqest pass, hashReq is: %s", hashReq)
	}
	if hashReqBatch == "" {
		t.Errorf("Hash ReqBatch error")
	} else {
		t.Logf("Hash ReqBatch pass, hashReqBatch is: %s", hashReqBatch)
	}
}
