package main

import (
    "testing"
    //"net"
    //"fmt"
    //"hyperchain/common/service"
    //"google.golang.org/grpc"
    //pb "hyperchain/common/protos"
    //"github.com/op/go-logging"
    "fmt"
)


func TestExecutor(t *testing.T) {
//    if ds, err := service.NewInternalServer(50071, "127.0.0.1"); err == nil {
//        logger = logging.MustGetLogger("service")
//        logger.Debugf("Service try to listen on addr: %s", ds.Addr())
//
//        lis, err := net.Listen("tcp", ds.Addr())
//        if err != nil {
//            fmt.Print(err)
//            return
//        }
//
//        grpcServer := grpc.NewServer()
//        pb.RegisterDispatcherServer(grpcServer, ds)
//
//        logger.Debugf("Service start successful!")
//        grpcServer.Serve(lis)
//    } else {
//        fmt.Print(err)
//    }

    fmt.Println("hello")
}