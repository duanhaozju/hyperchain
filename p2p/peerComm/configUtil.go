// author: chenquan
// date: 16-9-20
// last modified: 16-9-20 13:09
// last Modified Author: chenquan
// change log: 
//		
package peerComm

import "strconv"

type Config interface{
	GetPort(nodeId int) int
	GetIP(nodeId int) string
	GetMaxPeerNumber() int
	GetCname(nodeId int) string
}


type ConfigUtil struct{
	configs map[string]string

}

func NewConfigUtil(configpath string) *ConfigUtil{
	var newConfigUtil ConfigUtil
	newConfigUtil.configs = GetConfig(configpath)
	return &newConfigUtil
}

func (this *ConfigUtil)GetPort(nodeId int) int{
	port,err := strconv.Atoi(this.configs["port"+strconv.Itoa(nodeId)])
	if err !=nil{
		log.Error("cannot convert the port config")
		return -1
	}
	return port
}
func (this *ConfigUtil)GetIP(nodeId int) string{
	ip := this.configs["node"+strconv.Itoa(nodeId)]

	return ip
}

func (this *ConfigUtil)GetMaxPeerNumber() int{
	maxpeers,err := strconv.Atoi(this.configs["MAXPEERS"])
	if err != nil{
		log.Error("connot convert the MAXPEERS")
		return -1
	}
	return maxpeers

}
func (this *ConfigUtil) GetCname(nodeId int) string{
	cname := this.configs["cname"+strconv.Itoa(nodeId)]
	return cname
}