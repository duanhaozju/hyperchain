package main

import (
	"github.com/op/go-logging"
	"hyperchain/service/common"
	pb "hyperchain/service/common/protos"
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

	exit := make(chan struct{})

	<-exit
}
