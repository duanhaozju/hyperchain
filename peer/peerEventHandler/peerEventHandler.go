package peerEventHandler

import (
	"hyperchain-alpha/peer/peermessage"
)

type PeerEventHandler interface {
	ProcessEvent(*peermessage.Message) error
}
