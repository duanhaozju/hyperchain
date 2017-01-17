package hyperdb

import "os"
import "time"
import "strconv"

func writeLog(funcName string, num int, err error) {
	f, err1 := os.OpenFile(GetLogPath(), os.O_WRONLY|os.O_CREATE, 0644)
	if err1 != nil {
		log.Notice(GetLogPath() + " file create failed. err: " + err.Error())
	} else {
		n, _ := f.Seek(0, os.SEEK_END)
		currentTime := time.Now().Local()
		newFormat := currentTime.Format("2006-01-02 15:04:05.000")
		str := strconv.Itoa(grpcPort) + newFormat + funcName + err.Error() + " num:" + strconv.Itoa(num) + "\n"
		_, err1 = f.WriteAt([]byte(str), n)
		f.Close()
	}
}

func Bytes(reply interface{}) []byte {

	switch reply := reply.(type) {
	case []byte:
		return reply
	case string:
		return []byte(reply)
	case nil:
		return nil
	}
	return nil
}
