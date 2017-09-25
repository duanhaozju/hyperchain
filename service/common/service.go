package common

//Service interface to be implemented by component.
type Service interface {
	Namespace() string
	Id() string                // service identifier.
	Send(msg interface{})      // sync send msg.
	AsyncSend(msg interface{}) // async send msg.
	Close()
}

type serviceImpl struct {

}

func (si serviceImpl) Namespace() string  {
	return ""
}

// Id service identifier.
func (si *serviceImpl) Id() string {
	return ""
}

// Send sync send msg.
func (si *serviceImpl) Send(msg interface{}) {

}

// AsyncSend async send msg.
func (si *serviceImpl) AsyncSend(msg interface{}) {

}