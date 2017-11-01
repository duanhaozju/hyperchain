package server

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"hyperchain/common/service/client"
	pb "hyperchain/common/protos"
	"net"
	"sync"
	"testing"
)

func startServer(t *testing.T) *InternalServer {
	is, err := NewInternalServer(60061, "localhost")
	if err != nil {
		return nil
	}
	grpcServer := grpc.NewServer()

	pb.RegisterDispatcherServer(grpcServer, is)

	//t.Logf("Server listerning on %s ", is.Addr())
	lis, err := net.Listen("tcp", "localhost:60061")

	if err != nil {
		t.Error(err)
	}

	go grpcServer.Serve(lis)

	return is
}

func connectToServer(id uint64, t *testing.T) (*client.ServiceClient, error) {
	c, err := client.New(60061, "localhost", client.EXECUTOR, "global")
	if err != nil {
		t.Error(err)
		return nil, err
	}
	if err := c.Connect(); err != nil {
		return nil, err
	}
	if err := c.Register(id, pb.FROM_EXECUTOR, &pb.RegisterMessage{
		Address:   "localhost:9001",
		Namespace: "global",
	}); err != nil {
		return nil, err
	}
	return c, nil
}

type AuxHandler struct {
	t *testing.T
}

func (ah *AuxHandler) Handle(client pb.Dispatcher_RegisterClient, msg *pb.IMessage) {
	//fmt.Printf("receive message %v \n", msg)
	if msg.Type == pb.Type_SYNC_REQUEST {
		err := client.Send(&pb.IMessage{
			Type: pb.Type_RESPONSE,
			Id:   msg.Id,
			Ok:   true,
			From: pb.FROM_EXECUTOR,
		})

		if err != nil {
			ah.t.Error(err)
		}

	} else {
		ah.t.Error("Invalid msg type, %s ", msg.Type)
	}
}

func testOneConnection(id uint64, is *InternalServer, t *testing.T) {
	if is != nil {
		c, err := connectToServer(id, t)
		if err != nil {
			t.Error(err)
		}
		c.AddHandler(&AuxHandler{t: t})
		//time.Sleep(1 * time.Second)
		srv := is.ServerRegistry().Namespace("global").Service(
			fmt.Sprintf("EXECUTOR-%d", id))
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			len := 1000
			var pre uint64 = 0
			for i := 1; i < len; i++ {
				rsp, err := srv.SyncSend(&pb.IMessage{
					Type: pb.Type_SYNC_REQUEST,
				})
				if err != nil {
					t.Error(err)

				}
				if pre == 0 {
					pre = rsp.Id
				} else {
					assert.Equal(t, true, pre+1 == rsp.Id)
					pre = rsp.Id
				}
			}
			wg.Done()
		}()
		wg.Wait()
	}
}

func TestSyncSend(t *testing.T) {
	//TODO: clear the connection info
	is := startServer(t)
	wgg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wgg.Add(1)
		go func(id uint64) {
			testOneConnection(id, is, t)
			wgg.Done()
		}(uint64(i))
	}
	wgg.Wait()
	t.Log(is.Addr())
}
