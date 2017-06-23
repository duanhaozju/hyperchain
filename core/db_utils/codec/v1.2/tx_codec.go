package v1_2

import (
	"hyperchain/core/types"
	"github.com/golang/protobuf/proto"
)


type TxCodec struct {}

func (codec *TxCodec) Encode(tx *types.Transaction) ([]byte, error) {
	tx.Version = []byte(TransactionVersion)
	return proto.Marshal(tx)
}
