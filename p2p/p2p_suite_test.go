package p2p_test

import (
	"testing"
	"fmt"
)

func TestP2p(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "P2P Suite")
}

var _ = BeforeSuite(func() {
	Describe("P2P module", func() {
		fmt.Println("start p2p module test")
	})
});