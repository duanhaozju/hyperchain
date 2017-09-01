package main

import (
	//"context"
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	pb "hyperchain/core/vm/jcee/protos"
	"io"
	"time"
)

const totalCount = 10 * 10000

func main() {
	//streamTest()
	synCallTest()
}

func streamTest() {
	fmt.Println("client start !")

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Error(err)
	}

	client := pb.NewContractClient(conn)

	client.Execute(context.Background(), &pb.Request{
		Method: "test",
	})

	stream, err := client.StreamExecute(context.Background())
	if err != nil {
		fmt.Print(err)
		return
	}

	waitc := make(chan struct{})
	start := time.Now()
	//var totalCount int = 10 * 10000
	var c int = 0
	go func() {
		for c < totalCount {
			in, err := stream.Recv()
			c++
			if err == io.EOF {
				close(waitc)
				return
			}

			if err != nil {
				fmt.Print("Failed to receive a note : %v", err)
			}

			fmt.Printf("receive response %d: %v\n", c, in)
		}
		waitc <- struct{}{}
	}()

	for i := 0; i < totalCount; i++ {
		stream.Send(&pb.Request{
			Method: "Test Request",
		})
	}

	defer conn.Close()
	<-waitc

	end := time.Now()
	fmt.Printf("stream time used: %v", end.Sub(start))
}

func synCallTest() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Error(err)
	}

	client := pb.NewContractClient(conn)
	var totalCount int = 10 * 10000
	start := time.Now()
	for i := 0; i < totalCount; i++ {
		rsp, _ := client.Execute(context.Background(), &pb.Request{Method: "synCall"})

		fmt.Printf("receive response %d: %v\n", i, rsp)
	}
	fmt.Printf("syn call time used: %v", time.Now().Sub(start))

}
