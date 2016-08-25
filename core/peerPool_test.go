package core

import (
	"testing"



	"hyperchain-alpha/core/types"
	"fmt"
	"github.com/golang/protobuf/proto"
)

func TestPeersPool_PutPeer(t *testing.T) {
	tx:=&types.Transaction{Payload:[]byte{0x00, 0x00, 0x03, 0xe8},

	}

	a,err:=proto.Marshal(tx)
	fmt.Print(err)

	fmt.Print(a)
	b:=&types.Transaction{}
	proto.Unmarshal(a,b)
	fmt.Print(b)
}


