package manager

import (

	"sync/atomic"
	"io"
	"time"
	"github.com/ethereum/go-ethereum/core/types"
	"hyperchain-alpha/rlp"
	"fmt"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/logger"
)
//远程节点
type peer struct {
	id string

}

type Msg struct {
	Code       uint64
	Size       uint32 // size of the paylod
	Payload    io.Reader
	ReceivedAt time.Time
}
func (msg Msg) Decode(val interface{}) error {
	s := rlp.NewStream(msg.Payload, uint64(msg.Size))
	if err := s.Decode(val); err != nil {
		panic("invalid error code")
	}
	return nil
}
func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

func Handshake()error{
	//TODO 握手
	return nil
}

func (pm *ProtocolManager) handle(p *peer,msg Msg) error {


	// Execute the Ethereum handshake

	if err := Handshake(); err != nil {
		//glog.V(logger.Info).Infof("enter handshake failed")
		glog.V(logger.Debug).Infof("%v: handshake failed: %v", p, err)
		return err
	}


	// Propagate existing transactions. new transactions appearing
	// after this will be sent via broadcasts.
	pm.syncTransactions(p)


	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p,msg); err != nil {
			glog.V(logger.Debug).Infof("%v: message handling failed: %v", p, err)
			return err
		}
	}
}


//接受远程来的消息
func (pm *ProtocolManager) handleMsg(p *peer,msg Msg) error {

	// Read the next message from the remote peer, and ensure it's fully consumed


	// Handle the message depending on its contents
	switch {
	case msg.Code == StatusMsg:
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")




	case msg.Code == NewBlockMsg:
		// Retrieve and decode the propagated block
		var request newBlockData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		pm.fetcher.Enqueue(p.id, request.Block)


	case msg.Code == TxMsg:
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		if atomic.LoadUint32(&pm.synced) == 0 {
			break
		}
		// Transactions can be processed, parse all of them and deliver to the pool
		var txs []*types.Transaction
		if err := msg.Decode(&txs); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		for i, tx := range txs {
			// Validate and mark the remote transaction
			if tx == nil {
				return errResp(ErrDecode, "transaction %d is nil", i)
			}

		}
		pm.txpool.AddTransactions(txs)

	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}