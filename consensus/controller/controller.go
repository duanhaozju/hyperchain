package controller

import (
	"github.com/op/go-logging"

	"hyperchain-alpha/consensus"
	"hyperchain-alpha/consensus/pbft"
)

var logger *logging.Logger // package-level logger

func init() {
	logger = logging.MustGetLogger("consensus/controller")
}

// NewConsenter constructs a Consenter object if not already present
func NewConsenter(stack consensus.Stack) consensus.Consenter {

	plugin := "pbft"
	logger.Infof("Creating consensus plugin %s", plugin)
	return pbft.GetPlugin(stack)
}