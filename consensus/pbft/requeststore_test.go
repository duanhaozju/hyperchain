package pbft

import (
	"testing"
	"time"
	"strconv"

	"hyperchain/core/types"
	//"bytes"
	"container/list"
	"os"
)

func TestOrderedRequests(t *testing.T) {
	print(os.Getenv("GOPATH"))
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
	oq := &orderedRequests{presence:make(map[string]*list.Element), order:list.List{}}
	if oq.Len() != 0 {
		t.Errorf("error orderedRequests len() error!")
	}
	//oq = nil
	t1 := &types.Transaction{From:[]byte("from addr"), To:[]byte("to addr"), Value:[]byte("100")}
	t2 := &types.Transaction{From:[]byte("from addr"), To:[]byte("to addr"), Value:[]byte("200")}
	oq.add(t1)
	oq.add(t2)
	if oq.Len() != 2 {
		t.Errorf("error Len() = %d, expected: %d", oq.Len(), 2)
	}
}

func TestRemoves(t *testing.T)  {
	oq := &orderedRequests{presence:make(map[string]*list.Element)}
	t1 := &types.Transaction{From:[]byte("from addr"), To:[]byte("to addr"), Value:[]byte("100")}
	t2 := &types.Transaction{From:[]byte("from addr"), To:[]byte("to addr"), Value:[]byte("200")}
	oq.add(t1)
	oq.add(t2)
	tr := []*types.Transaction{t1, t2}
	oq.removes(tr)
	if oq.Len() != 0 {
		t.Errorf("error removes not worked!")
	}
	t3 := &types.Transaction{From:[]byte("from addr"), To:[]byte("to addr"), Value:[]byte("400")}
	tr = []*types.Transaction{t3}
	rs := oq.removes(tr)
	if rs != false {
		t.Errorf("error removes failed to removes(nil)")
	}
}

func TestNewRequestStore(t *testing.T)  {
	rs := newRequestStore()
	if rs.outstandingRequests == nil || rs.pendingRequests == nil {
		t.Errorf("error newRequestStore not worked!")
	}
}

func TestStoreOutstanding(t *testing.T)  {
	rs := newRequestStore()
	t1 := &types.Transaction{From:[]byte("from addr"), To:[]byte("to addr"), Value:[]byte("100")}
	rs.storeOutstanding(t1)
	if rs.outstandingRequests.has(hash(t1)) == false {
		t.Errorf("error storeOutstanding(%v) failed", t1)
	}
}

func TestStorePending(t *testing.T)  {
	rs := newRequestStore()
	t1 := &types.Transaction{From:[]byte("from addr"), To:[]byte("to addr"), Value:[]byte("100")}
	rs.storePending(t1)
	if rs.pendingRequests.has(hash(t1)) == false {
		t.Errorf("error storePending(%v) failed", t1)
	}
}

func TestStorePendingsAndRemove(t *testing.T)  {
	rs := newRequestStore()
	t1 := &types.Transaction{From:[]byte("from addr"), To:[]byte("to addr"), Value:[]byte("100")}
	t2 := &types.Transaction{From:[]byte("from addr"), To:[]byte("to addr"), Value:[]byte("200")}
	tr := []*types.Transaction{t1, t2}

	rs.storePendings(tr)
	for _, tx := range tr {
		if rs.pendingRequests.has(hash(tx)) == false {
			t.Errorf("error storePendings failed!")
		}
	}

	rs.remove(t1)
	if rs.pendingRequests.has(hash(t1)) || rs.outstandingRequests.has(hash(t1)) {
		t.Errorf("error pendingRequests failed!")
	}

}

func TestHasNonPending(t *testing.T)  {
	rs := newRequestStore()
	t1 := &types.Transaction{From:[]byte("from addr"), To:[]byte("to addr"), Value:[]byte("100")}
	hnp := rs.hasNonPending()
	if hnp == true {
		t.Errorf("error hasNonPending() = true, expected false")
	}
	rs.storeOutstanding(t1)
	hnp = rs.hasNonPending()
	if hnp == false {
		t.Errorf("error hasNonPending() = false, expected true")
	}
}

func TestGetNextNonPending(t *testing.T)  {
	rs := newRequestStore()
	t1 := &types.Transaction{From:[]byte("from addr1"), To:[]byte("to addr"), Value:[]byte("100")}
	t2 := &types.Transaction{From:[]byte("from addr2"), To:[]byte("to addr"), Value:[]byte("200")}
	t3 := &types.Transaction{From:[]byte("from addr3"), To:[]byte("to addr"), Value:[]byte("100")}
	t4 := &types.Transaction{From:[]byte("from addr4"), To:[]byte("to addr"), Value:[]byte("200")}

	rs.storeOutstanding(t1)
	rs.storeOutstanding(t2)
	rs.storeOutstanding(t3)
	rs.storeOutstanding(t4)

	nn := rs.getNextNonPending(4)
	if len(nn) != 4 {
		t.Errorf("error getNextNonPending(%d) failed", 4)
	}
	nn = rs.getNextNonPending(3)
	if len(nn) != 3 {
		t.Errorf("error getNextNonPending(%d) failed", 3)
	}
	nn = rs.getNextNonPending(5)
	if len(nn) != 4 {
		t.Errorf("error getNextNonPending(%d) failed", 5)
	}

	rs.pendingRequests.add(t1)
	rs.pendingRequests.add(t2)

	nn = rs.getNextNonPending(4)
	if len(nn) != 2 {
		t.Errorf("error getNextNonPending(%d) failed", 2)
	}
}