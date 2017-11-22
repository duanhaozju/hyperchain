package executor

import "github.com/hyperchain/hyperchain/core/oplog/proto"

func (e *Executor) Dispatch(ol *oplog.LogEntry) {
	if ol == nil {
		e.logger.Errorf("Invalid oplog entry %v", ol)
	}
	switch ol.Type {
	case oplog.LogEntry_TransactionList:
		e.processTransactions(ol)
	case oplog.LogEntry_RollBack:
		e.processRollBack(ol)
	case oplog.LogEntry_StateUpdate:
		e.processStateUpdate(ol)
	default:
		e.logger.Errorf("Invalid oplog type, %v", ol.Type)
	}
}

func (e *Executor) processTransactions(txs *oplog.LogEntry) {
	if txs.Payload == nil || len(txs.Payload) == 0 {
		e.logger.Error("Op log payload is nil")
	}
}

func (e *Executor) processRollBack(rb *oplog.LogEntry) {

}

func (e *Executor) processStateUpdate(su *oplog.LogEntry) {

}
