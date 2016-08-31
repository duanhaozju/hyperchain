// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:58
// last Modified Author: chenquan
// change log: add a comment of this file function
//
package peerComm

import (
	"net"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"github.com/labstack/gommon/log"
	"time"
)
// GetIpLocalIpAddr this function is used to get the real internal net ip address
// to use this make sure your net are valid
func GetIpLocalIpAddr()string{
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		return "localhost"
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				//fmt.Println(ipnet.IP.String())
				return ipnet.IP.String()
			}

		}
	}
	return "localhost"
}
// GetConfig this is a tool function for get the json file config
// configs return a map[string]string
func GetConfig(path string) map[string]string{
	content,fileErr := ioutil.ReadFile(path)
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	var configs map[string]string
	UmErr := json.Unmarshal(content,&configs)
	if UmErr != nil {
		log.Fatal(UmErr)
	}
	return configs
}

func GenUnixTimeStamp() int64{
	return time.Now().Unix()
}