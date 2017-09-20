package p2p_test

import (
	"fmt"
	"testing"
)

func TestP2p(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "P2P Suite")
}

var _ = BeforeSuite(func() {
	Describe("P2P module", func() {
		fmt.Println("start p2p module test")
	})
})
