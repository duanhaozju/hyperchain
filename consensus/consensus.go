package consensus

import (
	"hyperchain-alpha/consensus/pbft"
	"hyperchain-alpha/event"
	"hyperchain-alpha/consensus/helper"
	"github.com/op/go-logging"
)


// Consenter is used to receive messages from the network
// Every consensus plugin needs to implement this interface
type Consenter interface {
	RecvMsg(msg []byte) error // Called serially with incoming messages from gRPC
}

var logger *logging.Logger // package-level logger

func init() {
	logger = logging.MustGetLogger("consensus/controller")
}

// NewConsenter constructs a Consenter object if not already present
func (consenter *Consenter) Start(id uint64) Consenter {
	plugin := "pbft"
	logger.Infof("Creating consensus plugin %s", plugin)
	msgQ := new(event.TypeMux)
	h := helper.NewHelper(msgQ)
	return pbft.GetPlugin(id, h)
}