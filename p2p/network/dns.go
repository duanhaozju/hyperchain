package network

import (
	"github.com/spf13/viper"
	"github.com/pkg/errors"
	"fmt"
	"strings"
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

func (dnsr *DNSResolver)ListHosts(){
	for hostname,ip := range dnsr.DNSItems{
		fmt.Println(hostname,ip)
	}
}

func (dnsr *DNSResolver)GetDNS(hostname string) (string,error){
	if dnsitem,ok := dnsr.DNSItems[hostname]; ok{
		return dnsitem,nil
	}
	return "",errHostnameNotFoud
}