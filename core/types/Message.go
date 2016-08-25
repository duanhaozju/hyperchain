package types

import (
	"io"

)

const (
	// Protocol messages
	StatusMsg                   = 0x00

	TxMsg                       = 0x01

	NewBlockMsg                 = 0x02

)

type Msg struct {
	Type    uint64
	Size    uint32

	Payload io.Reader
}







