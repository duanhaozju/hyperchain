package network

import (
	"github.com/spf13/viper"
	"github.com/pkg/errors"
	"strings"
	"fmt"
)
var (
	errConfigNotFound  = errors.New("cannot read in the host config.")
	errHostnameNotFoud = errors.New("hostname cannot resolved.")
)

type DNSResolver struct {
	hostConfig *viper.Viper
	DNSItems map[string]string
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
	}
	dnsr.resolveHosts()
	return dnsr,nil
}

func (dnsr *DNSResolver)resolveHosts(){
	dnsr.DNSItems = make(map[string]string)
	hosts := dnsr.hostConfig.GetStringSlice("hosts")
	for _,host :=range hosts{
		item := strings.Split(host," ")
		dnsr.DNSItems[item[0]]=item[1]
	}
}

func (dnsr *DNSResolver)ListHosts()[]string{
	list:=make([]string,0)
	for hostname,ipaddr := range dnsr.DNSItems{
		list = append(list,fmt.Sprintf("host %s\t => \t %s\n",hostname,ipaddr))
	}
	return list
}

func (dnsr *DNSResolver)listHostnames()[]string{
	list:=make([]string,0)
	for hostname,_:= range dnsr.DNSItems{
		list = append(list,hostname)
	}
	return list
}
func (dnsr *DNSResolver)GetDNS(hostname string) (string,error){
	if dnsitem,ok := dnsr.DNSItems[hostname]; ok{
		return dnsitem,nil
	}
	return "",errHostnameNotFoud
}

func(dnsr *DNSResolver)AddItem(hostname string,addr string)error{
	if _,ok := dnsr.DNSItems[hostname]; !ok{
		dnsr.DNSItems[hostname] = addr
		fmt.Printf("add a new dns item: %s => %s \n",hostname,addr)
		return nil
	}
	return errHostnameNotFoud
}


func(dnsr *DNSResolver)DelItem(hostname string,addr string)error{
	if _,ok := dnsr.DNSItems[hostname]; ok{
		delete(dnsr.DNSItems,hostname)
		fmt.Printf("delete a dns item: %s => %s \n",hostname,addr)
	}
	return errHostnameNotFoud
}

