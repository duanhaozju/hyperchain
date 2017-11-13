package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/hyperchain/hyperchain/core/vm/jcee/protos"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"io"
	"time"
)

const (
	sync  = "sync"
	async = "async"
)

var (
	port       *string
	host       *string
	totalCount *int
)

func main() {

	m := flag.String("m", "sync", "decide run sync or async call")
	port = flag.String("p", "50051", "server port")
	host = flag.String("h", "localhost", "server host")
	totalCount = flag.Int("tc", 100000, "total num of request")

	flag.Parse()
	switch *m {
	case async:
		asyncCallTest()
	case sync:
		synCallTest()
	}
}

//test async invoke by stream
func asyncCallTest() {
	fmt.Println("client start !")
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", *host, *port), grpc.WithInsecure())
	if err != nil {
		log.Error(err)
	}
	client := pb.NewContractClient(conn)

	stream, err := client.Register(context.Background())
	if err != nil {
		fmt.Print(err)
		return
	}

	waitc := make(chan struct{})
	start := time.Now()
	//var totalCount int = 10 * 10000
	var c int = 0
	go func() {
		for c < *totalCount {
			in, err := stream.Recv()
			c++
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				fmt.Printf("Failed to receive a note : %v\n", err)
			}
			fmt.Printf("receive response %d: %v\n", c, in)
		}
		waitc <- struct{}{}
	}()

	for i := 0; i < *totalCount; i++ {
		stream.Send(&pb.Message{
			Type:    pb.Message_TRANSACTION,
			Payload: []byte("asyncCall"),
		})
	}

	defer conn.Close()
	<-waitc
	end := time.Now()
	fmt.Printf("stream time used: %v", end.Sub(start))
}

//test sync call
func synCallTest() {
	fmt.Println(fmt.Sprintf("%s:%s", *host, *port))
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", *host, *port), grpc.WithInsecure())
	if err != nil {
		log.Error(err)
	}

	client := pb.NewContractClient(conn)
	start := time.Now()
	for i := 0; i < *totalCount; i++ {
		rsp, _ := client.HeartBeat(context.Background(), &pb.Request{Method: "syncCall"})
		fmt.Printf("receive response %d: %v\n", i, rsp)
	}
	fmt.Printf("syn call time used: %v", time.Now().Sub(start))
}
