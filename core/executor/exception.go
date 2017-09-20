package executor

import (
	"fmt"
	"hyperchain/manager/appstat"
	"hyperchain/manager/event"
)

func NotifyViewChange(helper *Helper, seqNo uint64) {
	message := fmt.Sprintf("required viewchange to %d", seqNo)
	helper.PostExternal(event.FilterSystemStatusEvent{
		Module:    appstat.ExceptionModule_Executor,
		Status:    appstat.Exception,
		Subtype:   appstat.ExceptionSubType_ViewChange,
		ErrorCode: appstat.ExceptionCode_Executor_Viewchange,
		Message:   message,
	})
}
