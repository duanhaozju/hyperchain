package types

import "hyperchain-alpha/rlp"

type Msg struct {
	Type       uint64
	Size uint32

	Payload  []byte

}
func (msg *Msg) Decode(val interface{}) error {
	s := rlp.NewStream(msg.Payload, uint64(msg.Size))
	if err := s.Decode(val); err != nil {
		panic("invalid error code")
	}
	return nil
}