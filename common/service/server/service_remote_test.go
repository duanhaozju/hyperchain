package server

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	pb "github.com/hyperchain/hyperchain/common/protos"
	"github.com/hyperchain/hyperchain/common/service/client"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"net"
	"sync"
	"testing"
)

const content = "This is a test event"

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
	_, err = c.Register(id, pb.FROM_EXECUTOR, &pb.RegisterMessage{
		Address:   "localhost:9001",
		Namespace: "global",
	})

	if err != nil {
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
		fmt.Printf("%d", msg.Payload)
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

func TestMultiThreadOneStreamSend(t *testing.T) {

	is := startServer(t)
	if is != nil {
		c, err := connectToServer(1, t)
		if err != nil {
			t.Error(err)
		}

		go func() {
			srv := is.ServerRegistry().Namespace("global").Service(
				fmt.Sprintf("EXECUTOR-%d", 1))
			//go srv.Serve()

			s, ok := srv.(*remoteServiceImpl)
			if !ok {
				t.Error("service can not type assert")
			}
			rspCh := s.msg
			for {
				select {
				case rsp := <-rspCh:
					event := &event.TestEvent{}
					err := proto.Unmarshal(rsp.Payload, event)
					if err != nil {
						t.Error(err)
					}
					fmt.Printf("%v\n", event)
				}
			}
		}()

		wg := sync.WaitGroup{}
		n := 1000
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				for j := 0; j < n; j++ {
					fmt.Printf("Try to send msg: %d\n", j)

					payload, err := proto.Marshal(&event.TestEvent{
						Id:  uint64(j),
						Msg: content,
					})

					if err != nil {
						t.Error(err)
					}

					c.Send(&pb.IMessage{
						From:    pb.FROM_EXECUTOR,
						Payload: payload,
					})
				}

				wg.Done()
			}()
		}
		wg.Wait()
	}

}
