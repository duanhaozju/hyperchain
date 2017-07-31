package common

import (
	"github.com/cheggaaa/pb"
	"github.com/op/go-logging"
)

func InitPb(total int64, prefix string) *pb.ProgressBar {
	processBar := pb.New64(total)
	processBar.NotPrint = true
	processBar.ManualUpdate = true
	processBar.Prefix(prefix)
	return processBar
}

func AddPb(processBar *pb.ProgressBar, add int64) {
	processBar.Add64(add)
}

func SetPb(processBar *pb.ProgressBar, current int64) {
	processBar.Set64(current)
}

func PrintPb(processBar *pb.ProgressBar, interval int64, log *logging.Logger) {
	if interval == 0 {
		processBar.Update()
		log.Notice(processBar.String())
	} else if processBar.Get()%interval == 0 {
		processBar.Update()
		log.Notice(processBar.String())
	}
}

func IsPrintPb(processBar *pb.ProgressBar, interval int64) bool {
	if processBar.Get()%interval == 0 {
		return true
	} else {
		return false
	}
}
