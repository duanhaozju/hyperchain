package hyperdb

import (
	"testing"
	"fmt"
)

func TestInitDatabase(t *testing.T){
	InitDatabase("../config/db.yaml","8001")
	fmt.Println(logPath)
	fmt.Println(logStatus)
	fmt.Println(dbType)
	fmt.Println(ssdbProxyPort)
	fmt.Println(ssdbFirstPort)
	fmt.Println(ssdbServerNumber)
	fmt.Println(ssdbPoolSize)
	fmt.Println(ssdbTimeout)

	fmt.Println(ssdbMaxConnectTimes)
	fmt.Println(redisPort)
	fmt.Println(redisPoolSize)
	fmt.Println(redisTimeout)

	fmt.Println(redisMaxConnectTimes)
	fmt.Println(leveldbPath)
	fmt.Println(grpcPort)

}