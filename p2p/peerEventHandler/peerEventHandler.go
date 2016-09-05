// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 14:02
// last Modified Author: chenquan
// change log: 1. add a general comment of the PeerEventHandler interface
//
package peerEventHandler

import (
	"hyperchain/p2p/peermessage"
)
// PeerEventHandler interface define
type PeerEventHandler interface {
	ProcessEvent(*peermessage.Message) error
}
