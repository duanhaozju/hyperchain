package executor
import (
	"hyperchain/event"
)
type Helper struct {
	msgQ *event.TypeMux
}

func NewHelper(msgQ *event.TypeMux) *Helper {
	return &Helper{
		msgQ: msgQ,
	}
}

func (executor *Executor) informConsensus(informType string, message interface{}) {
	switch informType {
	case CONSENSUS_LOCAL:
		executor.consenter.RecvLocal(message)
	default:

	}
}

