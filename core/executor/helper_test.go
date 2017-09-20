package executor

import (
	"hyperchain/common"
	"testing"
)

func TestSplitVpId1(t *testing.T) {
	hash := "76c0e33b83ed0f98280560ae4675d81209a76acfba75aab2660d9ee13abe72c3"
	id := 1
	join := append(common.Hex2Bytes(hash), common.Int2Bytes(id)...)
	res, err := SplitVpId(join)
	if res != uint64(id) || err != nil {
		t.Errorf("expect %v got %v. error:%s.", uint64(id), res, err.Error())
	}
}

func TestSplitVpId2(t *testing.T) {
	hash := "76c0e33b83ed0f98280560ae4675d81209a76acfba75aab2660d9ee13abe72c3"
	id := 29
	join := append(common.Hex2Bytes(hash), common.Int2Bytes(id)...)
	res, err := SplitVpId(join)
	if res != uint64(id) || err != nil {
		t.Errorf("expect %v got %v. error:%s.", uint64(id), res, err.Error())
	}
}

func TestSplitVpId3(t *testing.T) {
	hash := "76c0e33b83ed0f98280560ae4675d81209a76acfba75aab2660d9ee13abe72c3"
	id := 0
	join := append(common.Hex2Bytes(hash), common.Int2Bytes(id)...)
	res, err := SplitVpId(join)
	if res != uint64(id) || err != nil {
		t.Errorf("expect %v got %v. error:%s.", uint64(id), res, err.Error())
	}
}

func TestSplitNVPHash1(t *testing.T) {
	hash := "76c0e33b83ed0f98280560ae4675d81209a76acfba75aab2660d9ee13abe72c3"
	id := 1
	join := append(common.Hex2Bytes(hash), common.Int2Bytes(id)...)
	res, err := SplitNvpHash(join)
	if res != hash || err != nil {
		t.Errorf("expect %v got %v. error:%s.", hash, res, err.Error())
	}
}

func TestSplitNVPHash2(t *testing.T) {
	hash := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	id := 1
	join := append(common.Hex2Bytes(hash), common.Int2Bytes(id)...)
	res, err := SplitNvpHash(join)
	if res != hash || err != nil {
		t.Errorf("expect %v got %v. error:%s.", hash, res, err.Error())
	}
}

func TestSplitNVPHash3(t *testing.T) {
	hash := "0000000000000000000000000000000000000000000000000000000000000000"
	id := 1
	join := append(common.Hex2Bytes(hash), common.Int2Bytes(id)...)
	res, err := SplitNvpHash(join)
	if res != hash || err != nil {
		t.Errorf("expect %v got %v. error:%s.", hash, res, err.Error())
	}
}
