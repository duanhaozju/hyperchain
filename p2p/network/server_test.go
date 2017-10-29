package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pb "github.com/hyperchain/hyperchain/p2p/message"
	"github.com/terasum/viper"
	"github.com/hyperchain/hyperchain/p2p/utils"
	"github.com/hyperchain/hyperchain/common"
	"context"
	"github.com/hyperchain/hyperchain/p2p/network/mocks"
	"time"
)

type helloHandler struct {}

func NewHelloHandler() *helloHandler{
	return &helloHandler{}
}

func (hanlder *helloHandler) Receive() chan<- interface{} { return make(chan interface{})}
func (hanlder *helloHandler) Process() {}
func (hanlder *helloHandler) Teardown() {}
func (hanlder *helloHandler) Execute(msg *pb.Message) (*pb.Message, error) {
	return  pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("execute it")), nil
}

var _ = Describe("Server", func() {
	Describe("starting server", func() {
		Context("with correct config file", func() {
			var s *Server
			It("should return non-error", func() {
				vip := viper.New()
				vip.SetConfigFile(utils.GetProjectPath()+"/p2p/test/global.toml")
				vip.ReadInConfig()
				vip.Set(common.P2P_ADDR, utils.GetProjectPath()+"/p2p/test/addr.toml")
				vip.Set(common.P2P_TLS_CA, utils.GetProjectPath()+"/p2p/test/tls/tlsca.ca")
				vip.Set(common.P2P_TLS_CERT, utils.GetProjectPath()+"/p2p/test/tls/tls_peer1.cert")
				vip.Set(common.P2P_TLS_CERT_PRIV, utils.GetProjectPath()+"/p2p/test/tls/tls_peer1.priv")

				// start server
				sec, err := NewSec(vip)
				Expect(err).To(BeNil())
				s = NewServer("server_test", nil, sec)
				Expect(s.StartServer(":50016")).To(BeNil())
				Expect(s.Claim()).Should(Equal("server_test"))
			})

			Context("to register handler", func() {
				It("should register successfully", func() {
					// register slot
					Expect(s.RegisterSlot("", pb.MsgType_CLIENTHELLO, NewHelloHandler())).To(BeNil())
				})
			})

			Context("to send Greeting message", func() {
				msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("it's greeting"))
				Context("without from", func() {
					It("should return error", func() {
						// greeting without 'from'
						_, err := s.Greeting(context.Background(), msg)
						Expect(err).NotTo(BeNil())
					})
				})
				Context("with from", func() {
					It("should send successfully", func() {
						// greeting with 'from'
						msg.From = &pb.Endpoint{Field: []byte("")}
						resp, err := s.Greeting(context.Background(), msg)
						Expect(err).To(BeNil())
						Expect(resp.Payload).Should(Equal([]byte("execute it")))
					})
				})
			})

			Context("to send Whisper message", func() {
				msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("it's whisper"))
				Context("without from", func() {
					It("should return error", func() {
						// greeting without 'from'
						_, err := s.Whisper(context.Background(), msg)
						Expect(err).NotTo(BeNil())
					})
				})
				Context("with from", func() {
					It("should send successfully", func() {
						// greeting with 'from'
						msg.From = &pb.Endpoint{Field: []byte("")}
						resp, err := s.Whisper(context.Background(), msg)
						Expect(err).To(BeNil())
						Expect(resp.Payload).Should(Equal([]byte("execute it")))
					})
				})
			})

			Context("to send Discuss message", func() {
				pkg := pb.NewPkg([]byte("it's discuss"), pb.ControlType_Notify)
				It("should send successfully", func() {
					_, err := s.Discuss(context.Background(), pkg)
					Expect(err).To(BeNil())
				})
			})

			Context("to send Chat message", func() {
				mockChatServer := &mocks.MockChat_ChatServer{}
				Context("without from", func() {
					It("should return error", func() {
						msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("hello"))
						mockChatServer.On("Recv").Return(msg, nil).Times(1)
						go s.Chat(mockChatServer)
					})
				})

				Context("with from", func() {
					It("should send successfully", func() {
						msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("hello"))
						msg.From = &pb.Endpoint{Field: []byte(""),}
						mockChatServer.On("Recv").Return(msg, nil)
						go s.Chat(mockChatServer)
						time.Sleep(time.Second * 1)
					})
				})

			})

			Context("to deregister handler", func() {
				It("should deregister slot successfully", func() {
					// deregister slot
					Expect(s.DeregisterSlot("", pb.MsgType_CLIENTHELLO)).To(BeNil())
				})
				It("should deregister slot failed", func() {
					Expect(s.DeregisterSlot("unknown", pb.MsgType_CLIENTHELLO)).NotTo(BeNil())

					// deregister slots and stop server
					s.DeregisterSlots("")
					s.StopServer()
				})
			})
		})
	})
})
