package network

import (
	"github.com/terasum/viper"
	"github.com/pkg/errors"
	"strings"
	"fmt"
	"sync"
)
var (
	errHostnameNotFoud = errors.New("hostname cannot resolved.")
)

type DNSResolver struct {
	hostConfig *viper.Viper
	DNSItems map[string]string
	lock *sync.RWMutex
}

func NewDNSResolver(hostsPath string) (*DNSResolver,error){
	vip := viper.New()
	vip.SetConfigFile(hostsPath)
	err := vip.ReadInConfig()
	if err != nil{
		return nil,err
	}
	dnsr := &DNSResolver{
		hostConfig:vip,
		lock:new(sync.RWMutex),
	}
	dnsr.resolveHosts()
	return dnsr,nil
}

func (dnsr *DNSResolver)resolveHosts(){
	dnsr.lock.Lock()
	defer dnsr.lock.Unlock()
	dnsr.DNSItems = make(map[string]string)
	hosts := dnsr.hostConfig.GetStringSlice("hosts")
	for _,host :=range hosts{
		item := strings.Split(host," ")
		dnsr.DNSItems[item[0]]=item[1]
	}
}

func (dnsr *DNSResolver)ListHosts()[]string{
	dnsr.lock.RLock()
	defer dnsr.lock.RUnlock()
	list:=make([]string,0)
	for hostname,ipaddr := range dnsr.DNSItems{
		list = append(list,fmt.Sprintf("host %s\t => \t %s\n",hostname,ipaddr))
	}
	return list
}

func (dnsr *DNSResolver)listHostnames()[]string{
	dnsr.lock.RLock()
	defer dnsr.lock.RUnlock()
	list:=make([]string,0)
	for hostname,_:= range dnsr.DNSItems{
		list = append(list,hostname)
	}
	return list
}
func (dnsr *DNSResolver)GetDNS(hostname string) (string,error){
	dnsr.lock.RLock()
	defer dnsr.lock.RUnlock()
	if dnsitem,ok := dnsr.DNSItems[hostname]; ok{
		return dnsitem,nil
	}
	return "",errHostnameNotFoud
}

func(dnsr *DNSResolver)AddItem(hostname string,addr string)error{
	dnsr.lock.Lock()
	defer dnsr.lock.Unlock()
	if _,ok := dnsr.DNSItems[hostname]; !ok{
		dnsr.DNSItems[hostname] = addr
		fmt.Printf("add a new dns item: %s => %s \n",hostname,addr)
		return nil
	}
	return errHostnameNotFoud
}


func(dnsr *DNSResolver)DelItem(hostname string)error{
	dnsr.lock.Lock()
	defer dnsr.lock.Unlock()
	if addr,ok := dnsr.DNSItems[hostname]; ok{
		delete(dnsr.DNSItems,hostname)
		fmt.Printf("delete a dns item: %s => %s \n",hostname,addr)
		return nil
	}
	return errHostnameNotFoud
}



func (dnsr *DNSResolver)Persisit() error{
	dnsr.lock.Lock()
	defer dnsr.lock.Unlock()
	var dns []string
	for key,value :=range dnsr.DNSItems{
		item := key + " " + value
		dns  = append(dns,item)
	}
	dnsr.hostConfig.Set("hosts",dns)
	err := dnsr.hostConfig.WriteConfig()
	if err != nil{
		logger.Critical("persist hosts file failed.")
		return  errors.New(fmt.Sprintf("something wrong when persist the hosts file. %v",err))
	}
	logger.Info("persist hosts file successed")
	return nil
}

