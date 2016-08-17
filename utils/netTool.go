package utils

import (
	"net"
	"fmt"
	"strings"
)

func GetIpAddr() string{
	conn, err := net.Dial("udp", "baidu.com:80")
	if err != nil {
		fmt.Println(err.Error())
		return "错误"
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0]
}