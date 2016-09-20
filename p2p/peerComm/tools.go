// author: chenquan
// date: 16-8-25
// last modified: 16-8-29 13:58
// last Modified Author: chenquan
// change log: add a comment of this file function
//
package peerComm

import (
	"net"
	"io/ioutil"
	"encoding/json"
	"github.com/op/go-logging"
	"strconv"
	pb "hyperchain/p2p/peermessage"
	"encoding/hex"
	"hyperchain/crypto"
)
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p/peerComm")
}

// GetIpLocalIpAddr this function is used to get the real internal net ip address
// to use this make sure your net are valid
func GetLocalIp()string{
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback then display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
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

func ExtactAddress(peerIp string, peerPort_s string) *pb.PeerAddress{
	peerPort_i,err := strconv.Atoi(peerPort_s)
	if err != nil{
		log.Error("port cannot convert into int",err)
	}
	peerPort_i32 := int32(peerPort_i)
	peerAddrString := peerIp + ":" + peerPort_s
	peerAddress := pb.PeerAddress{
		Ip:peerIp,
		Port:peerPort_i32,
		Address:peerAddrString,
		Hash:GetHash(peerAddrString),
	}
	return &peerAddress
}

func GetHash(needHashString string)string {
	hasher := crypto.NewKeccak256Hash("keccak256Hanser")
	return hex.EncodeToString(hasher.Hash(needHashString).Bytes())
}