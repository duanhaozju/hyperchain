package executor
import (
	"hyperchain/event"
)
type Helper struct {
	msgQ *event.TypeMux
}

func NewHelper(msgQ *event.TypeMux) *Helper {
	return &Helper{
		msgQ: msgQ,
	}
}

