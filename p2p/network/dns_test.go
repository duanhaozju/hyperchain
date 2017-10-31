package network

import (
	"github.com/hyperchain/hyperchain/p2p/utils"
	"github.com/hyperchain/hyperchain/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"fmt"
)

var _ = Describe("DNS", func() {
	Describe("create a new DNSResolver", func() {
		logger = common.GetLogger(common.DEFAULT_NAMESPACE, "network")
		Context("with a correct config file", func() {
			var dnsr *DNSResolver
			It("should create successfully", func() {
				var err error
				dnsr, err = NewDNSResolver(utils.GetProjectPath() + "/p2p/test/hosts.toml")
				Expect(err).To(BeNil())
				Expect(len(dnsr.DNSItems)).Should(Equal(4))
			})

			Context("to add a item", func() {
				It("should add successfully", func() {
					err := dnsr.AddItem("node5", "127.0.0.1:50015", false)
					Expect(err).To(BeNil())
				})

				Context("which existed", func() {
					It("should add failed", func() {
						err := dnsr.AddItem("node4", "127.0.0.1:50014", false)
						Expect(err).To(BeNil())
					})
				})
			})

			Context("to get a item", func() {
				It("should get successfully", func() {
					addr, err :=  dnsr.GetDNS("node5")
					Expect(err).To(BeNil())
					Expect(len(dnsr.DNSItems)).Should(Equal(5))
					Expect(addr).Should(Equal("127.0.0.1:50015"))
				})
			})

			Context("to cover a existing item", func() {
				It("should cover successfully", func() {
					err := dnsr.AddItem("node4", "127.0.0.1:50018", true)
					Expect(err).To(BeNil())
					addr, err :=  dnsr.GetDNS("node4")
					Expect(err).To(BeNil())
					Expect(addr).Should(Equal("127.0.0.1:50018"))
				})
			})

			Context("to delete a item", func() {
				It("should delete successfully", func() {
					err := dnsr.DelItem("node5")
					Expect(err).To(BeNil())
					Expect(len(dnsr.DNSItems)).Should(Equal(4))
				})
			})

			Context("to list host and hostname", func() {
				It("should list successfully", func() {
					hosts := dnsr.ListHosts()
					Expect(len(hosts)).Should(Equal(4))
					By(fmt.Sprintf("%v", hosts))

					names := dnsr.listHostnames()
					Expect(len(names)).Should(Equal(4))
					By(fmt.Sprintf("%v", names))
				})
			})
		})
	})
})