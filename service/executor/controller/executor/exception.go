// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package executor

import (
	"fmt"

	"github.com/hyperchain/hyperchain/manager/appstat"
	"github.com/hyperchain/hyperchain/manager/event"
)

// NotifyViewChange posts the ViewChange to FilterSystemStatusEvent.
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
