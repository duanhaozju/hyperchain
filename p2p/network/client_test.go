package network_test

import (
	. "hyperchain/p2p/network"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"hyperchain/p2p/message"
	"fmt"
)

var _ = Describe("Client", func() {
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
			It("client should got a non-error response",func(){
				client := NewClient("127.0.0.1:50013")
				Expect(client.Connect()).To(BeNil())
				msg := &message.Message{
					MessageType:message.Message_HELLO,
				}
				response,err := client.Greeting(msg)
				Expect(err).To(BeNil())
				Expect(response.MessageType).To(Equal(message.Message_HELLO_RESPONSE))
			})
		})
	})

	//Describe("Client BenchMark",func(){
	//	Context("with a weak server",func(){
	//		Measure("Greeting Method should be at least in 100000 op/s",func(b Benchmarker){
	//			var client *Client
	//			var msg *message.Message
	//			fmt.Println("init the before each")
	//			client = NewClient("127.0.0.1:50013)")
	//			Expect(client.Connect()).To(BeNil())
	//			msg = &message.Message{
	//				MessageType:message.Message_HELLO,
	//			}
	//			runtime := b.Time("runtime", func() {
	//				output,err :=  client.Wisper(msg)
	//				Expect(err).To(BeNil())
	//				Expect(output.MessageType).To(Equal(message.Message_HELLO))
	//			})
	//
	//			Ω(runtime.Seconds()).Should(BeNumerically("<", 0.2), "SomethingHard() shouldn't take too long.")
	//
	//			//b.RecordValue("disk usage (in MB)", HowMuchDiskSpaceDidYouUse())
	//		},100)
	//
	//	})
	//})
})
