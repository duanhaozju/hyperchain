package main

import (
	"github.com/op/go-logging"
	"hyperchain/service/common"
	pb "hyperchain/service/common/protos"
	"hyperchain/service/network/handler"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("network")
}

func main() {

	client, err := common.New("127.0.0.1", 60061)
	if err != nil {
		logger.Error(err)
	}

	client.AddHandler(handler.New())
	err = client.Connect()
	if err != nil {
		logger.Error(err)
	}

	err = client.Register(pb.FROM_NETWORK, &pb.RegisterMessage{
		Namespace: "global",
	})
	if err != nil {
		logger.Error(err)
	}
	logger.Debugf("Network register successful")
	go client.ProcessMessages()

	exit := make(chan struct{})

	<-exit
}
