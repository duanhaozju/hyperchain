package msg

import "hyperchain/p2p/message"

type MsgSlot struct {
	MsgType message.Message_MsgType
	handler MsgHandler
}
