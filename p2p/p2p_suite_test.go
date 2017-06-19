package p2p_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"fmt"
)

func TestP2p(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "P2p Suite")
}

var _ = BeforeSuite(func(){
	Describe("P2P module",func(){
		fmt.Println("start p2p module test")
	})
});