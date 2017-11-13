package utils

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto/sha3"
	"net"
	"os"
	"strconv"
	"strings"
)

func GetProjectPath() string {
	gopath := os.Getenv("GOPATH")
	if strings.Contains(gopath, ":") {
		gopath = strings.Split(gopath, ":")[0]
	}
	return gopath + "/src/github.com/hyperchain/hyperchain"
}

func GetPeerHash(namespace string, id int) string {
	hasher := sha3.NewKeccak256()
	ids := strconv.Itoa(id)
	hasher.Write([]byte(namespace + ids))
	return common.ToHex(hasher.Sum(nil))
}

func HashString(in string) string {
	hasher := sha3.NewKeccak256()
	hasher.Write([]byte(in))
	return common.Bytes2Hex(hasher.Sum(nil))
}

func IPcheck(ip string) bool {
	trial := net.ParseIP(ip)
	if trial == nil {
		return false
	}
	if trial.To4() == nil {
		return false
	} else {
		return true
	}
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func Sha3(data []byte) (hash []byte) {
	hasher := sha3.NewKeccak256()
	hasher.Write(data)
	hash = hasher.Sum(nil)
	return
}
