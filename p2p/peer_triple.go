package p2p

import (
	"github.com/pkg/errors"
	"github.com/terasum/viper"
	"sync"
)

var (
	P_STAT_CONNECTED = "CONNECTED"
	P_STAT_PENDDINGD = "PENDDINGD"
	P_STAT_CLOSED    = "CLOSED"
)

type PeerTriple struct {
	namespace string
	id        int
	hostname  string
	status    string
	rwlock    *sync.RWMutex
}

// NewPeerTriple creates and returns a new PeerTriple instance.
func NewPeerTriple(namespace string, id int, hostname string) *PeerTriple {
	return &PeerTriple{
		namespace: namespace,
		id:        id,
		hostname:  hostname,
		status:    P_STAT_PENDDINGD,
		rwlock:    new(sync.RWMutex),
	}
}

// SetOn sets the peer status to CONNECTED.
func (pt *PeerTriple) SetOn() {
	pt.rwlock.Lock()
	defer pt.rwlock.Unlock()
	pt.status = P_STAT_CONNECTED
}

// SetOff sets the peer status to CLOSED.
func (pt *PeerTriple) SetOff() {
	pt.rwlock.Lock()
	defer pt.rwlock.Unlock()
	pt.status = P_STAT_CLOSED
}

// GetStat returns the peer current status.
func (pt *PeerTriple) GetStat() string {
	pt.rwlock.RLock()
	defer pt.rwlock.RUnlock()
	return pt.status
}

// PeerTriples implements Len(), Less(), Swap() three methods for calling the standard library sort.Sort.
type PeerTriples struct {
	rwlock  *sync.RWMutex
	triples []*PeerTriple
}

// NewPeerTriples creates and returns a new PeerTriples instance.
func NewPeerTriples() *PeerTriples {
	return &PeerTriples{
		rwlock:  new(sync.RWMutex),
		triples: make([]*PeerTriple, 0),
	}
}

// QuickParsePeerTriples transfer type []interface{} to PeerTriples.
func QuickParsePeerTriples(namespcace string, nodes []interface{}) (*PeerTriples, error) {
	pts := NewPeerTriples()
	for _, item := range nodes {
		var id int
		var hostname string
		node, ok := item.(map[string]interface{})
		if !ok {
			return nil, errors.New("cannot parse the peer config nodes list, please check the peerconfig.toml file")
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
	return pts, nil
}

// PersistPeerTriples will write PeerTriples information to file.
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

func (pts *PeerTriples) Has(id int, namespace, hostname string) bool {
	pts.rwlock.RLock()
	defer pts.rwlock.RUnlock()
	for _, pt := range pts.triples {
		if pt.id == id && pt.hostname == hostname && pt.namespace == namespace {
			return true
		}
	}
	return false
}

func (pts *PeerTriples) Push(pt *PeerTriple) {
	pts.rwlock.Lock()
	defer pts.rwlock.Unlock()
	pts.triples = append(pts.triples, pt)
}

func (pts *PeerTriples) Remove(id int) *PeerTriple {
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

func (pts *PeerTriples) Pop() *PeerTriple {
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
	ret := pts.triples[len(pts.triples)-1]
	pts.triples = append(make([]*PeerTriple, 0), pts.triples[0:len(pts.triples)-1]...)
	return ret
}

func (pts *PeerTriples) Len() int {
	pts.rwlock.RLock()
	defer pts.rwlock.RUnlock()
	return len(pts.triples)
}

func (pts *PeerTriples) Swap(i, j int) {
	pts.rwlock.Lock()
	defer pts.rwlock.Unlock()
	pts.triples[i], pts.triples[j] = pts.triples[j], pts.triples[i]
}
func (pts *PeerTriples) Less(i, j int) bool {
	pts.rwlock.RLock()
	defer pts.rwlock.RUnlock()
	return pts.triples[i].id > pts.triples[j].id
}

func (pts *PeerTriples) HasNext() bool {
	pts.rwlock.RLock()
	defer pts.rwlock.RUnlock()
	return len(pts.triples) > 0
}
