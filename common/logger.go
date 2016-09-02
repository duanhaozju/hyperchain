// author: chenquan
// date: 16-9-2
// last modified: 16-9-2 13:59
// last Modified Author: chenquan
// change log: 
//		
package common

import (
	"github.com/Sirupsen/logrus"
	"path"
	"runtime"
	"strings"
)

type ContextHook struct{}

func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook ContextHook) Fire(entry *logrus.Entry) error {
	pc := make([]uintptr, 3, 3)
	cnt := runtime.Callers(6, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()
		if !strings.Contains(name, "github.com/Sirupsen/logrus") {
			file, line := fu.FileLine(pc[i] - 1)
			entry.Data["file"] = path.Base(file)
			entry.Data["func"] = path.Base(name)
			entry.Data["line"] = line
			//entry.Data["time"] = time.Now().Format("15:04:05")
			break
		}
	}
	return nil
}


func LoggerInit(){
	logrus.AddHook(ContextHook{})
	//logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
}
