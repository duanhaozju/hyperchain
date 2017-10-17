package main

import (
	"github.com/op/go-logging"
	pb "hyperchain/common/protos"
	"hyperchain/service/network/handler"
	"hyperchain/common/service"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("network")
}

func main() {

	client, err := service.New(60061, "127.0.0.1", service.NETWORK, "global")
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

	client.ProcessMessagesUntilExit()
}
