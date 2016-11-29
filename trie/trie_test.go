//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package trie

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	checker "gopkg.in/check.v1"
	"hyperchain/common"
	"hyperchain/hyperdb"
	"io/ioutil"
	"os"
	"testing"
	crand "crypto/rand"
	"math/rand"
	"reflect"
	"testing/quick"
)

type TrieSuite struct {
}

func Test(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&TrieSuite{})

func init() {
	spew.Config.Indent = "    "
	spew.Config.DisableMethods = true
}

// Used for testing
func newEmpty() *Trie {
	db, _ := hyperdb.NewMemDatabase()
	trie, _ := New(common.Hash{}, db)
	return trie
}

func TestEmptyTrie(t *testing.T) {
	var trie Trie
	res := trie.Hash()
	exp := emptyRoot
	if res != common.Hash(exp) {
		t.Errorf("expected %x got %x", exp, res)
	}
}

func TestNull(t *testing.T) {
	var trie Trie
	key := make([]byte, 32)
	value := common.FromHex("0x823140710bf13990e4500136726d8b55")
	trie.Update(key, value)
	value_1 := trie.Get(key)
	if bytes.Compare(value, value_1) != 0 {
		t.Errorf("Mismatch value")
	}
}

func TestMissingRoot(t *testing.T) {
	db, _ := hyperdb.NewMemDatabase()
	trie, err := New(common.HexToHash("0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33"), db)
	if trie != nil {
		t.Error("New returned non-nil trie for invalid root")
	}
	if _, ok := err.(*MissingNodeError); !ok {
		t.Errorf("New returned wrong error: %v", err)
	}
}

