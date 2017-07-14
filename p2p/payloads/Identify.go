package payloads

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/p2p/utils"
)

func NewIdentify(isvp bool,isOriginal bool,namespace string, hostname string, id int) *Identify {
	iden := &Identify{
		Id:       int64(id),
		IsVP:     isvp,
		IsOriginal:isOriginal,
		Hostname: hostname,
		Namespace:namespace,
	}
	if isvp{
		iden.Hash = utils.HashString(hostname + namespace)
	}else{
		iden.Hash = utils.HashString(hostname + namespace)
	}
	return iden
}

func (id *Identify) Serialize() ([]byte, error) {
	return proto.Marshal(id)
}

func IdentifyUnSerialize(raw []byte) (*Identify, error) {
	id := new(Identify)
	err := proto.Unmarshal(raw, id)
	return id, err
}
