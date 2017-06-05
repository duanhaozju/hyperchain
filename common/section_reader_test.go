package common

import (
	"testing"
	"time"
	"math/rand"
	"io/ioutil"
	"os"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestSectionReader_ReatAt(t *testing.T) {
	var filesize int64 = 10005
	var shardLen int64 = 100
	buf := RandStringRunes(int(filesize))
	ioutil.WriteFile("tmp", buf, 0644)
	defer os.Remove("tmp")

	err, reader := NewSectionReader("tmp", shardLen)
	if err != nil {
		t.Error(err.Error())
	}
	if _, _, err := reader.ReatAt(0); err == nil {
		t.Error("read at 0 shard, expect failure")
	}

	if _, _, err := reader.ReatAt(1); err != nil {
		t.Error("read at 1 shard, expect no failure")
	}
	if len, _, err := reader.ReatAt(101); len != 5 || err == nil || err.Error() != "EOF" {
		t.Error("read at the last shard, expect EOF error")
	}
}

func TestSectionReader_ReadNext(t *testing.T) {
	var filesize int64 = 10005
	var shardLen int64 = 100
	buf := RandStringRunes(int(filesize))
	ioutil.WriteFile("tmp", buf, 0644)
	defer os.Remove("tmp")

	err, reader := NewSectionReader("tmp", shardLen)
	if err != nil {
		t.Error(err.Error())
	}
	iterCnt := filesize / shardLen
	if filesize % shardLen > 0 {
		iterCnt += 1
	}
	t.Log("total iter count", iterCnt)
	var i int64 = 0
	for ; i < iterCnt; i += 1 {
		len, _, err := reader.ReadNext()
		t.Logf("(iter %d) content len(#%d)", i, len)
		if err != nil && len == 0 {
			t.Error(err.Error())
		}
		if i == iterCnt - 1 {
		} else {
			if int64(len) != shardLen {
				t.Error("shard len not satisfy")
			}
		}
	}
	reader.Close()
}

var letterBytes = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}