func TestMissingNode(t *testing.T) {
	var res []byte
	db, _ := hyperdb.NewMemDatabase()
	trie, _ := New(common.Hash{}, db)
	updateString(trie, "120000", "qwerqwerqwerqwerqwerqwerqwerqwer")
	updateString(trie, "123456", "asdfasdfasdfasdfasdfasdfasdfasdf")
	root, _ := trie.Commit()
	ClearGlobalCache()

	trie, _ = New(root, db)
	if trie == nil {
		t.Errorf("New trie failed")
	}
	res, err := trie.TryGet([]byte("120000"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		t.Logf("node 120000: value: %v", string(res))
	}

	trie, _ = New(root, db)
	res, err = trie.TryGet([]byte("120099"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		t.Logf("node 120099: value: %v", string(res))
	}

	trie, _ = New(root, db)
	res, err = trie.TryGet([]byte("123456"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		t.Logf("node 123456: value: %v", string(res))
	}
	trie, _ = New(root, db)
	err = trie.TryUpdate([]byte("120099"), []byte("zxcvzxcvzxcvzxcvzxcvzxcvzxcvzxcv"))

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	res, err = trie.TryGet([]byte("120099"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		t.Logf("node 120099: value: %v", string(res))
	}

	trie, _ = New(root, db)
	err = trie.TryDelete([]byte("123456"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	res, err = trie.TryGet([]byte("123456"))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else {
		t.Logf("node 123456: value: %v", string(res))
	}
	//listMem(db)
	db.Delete(common.FromHex("bf6903a2d26288bf89eac109d2d8e5de4764084fa45f923d9e6d644a7f1fe5c7"))
	ClearGlobalCache()

	trie, _ = New(root, db)
	_, err = trie.TryGet([]byte("123456"))
	if _, ok := err.(*MissingNodeError); !ok {
		t.Errorf("Wrong error: %v", err)
	}
}



func TestGet(t *testing.T) {
	trie := newEmpty()
	updateString(trie, "doe", "reindeer")
	updateString(trie, "dog", "puppy")
	updateString(trie, "dogglesworth", "cat")

	for i := 0; i < 2; i++ {
		res := getString(trie, "dog")
		if !bytes.Equal(res, []byte("puppy")) {
			t.Errorf("expected puppy got %x", res)
		}

		unknown := getString(trie, "unknown")
		if unknown != nil {
			t.Errorf("expected nil got %x", unknown)
		}

		if i == 1 {
			return
		}
		trie.Commit()
	}
}


func TestReplication(t *testing.T) {
	trie := newEmpty()
	vals := []struct{ k, v string }{
		{"do", "verb"},
		{"ether", "wookiedoo"},
		{"horse", "stallion"},
		{"shaman", "horse"},
		{"doge", "coin"},
		{"dog", "puppy"},
		{"somethingveryoddindeedthis is", "myothernodedata"},
	}
	for _, val := range vals {
		updateString(trie, val.k, val.v)
	}
	exp, err := trie.Commit()
	if err != nil {
		t.Fatalf("commit error: %v", err)
	}

	// create a new trie on top of the database and check that lookups work.
	trie2, err := New(exp, trie.db)
	if err != nil {
		t.Fatalf("can't recreate trie at %x: %v", exp, err)
	}
	for _, kv := range vals {
		if string(getString(trie2, kv.k)) != kv.v {
			t.Errorf("trie2 doesn't have %q => %q", kv.k, kv.v)
		}
	}
	hash, err := trie2.Commit()
	if err != nil {
		t.Fatalf("commit error: %v", err)
	}
	if hash != exp {
		t.Errorf("root failure. expected %x got %x", exp, hash)
	}

	// perform some insertions on the new trie.
	vals2 := []struct{ k, v string }{
		{"do", "verb"},
		{"ether", "wookiedoo"},
		{"horse", "stallion"},
		// {"shaman", "horse"},
		// {"doge", "coin"},
		// {"ether", ""},
		// {"dog", "puppy"},
		// {"somethingveryoddindeedthis is", "myothernodedata"},
		// {"shaman", ""},
	}
	for _, val := range vals2 {
		updateString(trie2, val.k, val.v)
	}
	if hash := trie2.Hash(); hash != exp {
		t.Errorf("root failure. expected %x got %x", exp, hash)
	}
}

func paranoiaCheck(t1 *Trie) (bool, *Trie) {
	t2 := new(Trie)
	it := NewIterator(t1)
	for it.Next() {
		t2.Update(it.Key, it.Value)
	}
	return t2.Hash() == t1.Hash(), t2
}

func TestParanoia(t *testing.T) {
	trie := newEmpty()

	vals := []struct{ k, v string }{
		{"do", "verb"},
		{"ether", "wookiedoo"},
		{"horse", "stallion"},
		{"shaman", "horse"},
		{"doge", "coin"},
		{"ether", ""},
		{"dog", "puppy"},
		{"shaman", ""},
		{"somethingveryoddindeedthis is", "myothernodedata"},
	}
	for _, val := range vals {
		updateString(trie, val.k, val.v)
	}
	trie.Commit()

	ok, t2 := paranoiaCheck(trie)
	if !ok {
		t.Errorf("trie paranoia check failed %x %x", trie.Hash(), t2.Hash())
	}
}

func (s *TrieSuite) TestDeleteAfterLoad(c *checker.C) {
	trie := newEmpty()

	vals := []struct{ k, v string }{
		{"do", "verb"},
		{"ether", "wookiedoo"},
		{"horse", "stallion"},
		{"shaman", "horse"},
		{"doge", "coin"},
		{"dog", "puppy"},
		{"somethingveryoddindeedthis is", "myothernodedata"},
	}
	for _, val := range vals {
		updateString(trie, val.k, val.v)
	}
	root, _ := trie.Commit()
	ClearGlobalCache()

	newTrie, _ := New(root, trie.db)

	for _, val := range vals {
		ret := string(newTrie.Get([]byte(val.k)))
		c.Assert(val.v, checker.Equals, ret)
	}
	var begin = 2
	var end = 4
	removeVals := vals[begin:end]
	var remainVals []struct{ k, v string }
	remainVals = append(remainVals, vals[:begin]...)
	remainVals = append(remainVals, vals[end:]...)
	for _, val := range removeVals {
		deleteString(newTrie, val.k)
	}
	for _, val := range remainVals {
		ret := string(newTrie.Get([]byte(val.k)))
		c.Assert(val.v, checker.Equals, ret)
	}

	newInstvals := []struct{ k, v string }{
		{"key1", "val1"},
		{"key2", "val2"},
		{"key3", "val3"},
		{"w1", "q1"},
		{"w2", "q2"},
		{"w3", "q3"},
	}
	for _, val := range newInstvals {
		updateString(newTrie, val.k, val.v)
	}

	remainVals = append(remainVals, newInstvals[:]...)
	for _, val := range remainVals {
		ret := string(newTrie.Get([]byte(val.k)))
		c.Assert(val.v, checker.Equals, ret)
	}

}


func TestLargeValue(t *testing.T) {
	trie := newEmpty()
	trie.Update([]byte("key1"), []byte{99, 99, 99, 99})
	trie.Update([]byte("key2"), bytes.Repeat([]byte{1}, 32))
	trie.Hash()
}

type kv struct {
	k, v []byte
	t    bool
}

func TestLargeData(t *testing.T) {
	trie := newEmpty()
	vals := make(map[string]*kv)

	for i := byte(0); i < 255; i++ {
		value := &kv{common.LeftPadBytes([]byte{i}, 32), []byte{i}, false}
		value2 := &kv{common.LeftPadBytes([]byte{10, i}, 32), []byte{i}, false}
		trie.Update(value.k, value.v)
		trie.Update(value2.k, value2.v)
		vals[string(value.k)] = value
		vals[string(value2.k)] = value2
	}

	it := NewIterator(trie)
	for it.Next() {
		vals[string(it.Key)].t = true
	}

	var untouched []*kv
	for _, value := range vals {
		if !value.t {
			untouched = append(untouched, value)
		}
	}

	if len(untouched) > 0 {
		t.Errorf("Missed %d nodes", len(untouched))
		for _, value := range untouched {
			t.Error(value)
		}
	}
}

func BenchmarkGet(b *testing.B)      { benchGet(b, false) }
func BenchmarkGetDB(b *testing.B)    { benchGet(b, true) }
func BenchmarkUpdateBE(b *testing.B) { benchUpdate(b, binary.BigEndian) }
func BenchmarkUpdateLE(b *testing.B) { benchUpdate(b, binary.LittleEndian) }
func BenchmarkHashBE(b *testing.B)   { benchHash(b, binary.BigEndian) }
func BenchmarkHashLE(b *testing.B)   { benchHash(b, binary.LittleEndian) }

const benchElemCount = 20000

func benchGet(b *testing.B, commit bool) {
	trie := new(Trie)
	if commit {
		dir, tmpdb := tempDB()
		defer os.RemoveAll(dir)
		trie, _ = New(common.Hash{}, tmpdb)
	}
	k := make([]byte, 32)
	for i := 0; i < benchElemCount; i++ {
		binary.LittleEndian.PutUint64(k, uint64(i))
		trie.Update(k, k)
	}
	binary.LittleEndian.PutUint64(k, benchElemCount/2)
	if commit {
		trie.Commit()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Get(k)
	}
}

func benchUpdate(b *testing.B, e binary.ByteOrder) *Trie {
	trie := newEmpty()
	k := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		e.PutUint64(k, uint64(i))
		trie.Update(k, k)
	}
	return trie
}

func benchHash(b *testing.B, e binary.ByteOrder) {
	trie := newEmpty()
	k := make([]byte, 32)
	for i := 0; i < benchElemCount; i++ {
		e.PutUint64(k, uint64(i))
		trie.Update(k, k)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Hash()
	}
}

func tempDB() (string, Database) {
	dir, err := ioutil.TempDir("", "trie-bench")
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory: %v", err))
	}
	db, err := hyperdb.NewLDBDataBase(dir)
	if err != nil {
		panic(fmt.Sprintf("can't create temporary database: %v", err))
	}
	return dir, db
}
func getString(trie *Trie, k string) []byte {
	return trie.Get([]byte(k))
}
func updateString(trie *Trie, k, v string) {
	trie.Update([]byte(k), []byte(v))
}

func deleteString(trie *Trie, k string) {
	trie.Delete([]byte(k))
}
func listMem(db *hyperdb.MemDatabase) {
	for _, key := range db.Keys() {
		val, _ := db.Get(key)
		fmt.Printf("MEM key: %v, val: %v\n", key, val)
	}
}
func traverse_trie(trie *Trie) {
	it := NewIterator(trie)
	for it.Next() {
		fmt.Printf("TRIE key: %v, val : %v\n", string(it.Key), string(it.Value))
	}
}

type randTestStep struct {
	op    int
	key   []byte
	value []byte
}
type randTest []randTestStep

const (
	opUpdate = iota
	opDelete
	opGet
	opCommit
	opHash
	opReset
	opItercheckhash
	opMax
)

func (randTest) Generate(r *rand.Rand, size int) reflect.Value {
	var allKeys [][]byte
	genKey := func() []byte {
		if len(allKeys) < 2 || r.Intn(100) < 10 {
			// new key
			key := make([]byte, 30)
			crand.Read(key)
			allKeys = append(allKeys, key)
			return key
		}
		// use existing key
		return allKeys[r.Intn(len(allKeys))]
	}

	var steps randTest
	for i := 0; i < size; i++ {
		step := randTestStep{op: r.Intn(opMax)}
		switch step.op {
		case opUpdate:
			step.key = genKey()
			step.value = []byte("value")
			//binary.BigEndian.PutUint64(step.value, uint64(i))
		case opGet, opDelete:
			step.key = genKey()
		}
		steps = append(steps, step)
	}
	return reflect.ValueOf(steps)
}

func runRandTest(rt randTest) bool {
	db, _ := hyperdb.NewMemDatabase()
	tr, _ := New(common.Hash{}, db)
	values := make(map[string]string) // tracks content of the trie

	for _, step := range rt {
		switch step.op {
		case opUpdate:
			tr.Update(step.key, step.value)
			values[string(step.key)] = string(step.value)
		case opDelete:
			tr.Delete(step.key)
			delete(values, string(step.key))
		case opGet:
			v := tr.Get(step.key)
			want := values[string(step.key)]
			if string(v) != want {
				fmt.Printf("mismatch for key 0x%x, got 0x%x want 0x%x", step.key, v, want)
				return false
			}
		case opCommit:
			if _, err := tr.Commit(); err != nil {
				panic(err)
			}
		case opHash:
			tr.Hash()
		case opReset:
			hash, err := tr.Commit()
			if err != nil {
				panic(err)
			}
			newtr, err := New(hash, db)
			if err != nil {
				panic(err)
			}
			tr = newtr
		case opItercheckhash:
			checktr, _:= New(common.Hash{}, nil)
			it := tr.Iterator()
			for it.Next() {
				checktr.Update(it.Key, it.Value)
			}
			if tr.Hash() != checktr.Hash() {
				return false
			}
			return true
		}
	}
	return true
}

func TestRandom(t *testing.T) {
	if err := quick.Check(runRandTest, nil); err != nil {
		t.Fatal(err)
	}
}

