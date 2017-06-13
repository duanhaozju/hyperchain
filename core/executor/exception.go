package executor

import (
	"hyperchain/manager/event"
	"time"
)

const (
	ExceptionSubType_ViewChange      = "viewchange"
	ExceptionSubType_CutdownBlock    = "cutdownBlock"
)

type ExceptionHandler interface {
	Throw(typ, msg string)
}

type ExceptionHandlerImpl struct {
	helper   *Helper
}

func NewExceptionHandler(helper *Helper) *ExceptionHandlerImpl {
	return &ExceptionHandlerImpl{
		helper:   helper,
	}
}

func (impl *ExceptionHandlerImpl) Throw(typ, msg string) {
	ev := event.FilterException{
		Module:       event.ExceptionModule_Executor,
		SubType:      typ,
		ErrorCode:    event.ExceptionCode_Executor_Viewchange,
		Message:      msg,
		Date:         time.Now(),
	}
	impl.helper.PostExternal(ev)
}
