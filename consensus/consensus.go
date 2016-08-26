package consensus

import (
	"hyperchain-alpha/consensus/pbft"
	"hyperchain-alpha/event"

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
func NewConsenter(id uint64) Consenter {
	plugin := "pbft"
	logger.Infof("Creating consensus plugin %s", plugin)
	//var msgQ *event.TypeMux
	eventmux:=new(event.TypeMux)
	eventmux.Post(event.ConsensusEvent{[]byte{0x00, 0x00, 0x03, 0xe8}})
	return pbft.GetPlugin(id, msgQ)
}