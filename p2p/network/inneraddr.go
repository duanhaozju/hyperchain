package network

import (
	"hyperchain/common"
	"github.com/pkg/errors"
	"github.com/terasum/viper"
	"fmt"
	"strings"
	"sync"
	"encoding/json"
)

type InnerAddr struct {
	//domain ipaddr
	addrs map[string]string
	lock *sync.RWMutex
}

func NewInnerAddr()*InnerAddr{
	return &InnerAddr{
		addrs:make(map[string]string),
		lock:new(sync.RWMutex),
	}
}

//Get the domain matched ipaddr, if not exist return ""
func(ia *InnerAddr)Get(domain string) (string){
	ia.lock.RLock()
	defer ia.lock.RUnlock()
	if ipaddr,ok := ia.addrs[domain];ok{
		return ipaddr
	}
	if ipaddr,ok:=ia.addrs["default"];ok{
		return ipaddr
	}
	// if not exist return first one
	for _,v := range ia.addrs{
		return v
	}
	return ""
}

func(ia *InnerAddr)Add(domain,ipaddr string){
	ia.lock.Lock()
	defer  ia.lock.Unlock()
	ia.addrs[domain] = ipaddr
}

func(ia *InnerAddr)Del(domain,ipaddr string){
	ia.lock.Lock()
	defer  ia.lock.Unlock()
	delete(ia.addrs,domain)
}

func (ia *InnerAddr)Serialize()([]byte,error){
	ia.lock.RLock()
	defer ia.lock.RUnlock()
	return json.Marshal(ia.addrs)
}

func InnerAddrUnSerialize(raw []byte)(*InnerAddr,error){
	tempMap := make(map[string]string)
	err := json.Unmarshal(raw,&tempMap)
	if err !=nil{
		return nil,err
	}
	return &InnerAddr{
		addrs:tempMap,
		lock:new(sync.RWMutex),
	},nil
}


func GetInnerAddr(addrFile string) (*InnerAddr,string,error){
	if !common.FileExist(addrFile){
		return nil,"default",errors.New("the addr file not exist!")
	}
	vip := viper.New()
	vip.SetConfigFile(addrFile)
	err := vip.ReadInConfig()
	if err != nil{
		return nil,"default",err
	}

	addrs := NewInnerAddr()

	items := vip.GetStringSlice("addrs")
	for _,item := range items{
		temp_items := strings.Split(item," ")
		if len(temp_items) !=2{
			return nil,"default",errors.New(fmt.Sprintf("illegal domain addr (%s)",item))
		}
		temp_domain := temp_items[0]
		//todo check the ip format
		temp_ipaddr := temp_items[1]
		addrs.Add(temp_domain,temp_ipaddr)
	}
	domain := vip.GetString("domain")
	return addrs,domain,nil
}


