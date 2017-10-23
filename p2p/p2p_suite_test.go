package p2p_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestP2P(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "P2P Suite")
}

var _ = BeforeSuite(func() {
	Describe("P2P module", func() {
		fmt.Println("start to run p2p module test...")
	})
})
