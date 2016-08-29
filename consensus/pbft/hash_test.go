package pbft

//import (
//	"testing"
//
//	"github.com/golang/protobuf/ptypes/timestamp"
//)
//
//
//func TestHash(t *testing.T) {
//
//	var a, b []byte
//	copy(a[:], "abc")
//	copy(b[:], "def")
//	time := &timestamp.Timestamp{Seconds: 3, Nanos: 2}
//
//	req1 := &Request{
//		Timestamp:	time,
//		Payload:	a,
//		ReplicaId:	0,
//		Signature:	a,
//	}
//
//	req2 := &Request{
//		Timestamp:	time,
//		Payload:	b,
//		ReplicaId:	1,
//		Signature:	b,
//	}
//
//	reqBatch := &Request{req1, req2}
//
//	hashReq := hash(req1)
//	hashReqBatch := hash(reqBatch)
//	t.Logf("hashReq is: %s", hashReq)
//	t.Logf("hashReq is: %s", hashReqBatch)
//}
