package main

import (
	"github.com/op/go-logging"
	"hyperchain/service/common"
	pb "hyperchain/service/common/protos"
	//"time"
	"hyperchain/manager/event"
	"github.com/gogo/protobuf/proto"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("consenter")
}

func main() {

	client, err := common.New("127.0.0.1", 60061)
	if err != nil {
		logger.Error(err)
	}

	err = client.Connect()
	if err != nil {
		logger.Error(err)
	}

	err = client.Register(pb.FROM_CONSENSUS, &pb.RegisterMessage{
		Namespace: "global",
	})
	if err != nil {
		logger.Error(err)
	}
	logger.Debugf("Consenter register successful")


	go func() {
		var i uint64 = 0
		for ; i < 10000000; i ++ {

			e := &event.InformPrimaryEvent{
				Primary:i,
			}

			payload, _ := proto.Marshal(e)

			logger.Debugf("send dispatch event: %v", e)
			err := client.Send(&pb.Message{
				Type:pb.Type_DISPATCH,
				From:pb.FROM_CONSENSUS,
				Event:pb.Event_InformPrimaryEvent,
				Payload:payload,
			})

			if err != nil {
				logger.Error(err)
				return
			}
			//time.Sleep(1 * time.Second)
		}


	}()


	exit := make(chan struct{})

	<-exit
}
