//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/hyperchain/hyperchain/core/types"
)

func TestHash(t *testing.T) {

	var a, b []byte
	copy(a[:], "abc")
	copy(b[:], "def")

	req1 := &types.Transaction{
		Timestamp: time.Now().UnixNano(),
		Value:     a,
		Id:        0,
		Signature: a,
	}

	req2 := &types.Transaction{
		Timestamp: time.Now().UnixNano(),
		Value:     b,
		Id:        1,
		Signature: b,
	}

	reqBatch := &TransactionBatch{TxList: []*types.Transaction{req1, req2}}

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

func TestByteToString(t *testing.T) {
	s := []byte("hello, world!")
	re := byteToString(s)
	if strings.Compare(base64.StdEncoding.EncodeToString(s), re) != 0 {
		t.Error("error byteToString()")
	}
}
