package p2p

import (
	"sync"
	"container/list"
	"github.com/golang/go/test/fixedbugs"
)

var(
	P_STAT_CONNECTED = "CONNECTED"
	P_STAT_PENDDINGD = "CONNECTED"
	P_STAT_CLOSED = "CONNECTED"
)


type PeerTriple struct {
	namespace string
	id int
	hostname string
	status string
	rwlock *sync.RWMutex
}

func NewPeerTriple(namespace string, id int ,hostname string)*PeerTriple{
	return &PeerTriple{
		namespace:namespace,
		id:id,
		hostname:hostname,
		status:P_STAT_PENDDINGD,
		rwlock:new(sync.RWMutex),
	}
}

//set the peer tripe status off
func(pt *PeerTriple)SetOn(){
	pt.rwlock.Lock()
	defer pt.rwlock.Unlock()
	pt.status = P_STAT_CONNECTED
}

//set the peer tripe status off
func(pt *PeerTriple)SetOff(){
	pt.rwlock.Lock()
	defer pt.rwlock.Unlock()
	pt.status = P_STAT_CLOSED
}

//get current peer tripe's stat
func(pt *PeerTriple)GetStat()string{
	pt.rwlock.RLock()
	defer pt.rwlock.RUnlock()
	return pt.status
}

type PeerTriples struct {
	rwlock *sync.RWMutex
	triples []*PeerTriple
}

func QuickParsePeerTriples(namespcace string,nodes []interface{}) *PeerTriples{
	pts := NewPeerTriples()
	for _,item := range nodes{
		var id int
		var hostname string
		node  := item.(map[interface{}]interface{})
		for key,value := range node{
			if key.(string) == "id"{
				id  = value.(int)
			}
			if key.(string) == "hostname"{
				hostname = value.(string)
			}
		}
		pt := NewPeerTriple(namespcace,id,hostname)
		pts.Push(pt)
	}
	return pts
}

func NewPeerTriples() *PeerTriples{
	return &PeerTriples{
		rwlock:new(sync.RWMutex),
		triples:make([]*PeerTriple,0),
	}
}

func(pts *PeerTriples)Push(pt *PeerTriple){
	pts.rwlock.Lock()
	defer pts.rwlock.Lock()
	pts.triples = append(pts.triples,pt)
}

func(pts *PeerTriples)Pop()*PeerTriple{
	pts.rwlock.Lock()
	defer pts.rwlock.Lock()
	if len(pts.triples)<1{
		return nil
	}
	if len(pts.triples) == 1{
		ret := pts.triples[0]
		pts.triples = make([]*PeerTriple,0)
		return ret
	}
	ret := pts.triples[len(pts)-1]
	pts.triples = append(make([]*PeerTriple,0),pts.triples[0:len(pts)-2]...)
	return ret
}

func (pts *PeerTriples)HasNext() bool{
	pts.rwlock.RLock()
	defer pts.rwlock.RUnlock()
	return len(pts.triples) > 0
}
