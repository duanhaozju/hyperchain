package info

import (
	"encoding/json"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/p2p/utils"
	"sync"
)

var logger *logging.Logger

// Info represents the basic information of a node.
type Info struct {
	rwmutex     *sync.RWMutex
	Id          int `json:"id"`
	isPrimary   bool
	Hostname    string `json:"hostname"`
	Namespace   string `json:"namespace"`
	Hash        string `json:"hash"`
	IsVP        bool   `json:"isvp"`
	isOriginal  bool   `json"isorg`
	isReconnect bool   `json"isorg`
}

// NewInfo creates and returns a new Info instance.
func NewInfo(id int, hostname string, namespcace string) *Info {
	logger = common.GetLogger(common.DEFAULT_LOG, "p2p")
	hash := utils.Sha3([]byte(hostname + namespcace))
	return &Info{
		rwmutex:    new(sync.RWMutex),
		Id:         id,
		isPrimary:  false,
		Hostname:   hostname,
		Hash:       common.Bytes2Hex(hash),
		Namespace:  namespcace,
		IsVP:       true,
		isOriginal: false,
	}
}

// GetHash returns the node hash.
func (i *Info) GetHash() string {
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.Hash
}

func (i *Info) SetOrg() {
	i.rwmutex.Lock()
	defer i.rwmutex.Unlock()
	i.isOriginal = true
}

func (i *Info) IsOrg() bool {
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.isOriginal
}

func (i *Info) SetRec() {
	i.rwmutex.Lock()
	defer i.rwmutex.Unlock()
	i.isReconnect = true
}

func (i *Info) IsRec() bool {
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.isReconnect
}
func (i *Info) SetHostName(hostname string) {
	i.rwmutex.Lock()
	defer i.rwmutex.Unlock()
	i.Hostname = hostname
}

func (i *Info) GetHostName() string {
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.Hostname
}

func (i *Info) SetID(id int) {
	i.rwmutex.Lock()
	defer i.rwmutex.Unlock()
	i.Id = id
}

//get nodeID
func (i *Info) GetID() int {
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.Id
}

func (i *Info) SetPrimary(flag bool) {
	i.rwmutex.Lock()
	defer i.rwmutex.Unlock()
	i.isPrimary = flag
}

func (i *Info) GetPrimary() bool {
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.isPrimary
}

func (i *Info) GetVP() bool {
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.IsVP
}

func (i *Info) Serialize() []byte {
	b, e := json.Marshal(i)
	if e != nil {
		logger.Errorf("serialize info err:%s \n", e.Error())
		return nil
	}
	return b
}

func InfoUnmarshal(raw []byte) *Info {
	i := new(Info)
	err := json.Unmarshal(raw, i)
	if err != nil {
		logger.Errorf("cannnot unmarshal info %s", err.Error())
		return nil
	}
	return i
}

func (i *Info) GetNameSpace() string {
	return i.Namespace
}

func (i *Info) SetNVP() {
	i.IsVP = false
}
