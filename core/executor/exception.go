package executor

import (
	"hyperchain/manager/exception"
	"fmt"
	"hyperchain/manager/event"
)

func NotifyViewChange(helper *Helper, seqNo uint64) {
	helper.PostExternal(event.FilterExceptionEvent{
		Module:    exception.ExceptionModule_Executor,
		Exception: exception.ExecutorViewchangeError{Msg: fmt.Sprintf("requried to reset status to %d", seqNo - 1)},
	})
}
