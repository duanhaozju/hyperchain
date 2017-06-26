package message

import (
	"time"
	"strconv"
	"golang.org/x/crypto/sha3"
)

func NewMsg(msgType MsgType,payload []byte) *Message{
	return &Message{
		MessageType:msgType,
		Payload:payload,
		TimeStamp:time.Now().UnixNano(),
	}

}

func (msg *Message)Sign()error{
	//todo implements this
	//hash := msg.hash()
	return nil
}

func(msg *Message)hash()[]byte{
	//todo here should check the msg data filed is nil or not
	hasher := sha3.New256()
	hasher.Write([]byte(strconv.FormatInt(int64(msg.MessageType),10)))
	hasher.Write([]byte(strconv.FormatInt(int64(msg.TimeStamp),10)))
	hasher.Write(msg.Payload)
	hasher.Write(msg.From.Hostname)
	hasher.Write(msg.From.Field)
	hasher.Write(msg.From.UUID)
	hash := hasher.Sum([]byte(strconv.FormatInt(int64(msg.From.Version),16)))
	return hash
}
