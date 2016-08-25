package peerEventHandler

import (
	"hyperchain-alpha/p2p/peermessage"
)

type PeerEventHandler interface {
	ProcessEvent(*peermessage.Message) error
}
