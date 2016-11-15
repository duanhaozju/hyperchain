package pbft

import (
	"testing"
	"time"
	"strconv"

	"hyperchain/core/types"
)

func TestOrderedRequests(t *testing.T) {
	or := &orderedRequests{}
	or.empty()

	r1 := createPbftReq(2, 1)
	r2 := createPbftReq(2, 2)
	r3 := createPbftReq(19, 1)
	if or.has(or.wrapRequest(r1).key) {
		t.Error("should not have req")
	}
	or.add(r1)
	if !or.has(or.wrapRequest(r1).key) {
		t.Error("should have req")
	}
	if or.has(or.wrapRequest(r2).key) {
		t.Error("should not have req")
	}
	if or.remove(r2) {
		t.Error("should not have removed req")
	}
	if !or.remove(r1) {
		t.Error("should have removed req")
	}
	if or.remove(r1) {
		t.Error("should not have removed req")
	}
	if or.order.Len() != 0 || len(or.presence) != 0 {
		t.Error("should have 0 len")
	}
	or.adds([]*types.Transaction{r1, r2, r3})

	if or.order.Back().Value.(requestContainer).req != r3 {
		t.Error("incorrect order")
	}
}

func createPbftReq(tag int, replica uint64) (tx *types.Transaction) {

	var payload []byte
	temp := strconv.Itoa(tag)
	copy(payload[:], temp)

	tx = &types.Transaction{
		Timestamp:	time.Now().UnixNano(),
		Id:		replica,
		Value:		payload,
	}

	return
}

func BenchmarkOrderedRequests(b *testing.B) {
	or := &orderedRequests{}
	or.empty()

	Nreq := 1000

	reqs := make(map[string]*types.Transaction)
	for i := 0; i < Nreq; i++ {
		rc := or.wrapRequest(createPbftReq(i, 0))
		reqs[rc.key] = rc.req
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, r := range reqs {
			or.add(r)
		}

		for k := range reqs {
			_ = or.has(k)
		}

		for _, r := range reqs {
			or.remove(r)
		}
	}
}

func TestLen(t *testing.T)  {
	oq := &orderedRequests{}
	if oq.Len() != 0 {
		t.Errorf("orderedRequests len() error!")
	}
	oq = nil
	oq.Len()


}