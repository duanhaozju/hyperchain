package msg

import (
	"fmt"
	"github.com/hyperchain/hyperchain/p2p/message"
	"github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"
	"strconv"
)

type MsgSlot struct {
	// this is map of map[message.MsgType]msg.handler
	slot cmap.ConcurrentMap
}

func NewMsgSlot() *MsgSlot {
	return &MsgSlot{
		slot: cmap.New(),
	}
}

func (slot *MsgSlot) Register(msgType message.MsgType, handler MsgHandler) {
	slot.slot.SetIfAbsent(strconv.FormatInt(int64(msgType), 10), handler)
}

func (slot *MsgSlot) DeRegister(msgType message.MsgType) {
	slot.slot.Remove(strconv.FormatInt(int64(msgType), 10))
}

func (slot *MsgSlot) Clear() {
	for _, key := range slot.slot.Keys() {
		slot.slot.Remove(key)
	}
}

func (slot *MsgSlot) GetHandler(msgType message.MsgType) (MsgHandler, error) {
	h, ok := slot.slot.Get(strconv.FormatInt(int64(msgType), 10))
	if !ok {
		return nil, errors.New(fmt.Sprintf("Cannot find the msg handler,message type: %s", msgType.String()))
	}
	return h.(MsgHandler), nil
}

type MsgSlots struct {
	//this is map of  map[filed string]MsgSlot
	slots cmap.ConcurrentMap
}

func NewMsgSlots() *MsgSlots {
	return &MsgSlots{
		slots: cmap.New(),
	}
}

func (slots *MsgSlots) Register(filed string, slot *MsgSlot) {
	slots.slots.SetIfAbsent(filed, slot)
}

func (slots *MsgSlots) GetSlot(filed string) (*MsgSlot, error) {
	s, ok := slots.slots.Get(filed)
	if !ok {
		return nil, errors.New(fmt.Sprintf("cannot find the filed msg slot, filed: %s", filed))
	}
	return s.(*MsgSlot), nil
}

func (slots *MsgSlots) DeRegister(filed string) {
	slots.slots.Remove(filed)
}
