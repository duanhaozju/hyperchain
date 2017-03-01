package blockpool
import "hyperchain/event"

func (pool *BlockPool) SuspendForPeerManage() {
	log.Notice("suspend for new peer")
	pool.NotifyValidateToStop()
	pool.WaitResetAvailable()
	pool.NotifyValidateToBegin()
	log.Notice("suspend for new peer done")
	pool.consenter.RecvLocal(event.ResetValidateQDone{})
}
