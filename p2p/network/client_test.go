package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/terasum/viper"
	"github.com/hyperchain/hyperchain/p2p/utils"
	"github.com/hyperchain/hyperchain/common"
	pb "github.com/hyperchain/hyperchain/p2p/message"
	"fmt"
	"time"
)

var _ = Describe("Client", func() {
	var (
		 vip *viper.Viper
		 sec *Sec
		 server *Server
	)

	Context("started up server", func() {
		It("should start successfully", func() {
			vip = viper.New()
			vip.SetConfigFile(utils.GetProjectPath()+"/p2p/test/global.toml")
			vip.ReadInConfig()
			vip.Set(common.P2P_ADDR, utils.GetProjectPath()+"/p2p/test/addr.toml")
			vip.Set(common.P2P_TLS_CA, utils.GetProjectPath()+"/p2p/test/tls/tlsca.ca")
			vip.Set(common.P2P_TLS_CERT, utils.GetProjectPath()+"/p2p/test/tls/tls_peer1.cert")
			vip.Set(common.P2P_TLS_CERT_PRIV, utils.GetProjectPath()+"/p2p/test/tls/tls_peer1.priv")
			vip.Set("p2p.keepAliveDuration", "1s")
			vip.Set("p2p.pendingDuration", "1s")

			// start server
			var err error
			sec, err = NewSec(vip)
			Expect(err).To(BeNil())
			server = NewServer("server_test", nil, sec)
			Expect(server.StartServer(":50015")).To(BeNil())
		})

		Context("to register server handler", func() {
			It("should register successfully", func() {
				// register slot
				Expect(server.RegisterSlot("test", pb.MsgType_CLIENTHELLO, NewHelloHandler())).To(BeNil())
			})
		})
	})


	Context("creating a new client", func() {
		var client *Client

		Context("with correct client config", func() {
			It("should create successfully", func() {
				ccnf :=  NewClientConf(vip)
				var err error
				client, err = NewClient("node5", "127.0.0.1:50015", sec, ccnf)
				Expect(err).To(BeNil())
			})
		})

		Context("to send greeting message", func() {
			msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("it's greeting"))
			Context("without from", func() {
				It("should return error", func() {
					// greeting without 'from'
					_, err := client.Greeting(msg)
					Expect(err).NotTo(BeNil())
				})
			})
			Context("with from", func() {
				It("should send successfully", func() {
					// greeting with 'from'
					msg.From = &pb.Endpoint{Field: []byte("test")}
					By(fmt.Sprintf("%#v", msg.From))
					By(fmt.Sprintf("%#v", msg.From.Field))
					resp, err := client.Greeting(msg)
					Expect(err).To(BeNil())
					Expect(resp.Payload).Should(Equal([]byte("execute it")))
				})

				//client.Close()
				//server.StopServer()
			})
		})

		Context("to send Whisper message", func() {
			msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("it's whisper"))
			Context("without from", func() {
				It("should return error", func() {
					// greeting without 'from'
					_, err := client.Whisper(msg)
					Expect(err).NotTo(BeNil())
				})
			})
			Context("with from", func() {
				It("should send successfully", func() {
					// greeting with 'from'
					msg.From = &pb.Endpoint{Field: []byte("test")}
					resp, err := client.Whisper(msg)
					Expect(err).To(BeNil())
					Expect(resp.Payload).Should(Equal([]byte("execute it")))
				})
			})
		})

		Context("to send Discuss message", func() {
			pkg := pb.NewPkg([]byte("it's discuss"), pb.ControlType_Notify)
			It("should send successfully", func() {
				_, err := client.Discuss(pkg)
				Expect(err).To(BeNil())
			})
		})

		Context("to send Chat message", func() {
			Context("without from", func() {
				It("should return error", func() {
					go client.Chat()
				})
			})

			Context("with from", func() {
				It("should send successfully", func() {
					msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("hello"))
					msg.From = &pb.Endpoint{Field: []byte("test"),}
					client.MsgChan <- msg
					go client.Chat()
					time.Sleep(time.Millisecond * 500)
				})
			})
		})

		Context("to test state machine", func() {
			Context("from working to pending", func() {
				It("should transition successfully", func() {
					err := client.stateMachine.Event(c_EventError)
					Expect(err).To(BeNil())
					time.Sleep(time.Second * 1)
				})
			})

			Context("from pending to working", func() {
				It("should transition successfully", func() {
					err := client.stateMachine.Event(c_EventRecovery)
					Expect(err).To(BeNil())
					time.Sleep(time.Second * 1)
				})
			})

			Context("from pending to closing", func() {
				It("should transition successfully", func() {
					err := client.stateMachine.Event(c_EventError)
					Expect(err).To(BeNil())
					time.Sleep(time.Second * 1)

					err = client.stateMachine.Event(c_EventClose)
					Expect(err).To(BeNil())
					time.Sleep(time.Second * 1)

					server.StopServer()
					client.Close()
				})
			})
		})
	})
})
