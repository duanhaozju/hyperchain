package p2p

import (
	"sync"
	"github.com/terasum/viper"
	"github.com/pkg/errors"
)

var (
	P_STAT_CONNECTED = "CONNECTED"
	P_STAT_PENDDINGD = "CONNECTED"
	P_STAT_CLOSED = "CONNECTED"
)

type PeerTriple struct {
	namespace string
	id        int
	hostname  string
	status    string
	rwlock    *sync.RWMutex
}

func NewPeerTriple(namespace string, id int, hostname string) *PeerTriple {
	return &PeerTriple{
		namespace:namespace,
		id:id,
		hostname:hostname,
		status:P_STAT_PENDDINGD,
		rwlock:new(sync.RWMutex),
	}
}

//set the peer tripe status off
func (pt *PeerTriple)SetOn() {
	pt.rwlock.Lock()
	defer pt.rwlock.Unlock()
	pt.status = P_STAT_CONNECTED
}

//set the peer tripe status off
func (pt *PeerTriple)SetOff() {
	pt.rwlock.Lock()
	defer pt.rwlock.Unlock()
	pt.status = P_STAT_CLOSED
}

//get current peer tripe's stat
func (pt *PeerTriple)GetStat() string {
	pt.rwlock.RLock()
	defer pt.rwlock.RUnlock()
	return pt.status
}

type PeerTriples struct {
	rwlock  *sync.RWMutex
	triples []*PeerTriple
}

func QuickParsePeerTriples(namespcace string, nodes []interface{}) (*PeerTriples,error) {
	pts := NewPeerTriples()
	for _, item := range nodes {
		var id int
		var hostname string
		node,ok := item.(map[string]interface{})
		if !ok {
			return errors.New("cannot parse the peer config nodes list, please check the peerconfig.yaml file")
		}
		for key, value := range node {
			if key == "id" {
				id = int(value.(int64))
			}
			if key == "hostname" {
				hostname = value.(string)
			}
		}
		pt := NewPeerTriple(namespcace, id, hostname)
		pts.Push(pt)
	}
	return pts,nil
}

func SwitchPeerTriples(pts *PeerTriples, tpts *PeerTriples) {
	pts = tpts
}

func PersistPeerTriples(vip *viper.Viper, pts *PeerTriples) error {
	pts.rwlock.RLock()
	defer pts.rwlock.RUnlock()
	s := make([]interface{}, 0)
	for _, t := range pts.triples {
		tmpm := make(map[string]interface{})
		tmpm["id"] = t.id
		tmpm["hostname"] = t.hostname
		s = append(s, tmpm)
	}
	vip.Set("nodes", s)
	vip.Set("self.n", int64(len(s)))
	return vip.WriteConfig()
}

func NewPeerTriples() *PeerTriples {
	return &PeerTriples{
		rwlock:new(sync.RWMutex),
		triples:make([]*PeerTriple, 0),
	}
}

func (pts *PeerTriples)Has(id int, namespace, hostname string) bool {
	pts.rwlock.RLock()
	defer pts.rwlock.RUnlock()
	for _, pt := range pts.triples {
		if pt.id == id && pt.hostname == hostname && pt.namespace == namespace {
			return true
		}
	}
	return false
}

func (pts *PeerTriples)Push(pt *PeerTriple) {
	pts.rwlock.Lock()
	defer pts.rwlock.Unlock()
	pts.triples = append(pts.triples, pt)
}

func (pts *PeerTriples)Remove(id int) *PeerTriple {
	pts.rwlock.Lock()
	defer pts.rwlock.Unlock()
	if len(pts.triples) < 1 {
		return nil
	}

	tmp_t := make([]*PeerTriple, 0)
	var ret *PeerTriple
	for _, pt := range pts.triples {
		if pt.id == id {
			ret = pt
			continue
		}
		tmp_t = append(tmp_t, pt)
	}
	pts.triples = tmp_t
	return ret
}

func (pts *PeerTriples)Pop() *PeerTriple {
	pts.rwlock.Lock()
	defer pts.rwlock.Unlock()
	if len(pts.triples) < 1 {
		return nil
	}
	if len(pts.triples) == 1 {
		ret := pts.triples[0]
		pts.triples = make([]*PeerTriple, 0)
		return ret
	}
	ret := pts.triples[len(pts.triples) - 1]
	pts.triples = append(make([]*PeerTriple, 0), pts.triples[0:len(pts.triples) - 1]...)
	return ret
}

func (pts *PeerTriples)Len() int {
	pts.rwlock.RLock()
	defer pts.rwlock.RUnlock()
	return len(pts.triples)
}

func (pts *PeerTriples)Swap(i, j int) {
	pts.rwlock.Lock()
	defer pts.rwlock.Unlock()
	pts.triples[i], pts.triples[j] = pts.triples[j], pts.triples[i]
}
func (pts *PeerTriples)Less(i, j int) bool {
	pts.rwlock.RLock()
	defer pts.rwlock.RUnlock()
	return pts.triples[i].id > pts.triples[j].id
}

func (pts *PeerTriples)HasNext() bool {
	pts.rwlock.RLock()
	defer pts.rwlock.RUnlock()
	return len(pts.triples) > 0
}
