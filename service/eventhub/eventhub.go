package main

import (
	"fmt"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	"hyperchain/service/common/dispatcher"
	pb "hyperchain/service/common/protos"
	"net"
)

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("eventhub")
}

func main() {

	if ds, err := dispatcher.NewDispatchServer(60061, "127.0.0.1"); err == nil {
		logger.Debugf("Eventhub try to listen on addr: %s", ds.Addr())

		lis, err := net.Listen("tcp", ds.Addr())
		if err != nil {
			fmt.Print(err)
			return
		}

		grpcServer := grpc.NewServer()
		pb.RegisterDispatcherServer(grpcServer, ds)

		logger.Debugf("Eventhub start successful!")
		grpcServer.Serve(lis)

	} else {
		fmt.Print(err)
	}

}
