package manager

import (
	"hyperchain/consensus/pbft"
	"github.com/golang/protobuf/proto"
	m "hyperchain/manager/message"
)

func (hub *EventHub) invokePbftLocal(serviceType, eventType int, content interface{}) {
	e := &pbft.LocalEvent{
		Service:   serviceType,
		EventType: eventType,
		Event:     content,
	}
	hub.consenter.RecvLocal(e)
}

func (hub *EventHub) broadcast(bType int, t m.SessionMessage_Type, message []byte) {
	hub.logger.Debugf("broadcast session message %s", t.String())
	if ctx, err := proto.Marshal(&m.SessionMessage{
		Type:    t,
		Payload: message,
	}); err != nil {
		hub.logger.Errorf("marshal message %d failed.", t)
		return
	} else {
		switch bType {
		case BROADCAST_NVP:
			hub.peerManager.BroadcastNVP(ctx)
		case BROADCAST_VP:
			fallthrough
		case BROADCAST_ALL:
			hub.peerManager.Broadcast(ctx)
		}
	}
}

func (hub *EventHub) send(t m.SessionMessage_Type, message []byte, peers []uint64) {
	hub.logger.Debugf("send session message %s", t.String())
	if ctx, err := proto.Marshal(&m.SessionMessage{
		Type:    t,
		Payload: message,
	}); err != nil {
		hub.logger.Errorf("marshal message %d failed.", t)
		return
	} else {
		hub.peerManager.SendMsg(ctx, peers)
	}
}

func (hub *EventHub) sendToRandomVP(t m.SessionMessage_Type, message []byte) {
	hub.logger.Debugf("send session message %s", t.String())
	if ctx, err := proto.Marshal(&m.SessionMessage{
		Type:    t,
		Payload: message,
	}); err != nil {
		hub.logger.Errorf("marshal message %d failed.", t)
		return
	} else {
		hub.peerManager.SendRandomVP(ctx)
	}
}

func (hub *EventHub) sendToNVP(t m.SessionMessage_Type, message []byte, peers []string) {
	hub.logger.Debugf("send session message %s", t.String())
	if ctx, err := proto.Marshal(&m.SessionMessage{
		Type:    t,
		Payload: message,
	}); err != nil {
		hub.logger.Errorf("marshal message %d failed.", t)
		return
	} else {
		hub.peerManager.SendMsgNVP(ctx, peers)
	}
}