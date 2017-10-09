package rbft

import (
	"github.com/facebookgo/ensure"
	"github.com/gogo/protobuf/proto"
	"hyperchain/common"
	"hyperchain/manager/protos"
	"strconv"
	"testing"
	"time"
)

func TestPbftImpl_NewPbft(t *testing.T) {

	//new PBFT
	pbft, conf, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)

	ensure.Nil(t, err)

	ensure.DeepEqual(t, pbft.namespace, "global-"+strconv.Itoa(int(pbft.id)))
	ensure.DeepEqual(t, pbft.activeView, uint32(1))
	ensure.DeepEqual(t, pbft.f, (pbft.N-1)/3)
	ensure.DeepEqual(t, pbft.N, conf.GetInt("self.N"))
	ensure.DeepEqual(t, pbft.h, uint64(0))
	ensure.DeepEqual(t, pbft.id, uint64(conf.GetInt64(common.C_NODE_ID)))
	ensure.DeepEqual(t, pbft.K, uint64(10))
	ensure.DeepEqual(t, pbft.logMultiplier, uint64(4))
	ensure.DeepEqual(t, pbft.L, pbft.logMultiplier*pbft.K)
	ensure.DeepEqual(t, pbft.seqNo, uint64(0))
	ensure.DeepEqual(t, pbft.view, uint64(0))

	//Test Consenter interface

	pbft.Start()

	//pbft.RecvMsg()
	pbft.Close()

}

//pbMsg := &protos.Message{
//Type:      protos.Message_NULL_REQUEST,
//Payload:   nil,
//Timestamp: time.Now().UnixNano(),
//Id:        id,
//}
func TestProcessNullRequest(t *testing.T) {
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	ensure.Nil(t, err)
	pbMsg := &protos.Message{
		Type:      protos.Message_NULL_REQUEST,
		Payload:   nil,
		Timestamp: time.Now().UnixNano(),
		Id:        1,
	}
	event, err := proto.Marshal(pbMsg)
	rbft.RecvMsg(event)
	time.Sleep(3 * time.Second)
	err = CleanData(rbft.namespace)
	ensure.Nil(t, err)

}
