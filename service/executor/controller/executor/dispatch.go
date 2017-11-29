package executor

import (
	"github.com/gogo/protobuf/proto"
	"github.com/hyperchain/hyperchain/core/oplog/proto"
	"github.com/hyperchain/hyperchain/manager/event"
)

//Dispatch receive oplog in random order, but should dispatch them in order
func (e *Executor) Dispatch(ol *oplog.LogEntry) {
	e.logger.Debugf("dispatch log id:%d", ol.Lid)
	if ol.Type == oplog.LogEntry_RollBack { //TODO(Xiaoyi Wang): this have the highest priority ?

	} else {
		e.cache.opLogC <- ol
	}
}

//dispatch oplog sequentially
func (e *Executor) sequentialDispatch() {
	counter := 0
	for {
		select {
		case <-e.context.exit:
			e.logger.Notice("dispatch oplog thread exit")
			return
		case ol := <-e.cache.opLogC:
			e.logger.Debugf("fetch log with id %d", ol.Lid)
			demandIndex := e.context.getDemandOpLogIndex()
			if ol.Lid == 10 {
				break
			}
			if ol.Lid == demandIndex {
				counter = 0
				e.dispatch(ol)
				e.context.setDemandOpLogIndex(demandIndex + 1)
				e.dispatchPendingOpLogs()
			} else if ol.Lid > demandIndex {
				counter += 1
				e.logger.Debugf("log id %d is bigger than demandIndex %d", ol.Lid, demandIndex)
				e.cache.pendingOpLogs.Add(ol.Lid, ol)
				if counter >= 10 {
					e.logger.Infof("fetchLogEntry with log id %v", demandIndex)
					logEntry := e.helper.fetchLogEntry(demandIndex)
					if logEntry != nil {
						if logEntry.Lid == demandIndex {
							e.dispatch(logEntry)
							e.context.setDemandOpLogIndex(demandIndex + 1)
							e.dispatchPendingOpLogs()
						} else {
							e.logger.Errorf("fetch log with %d but %d", demandIndex, logEntry.Lid)
						}
					}
					counter = 0
				}
			} else {
				e.logger.Criticalf("log entry id %d is less than demand index %d, discard this log %v",
					ol.Lid, demandIndex, ol)
			}
		}
	}
}

func (e *Executor) dispatchPendingOpLogs() {
	if e.cache.pendingOpLogs.Len() > 0 {
		for i := e.context.getDemandOpLogIndex(); e.cache.pendingOpLogs.Contains(i); i++ {
			if o, ok := e.cache.pendingOpLogs.Get(i); ok {
				if ol, ok := o.(*oplog.LogEntry); ok {
					e.dispatch(ol)
					e.context.setDemandOpLogIndex(i + 1)
					e.cache.pendingOpLogs.RemoveWithCond(e.context.demandOpLogIndex-1, RemoveLessThan)
				} else {
					e.cache.pendingOpLogs.RemoveWithCond(e.context.demandOpLogIndex, RemoveLessThan)
					break
				}
			} else {
				e.logger.Errorf("get pending oplog with id %v error", i)
				break
			}
		}
	}
}

func (e *Executor) dispatch(ol *oplog.LogEntry) {
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
	ve := &event.ValidationEvent{}
	if err := proto.Unmarshal(txs.Payload, ve); err != nil {
		e.logger.Error(err)
	}
	e.cache.opLogIndexCache.Add(ve.SeqNo, txs.Lid)
	e.Validate(ve)
}

func (e *Executor) processRollBack(rb *oplog.LogEntry) {
	//TODO(Xiaoyi Wang): add rollback logic
}

func (e *Executor) processStateUpdate(su *oplog.LogEntry) {
	//TODO(Xiaoyi Wang): add state update logic
}
