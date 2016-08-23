package manager

import (


	"fmt"

	"hyperchain-alpha/core/types"
)
//远程节点
type peer struct {
	id string
	msg types.Msg

}



func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

func Handshake()error{
	//TODO 握手
	return nil
}

func (pm *ProtocolManager) handle(p *peer) error {



	if err := Handshake(); err != nil {

		return err
	}





	// main loop. handle incoming messages.
	for {
		if err := pm.handleMsg(p); err != nil {

			return err
		}
	}
}


//接受远程来的消息
func (pm *ProtocolManager) handleMsg(p *peer) error {


	switch {
	case p.msg.Type == StatusMsg:

		return errResp(ErrExtraStatusMsg, "uncontrolled status message")




	case p.msg.Type == NewBlockMsg:

		var request newBlockData
		if err := p.msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", p.msg, err)
		}
		pm.fetcher.Enqueue(p.id, request.Block)


	case p.msg.Type == TxMsg:

		var txs []*types.Transaction
		if err := p.msg.Decode(&txs); err != nil {
			return errResp(ErrDecode, "msg %v: %v", p.msg, err)
		}
		for i, tx := range txs {

			if tx == nil {
				return errResp(ErrDecode, "transaction %d is nil", i)
			}

		}
		pm.txpool.AddTransactions(txs)

	default:
		return errResp(ErrInvalidMsgCode, "%v", p.msg.Type)
	}
	return nil
}