package payloads

import "github.com/golang/protobuf/proto"

func NewIdentify(isvp bool,isOriginal bool,namespace string, hostname string, id int) *Identify {
	return &Identify{
		Id:       int64(id),
		IsVP:     isvp,
		IsOriginal:isOriginal,
		Hostname: hostname,
		Namespace:namespace,
	}
}

func (id *Identify) Serialize() ([]byte, error) {
	return proto.Marshal(id)
}

func IdentifyUnSerialize(raw []byte) (*Identify, error) {
	id := new(Identify)
	err := proto.Unmarshal(raw, id)
	return id, err
}
