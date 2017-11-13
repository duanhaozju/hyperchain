package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/terasum/viper"
	"github.com/hyperchain/hyperchain/p2p/utils"
	"github.com/hyperchain/hyperchain/common"
	pb "github.com/hyperchain/hyperchain/p2p/message"
)

var _ = Describe("Hypernet", func() {
	Describe("creating a new hypernet instance", func(){
		vip := viper.New()
		vip.SetConfigFile(utils.GetProjectPath()+"/p2p/test/global.toml")
		vip.ReadInConfig()

		Context("with incorrect config file", func() {
			It("should create failed", func() {
				_, err := NewHyperNet(nil, "hyperchain")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("with invalid port", func() {
			It("should create failed", func() {
				vip.Set(common.P2P_PORT, "abc")
				_, err := NewHyperNet(vip, "hyperchain")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("with invalid hosts.toml", func() {
			It("should create failed", func() {
				vip.Set(common.P2P_PORT, "50011")
				vip.Set(common.P2P_HOSTS, utils.GetProjectPath()+"/p2p/test/notfound.toml")
				_, err := NewHyperNet(vip, "hyperchain")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("with invalid addr.toml", func() {
			It("should create failed", func() {
				vip.Set(common.P2P_HOSTS, utils.GetProjectPath()+"/p2p/test/hosts.toml")
				vip.Set(common.P2P_ADDR, utils.GetProjectPath()+"/p2p/test/notfound.toml")
				_, err := NewHyperNet(vip, "hyperchain")
				Expect(err).NotTo(BeNil())
			})
		})

		vip.Set(common.P2P_ADDR, utils.GetProjectPath()+"/p2p/test/addr.toml")
		vip.Set(common.P2P_HOSTS, utils.GetProjectPath()+"/p2p/test/hosts.toml")
		vip.Set(common.P2P_TLS_CA, utils.GetProjectPath()+"/p2p/test/tls/tlsca.ca")
		vip.Set(common.P2P_TLS_CERT, utils.GetProjectPath()+"/p2p/test/tls/tls_peer1.cert")
		vip.Set(common.P2P_TLS_CERT_PRIV, utils.GetProjectPath()+"/p2p/test/tls/tls_peer1.priv")
		hypernet, err := NewHyperNet(vip, "hyperchain")
		Context("with correct config file", func() {
			It("should create successfully", func() {
				Expect(err).To(BeNil())
			})
		})

		Context("to initialize server", func() {
			It("should init successfully", func() {
				err := hypernet.InitServer()
				Expect(err).To(BeNil())
			})
		})

		Context("to initialize clients", func() {
			It("should init successfuly", func() {
				err := hypernet.InitClients()
				Expect(err).To(BeNil())
			})
		})

		Context("to register handler", func() {
			It("should register successfully", func() {
				err := hypernet.RegisterHandler("test", pb.MsgType_CLIENTHELLO,  NewHelloHandler())
				Expect(err).To(BeNil())
			})
		})

		Context("to deregister handler", func() {
			It("should diregister successfully", func() {
				hypernet.DeRegisterHandlers("test")
			})
		})

		Context("to get DNS", func() {
			It("should get successfully", func() {
				ip, port := hypernet.GetDNS("node1")
				Expect(ip).Should(Equal("127.0.0.1"))
				Expect(port).Should(Equal("50011"))
			})
		})

		Context("to connect by address", func() {
			It("should connect successfully", func() {
				err := hypernet.ConnectByAddr("node1", "127.0.0.1:50011")
				Expect(err).To(BeNil())

				// register handler for the next test
				hypernet.RegisterHandler("test", pb.MsgType_CLIENTHELLO,  NewHelloHandler())
			})
		})

		Context("to send greeting message", func() {
			It("should send successfully", func() {
				msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("it's greeting"))
				msg.From = &pb.Endpoint{Field: []byte("test")}
				resp, err := hypernet.Greeting("node1", msg)
				Expect(err).To(BeNil())
				Expect(resp.Payload).Should(Equal([]byte("execute it")))
			})
			Context("with unknown client", func() {
				It("should send failed", func() {
					msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("it's greeting"))
					msg.From = &pb.Endpoint{Field: []byte("test")}
					_, err := hypernet.Greeting("whoareyou", msg)
					Expect(err).NotTo(BeNil())
				})
			})
		})

		Context("to send whisper message", func() {
			It("should send successfully", func() {
				msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("it's whisper"))
				msg.From = &pb.Endpoint{Field: []byte("test")}
				resp, err := hypernet.Whisper("node1", msg)
				Expect(err).To(BeNil())
				Expect(resp.Payload).Should(Equal([]byte("execute it")))
			})
			Context("with unknown client", func() {
				It("should send failed", func() {
					msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("it's whisper"))
					msg.From = &pb.Endpoint{Field: []byte("test")}
					_, err := hypernet.Whisper("whoareyou", msg)
					Expect(err).NotTo(BeNil())
				})
			})
		})

		Context("to send discuss message", func() {
			It("should send successfully", func() {
				pkg := pb.NewPkg([]byte("it's discuss"), pb.ControlType_Notify)
				_, err := hypernet.Discuss("node1", pkg)
				Expect(err).To(BeNil())
			})
			Context("with unknown client", func() {
				It("should send failed", func() {
					pkg := pb.NewPkg([]byte("it's discuss"), pb.ControlType_Notify)
					_, err := hypernet.Discuss("whoareyou", pkg)
					Expect(err).NotTo(BeNil())
				})
			})
		})

		Context("to send chat message", func() {
			It("should chat successfully", func() {
				msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("it's chat"))
				msg.From = &pb.Endpoint{Field: []byte("test")}
				err := hypernet.Chat("node1", msg)
				Expect(err).To(BeNil())
			})
			Context("with unknown client", func() {
				It("should chat failed", func() {
					msg := pb.NewMsg(pb.MsgType_CLIENTHELLO, []byte("it's chat"))
					msg.From = &pb.Endpoint{Field: []byte("test")}
					err := hypernet.Chat("whoareyou", msg)
					Expect(err).NotTo(BeNil())
				})
			})
		})

		Context("to disconnect", func() {
			It("should disconnect successfully", func() {
				err := hypernet.DisConnect("node1")
				Expect(err).To(BeNil())
			})
		})

		Context("to stop hypernet", func() {
			It("should stop successfully", func() {
				hypernet.Stop()
			})
		})
	})
})
