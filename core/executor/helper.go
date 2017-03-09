package executor
import (
	"hyperchain/event"
	"hyperchain/recovery"
	"github.com/golang/protobuf/proto"
	edb "hyperchain/core/db_utils"
)
type Helper struct {
	msgQ *event.TypeMux
}

func NewHelper(msgQ *event.TypeMux) *Helper {
	return &Helper{
		msgQ: msgQ,
	}
}

// informConsensus - communicate with consensus module.
func (executor *Executor) informConsensus(informType string, message interface{}) error {
	switch informType {
	case CONSENSUS_LOCAL:
		return executor.consenter.RecvLocal(message)
	default:
		return NoDefinedCaseErr
	}
}

// informP2P - communicate with p2p module.
func (executor *Executor) informP2P(msgType string, message interface{}) error {
	switch msgType {
	case P2P_SEND_SYNC_REQ:
		required := &recovery.CheckPointMessage{
			RequiredNumber: executor.status.syncFlag.SyncTarget,
			CurrentNumber:  edb.GetHeightOfChain(executor.namespace),
			PeerId:         executor.status.syncFlag.LocalId,
		}
		payload, err := proto.Marshal(required)
		if err != nil {
			log.Errorf("[Namespace = %s] SendSyncRequest marshal message failed", executor.namespace)
			return err
		}
		executor.peerManager.SendMsgToPeers(payload, executor.status.syncFlag.SyncPeers, recovery.Message_SYNCCHECKPOINT)
	default:
		return NoDefinedCaseErr
	}
}

