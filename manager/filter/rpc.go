package filter

import (
	"sync"
	"encoding/hex"
	"math/rand"
	"strings"
	"time"
)

var (
	idGeneratorLock sync.Mutex
)

func init() {
	rand.Seed(time.Now().UnixNano())
}


// idGenerator helper utility that generates a (pseudo) random sequence of
// bytes that are used to generate identifiers.
func GenerateID() int64 {
	return rand.Int63()
}


func NewFilterID() string {
	idGeneratorLock.Lock()
	defer idGeneratorLock.Unlock()
	id := make([]byte, 16)
	for i := 0; i < len(id); i += 7 {
		val := GenerateID()
		for j := 0; i+j < len(id) && j < 7; j++ {
			id[i+j] = byte(val)
			val >>= 8
		}
	}

	rpcId := hex.EncodeToString(id)
	// rpc ID's are RPC quantities, no leading zero's and 0 is 0x0
	rpcId = strings.TrimLeft(rpcId, "0")
	if rpcId == "" {
		rpcId = "0"
	}
	return  "0x" + rpcId
}

