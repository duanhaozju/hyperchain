package msg

import "time"

type Msg struct {
	from *endpoint
	to *endpoint
	payload []byte
	timestamp int64
}

//NewMsg return a hyperNet inner message
func NewMsg(namespace string,from ,to int,payload []byte) *Msg{
	return &Msg{
		from:newEndpoint(namespace,from),
		to:newEndpoint(namespace,to),
		payload:payload,
		timestamp:time.Now().UnixNano(),
	}
}


func newEndpoint(namespace string,id int) *endpoint{
	return &endpoint{
		namespace:namespace,
		id:id,
	}
}

type endpoint struct {
	namespace string
	id int
}