package info

import (
	"sync"
	"hyperchain/crypto/sha3"
	"hyperchain/common"
	"encoding/json"
	"fmt"
)

type Info struct {
	rwmutex   *sync.RWMutex
	Id        int `json:"id"`
	isPrimary bool
	Hostname  string `json:"hostname"`
	Namespace string `json:"namespace"`
	Hash      string `json:"hash"`
}

func NewInfo(id int,hostname string,namespcace string)*Info {
	fmt.Println("new info namespace:",namespcace)
	h := sha3.NewKeccak256()
	h.Write([]byte(namespcace))
	hash := h.Sum([]byte(hostname))
	return &Info{
		rwmutex:new(sync.RWMutex),
		Id:id,
		isPrimary:false,
		Hostname:hostname,
		Hash:common.Bytes2Hex(hash),
		Namespace:namespcace,
	}
}

func (i *Info)GetHash()string{
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.Hash
}


func(i *Info)SetHostName(hostname string){
	i.rwmutex.Lock()
	defer i.rwmutex.Unlock()
	i.Hostname = hostname
}

func(i *Info)GetHostName() string{
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.Hostname
}

//get nodeID
func(i *Info)SetID(id int){
	i.rwmutex.Lock()
	defer i.rwmutex.Unlock()
	i.Id = id
}

//get nodeID
func(i *Info)GetID() int{
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.Id
}

//set node primary info
func(i *Info)SetPrimary(flag bool){
	i.rwmutex.Lock()
	defer i.rwmutex.Unlock()
	i.isPrimary = flag
}
// get the node info primary info
func(i *Info)GetPrimary() bool{
	i.rwmutex.RLock()
	defer i.rwmutex.RUnlock()
	return i.isPrimary
}

func (i *Info)Serialize()[]byte{
	b,e := json.Marshal(i)
	if e != nil{
		fmt.Println(e)
		return nil
	}
	fmt.Println("Info Serialize",string(b))
	return b
}

func InfoUnmarshal(raw []byte)*Info{
	i := new(Info)
	err := json.Unmarshal(raw,i)
	if err != nil{
		fmt.Errorf("cannnot unmarshal info %s",err.Error())
		return nil
	}
	return i
}

func (i *Info)GetNameSpace()string{
	return i.Namespace
}
