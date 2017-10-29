package network

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"github.com/hyperchain/hyperchain/common"
)

func TestNetwork(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Network Suite")
}

var _ = BeforeSuite(func() {
	Describe("Network module", func() {
		logger = common.GetLogger("", "network")
		fmt.Println("start to run network module test...")
	})
})