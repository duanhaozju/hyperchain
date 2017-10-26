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

// This package is a simple wrap for opensource prgressbar util.
package common

import (
	"github.com/cheggaaa/pb"
	"github.com/op/go-logging"
)

// InitPb initializes a progress bar to display how well this process continues,
// and total is the total workload.
func InitPb(total int64, prefix string) *pb.ProgressBar {
	processBar := pb.New64(total)
	processBar.NotPrint = true
	processBar.ManualUpdate = true
	processBar.Prefix(prefix)
	return processBar
}

// AddPb makes the progress bar move 'add' steps forward
func AddPb(processBar *pb.ProgressBar, add int64) {
	processBar.Add64(add)
}

// SetPb sets the progress bar to 'current' place
func SetPb(processBar *pb.ProgressBar, current int64) {
	processBar.Set64(current)
}

// PrintPb prints the progress bar once in a while
func PrintPb(processBar *pb.ProgressBar, interval int64, log *logging.Logger) {
	if interval == 0 {
		processBar.Update()
		log.Notice(processBar.String())
	} else if processBar.Get()%interval == 0 {
		processBar.Update()
		log.Notice(processBar.String())
	}
}

// IsPrintPb tells if this progress bar gets to the end
func IsPrintPb(processBar *pb.ProgressBar, interval int64) bool {
	if processBar.Get()%interval == 0 {
		return true
	} else {
		return false
	}
}
