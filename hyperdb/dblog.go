package hyperdb

import (
	"fmt"
	"time"
)

func DBLog(log string)  {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05") , " ", log)
}
