//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

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
