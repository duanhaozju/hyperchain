//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package peerComm

import (
	"encoding/hex"
	"github.com/op/go-logging"
	"hyperchain/crypto"
	pb "hyperchain/p2p/peermessage"
	"net"
	"strconv"
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p/peerComm")
}

// GetIpLocalIpAddr this function is used to get the real internal net ip address
// to use this make sure your net are valid
func GetLocalIp() string {
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

func ExtractAddress(peerIp string, peerPort int64, ID uint64) *pb.PeerAddress {
	peerAddrString := strconv.FormatUint(ID, 10) + ":" + strconv.FormatInt(peerPort, 10)
	peerAddress := pb.PeerAddress{
		IP:   peerIp,
		Port: peerPort,
		Hash: GetHash(peerAddrString),
		ID:   ID,
	}
	return &peerAddress
}

func GetHash(needHashString string) string {
	hasher := crypto.NewKeccak256Hash("keccak256Hanser")
	return hex.EncodeToString(hasher.Hash(needHashString).Bytes())
}
