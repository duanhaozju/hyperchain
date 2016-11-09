//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package util

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestComputeCryptoHash(t *testing.T) {
	if bytes.Compare(ComputeCryptoHash([]byte("foobar")), ComputeCryptoHash([]byte("foobar"))) != 0 {
		t.Fatalf("Expected hashes to match, but they did not match")
	}
	if bytes.Compare(ComputeCryptoHash([]byte("foobar1")), ComputeCryptoHash([]byte("foobar2"))) == 0 {
		t.Fatalf("Expected hashes to be different, but they match")
	}
}

func TestUUIDGeneration(t *testing.T) {
	uuid := GenerateUUID()
	if len(uuid) != 36 {
		t.Fatalf("UUID length is not correct. Expected = 36, Got = %d", len(uuid))
	}
	uuid2 := GenerateUUID()
	if uuid == uuid2 {
		t.Fatalf("Two UUIDs are equal. This should never occur")
	}
}

func TestIntUUIDGeneration(t *testing.T) {
	uuid := GenerateIntUUID()

	uuid2 := GenerateIntUUID()
	if uuid == uuid2 {
		t.Fatalf("Two UUIDs are equal. This should never occur")
	}
}
func TestTimestamp(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Logf("timestamp now: %v", CreateUtcTimestamp())
		time.Sleep(200 * time.Millisecond)
	}
}

func TestGenerateHashFromSignature(t *testing.T) {
	if bytes.Compare(GenerateHashFromSignature("aPath", "aCtor", []string{"1", "2"}),
		GenerateHashFromSignature("aPath", "aCtor", []string{"1", "2"})) != 0 {
		t.Fatalf("Expected hashes to match, but they did not match")
	}
	if bytes.Compare(GenerateHashFromSignature("aPath", "aCtor", []string{"1", "2"}),
		GenerateHashFromSignature("bPath", "bCtor", []string{"3", "4"})) == 0 {
		t.Fatalf("Expected hashes to be different, but they match")
	}
}

func TestFindMissingElements(t *testing.T) {
	all := []string{"a", "b", "c", "d"}
	some := []string{"b", "c"}
	expectedDelta := []string{"a", "d"}
	actualDelta := FindMissingElements(all, some)
	if len(expectedDelta) != len(actualDelta) {
		t.Fatalf("Got %v, expected %v", actualDelta, expectedDelta)
	}
	for i := range expectedDelta {
		if strings.Compare(expectedDelta[i], actualDelta[i]) != 0 {
			t.Fatalf("Got %v, expected %v", actualDelta, expectedDelta)
		}
	}
}
