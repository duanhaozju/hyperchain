// author: chenquan
// date: 16-8-25
// last modified: 16-8-25 20:01
// last Modified Author: chenquan
// change log:
//
package peerComm

import (
	"net"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"github.com/labstack/gommon/log"
)

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
				fmt.Println(ipnet.IP.String())
				return ipnet.IP.String()
			}

		}
	}
	return "localhost"
}

func GetConfig(path string) map[string]interface{}{
	content,fileErr := ioutil.ReadFile(path)
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	var configs map[string]interface{}
	UmErr := json.Unmarshal(content,&configs)
	if UmErr != nil {
		log.Fatal(UmErr)
	}
	return configs
}