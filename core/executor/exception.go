package executor

import (
	"fmt"
	"hyperchain/manager/event"
	"hyperchain/manager/exception"
)

func NotifyViewChange(helper *Helper, seqNo uint64) {
	helper.PostExternal(event.FilterExceptionEvent{
		Module:    exception.ExceptionModule_Executor,
		Exception: exception.ExecutorViewchangeError{Msg: fmt.Sprintf("requried to reset status to %d", seqNo-1)},
	})
}
