package network

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testCase struct {
	src      string
	want     string
	wantport int
}

var _ = Describe("DNS", func() {
	var testcases1 = []testCase{
		{"127.0.0.1:8081", "127.0.0.1", 8081},
		{"localhost:8081", "localhost", 8081},
		{"172.16.10.10:12121", "172.16.10.10", 12121},
	}
	_ = []testCase{
		{"127.0.0.18081", "", 0},
		{"hellow8081", "", 0},
		{"hellow:whwe", "", 0},
	}

	Describe("parse the dns item", func() {
		Context("with some correct items", func() {
			It("should parse successfully", func() {
				for _, kase := range testcases1 {
					hostname, port, err := ParseRoute(kase.src)
					Expect(hostname).Should(Equal(kase.want))
					Expect(port).Should(Equal(kase.wantport))
					Expect(err).To(BeNil())
				}
			})
		})

		Context("with incorrect item", func() {
			It("should parse successfully", func() {
				_, _, err := ParseRoute("127.0.0.1")
				Expect(err).Should(Equal(errInvalidRoute))

				_, _, err = ParseRoute("127.0.0.1:aa")
				Expect(err).NotTo(BeNil())
			})
		})
	})
})
