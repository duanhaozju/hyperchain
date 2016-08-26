// author: chenquan
// date: 16-8-25
// last modified: 16-8-25 20:01
// last Modified Author: chenquan
// change log:
//
package peerEventHandler

import (
	"hyperchain-alpha/p2p/peermessage"
)

type PeerEventHandler interface {
	ProcessEvent(*peermessage.Message) error
}
