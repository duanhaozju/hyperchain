package network_test

import (
	. "hyperchain/p2p/network"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"hyperchain/p2p/message"
	"fmt"
	"hyperchain/p2p/msg"
	"time"
)

var _ = Describe("Client", func() {
	var server *Server
	BeforeEach(func(){
		server = NewServer()
	})
	JustBeforeEach(func(){
		time.After(time.Second)
		err := server.StartServer(50012)
		if err != nil{
			fmt.Println(err)
		}
		Expect(err).To(BeNil())
	})
	AfterEach(func(){
		server.StopServer()
	})
	Describe("Client connect",func(){
		Context("when create a client and connect to server",func(){
			It("client should return non-error",func(){
				client := NewClient("127.0.0.1:50012")
				Expect(client.Connect()).To(BeNil())
			})
		})
		Context("when create multi client and connect to server",func(){
			It("two clients should return non-error",func(){
				client1 := NewClient("127.0.0.1:50012")
				client2 := NewClient("127.0.0.1:50012")
				Expect(client1.Connect()).To(BeNil())
				Expect(client2.Connect()).To(BeNil())
			})
			It("three clients should return non-error",func(){
				client1 := NewClient("127.0.0.1:50012")
				client2 := NewClient("127.0.0.1:50012")
				client3 := NewClient("127.0.0.1:50012")
				Expect(client1.Connect()).To(BeNil())
				Expect(client2.Connect()).To(BeNil())
				Expect(client3.Connect()).To(BeNil())
			})
		})
	})

	Describe("Client Communication",func(){
		Context("When chat with a weak server",func(){
			It("client should got a error response, because server not support",func(){
				client := NewClient("127.0.0.1:50012")
				Expect(client.Connect()).To(BeNil())
				msg := &message.Message{
					MessageType:message.Message_HELLO,
				}
				response,err := client.Greeting(msg)
				Expect(response).To(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf("rpc error: code = 2 desc = This message type is not support, %v",msg.MessageType)))
			})

		})
		Context("When chat with a stronger server",func(){
			BeforeEach(func(){
				server := NewServer()
				fmt.Println("this should fun for setup a stronger server")
				blackHoleChain := make(chan *message.Message)
				helloHandler := msg.NewHelloHandler(blackHoleChain)
				err := server.RegisterSlot(message.Message_HELLO,helloHandler)
				Expect(err).To(BeNil())
			})
			It("client should got a non-error response",func(){
				client := NewClient("127.0.0.1:50012")
				Expect(client.Connect()).To(BeNil())
				msg := &message.Message{
					MessageType:message.Message_HELLO,
				}
				response,err := client.Greeting(msg)
				Expect(response).To(BeNil())
				Expect(err.Error()).To(Equal(BeNil()))
			})
		})
	})
})
