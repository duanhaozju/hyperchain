package consensus

import (
	"hyperchain-alpha/consensus/pbft"
	"hyperchain-alpha/event"
	"hyperchain-alpha/consensus/helper"
	"github.com/op/go-logging"
	"hyperchain-alpha/consensus/events"
)


// Consenter is used to receive messages from the network
// Every consensus plugin needs to implement this interface
type Consenter interface {
	RecvMsg(e events.Event) error // Called serially with incoming messages from gRPC
}

var logger *logging.Logger // package-level logger

func init() {
	logger = logging.MustGetLogger("consensus/controller")
}

func NewConsenter(id uint64) Consenter {
	plugin := "pbft"
	logger.Infof("Creating consensus plugin %s", plugin)
	msgQ :=new(event.TypeMux)
	h:=helper.NewHelper(msgQ)
	return pbft.GetPlugin(id, h)
}