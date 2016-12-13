package hyperstate
import (
	"testing"
	"hyperchain/core/hyperstate/testutils"
)

func TestStateDeltaMarshalling(t *testing.T) {
	stateDelta := NewStateDelta()
	stateDelta.Set("account1", "key1", []byte("value1"), nil)
	stateDelta.Set("account2", "key2", []byte("value2"), nil)
	stateDelta.Delete("account3", "key3", nil)

	by := stateDelta.Marshal()
	t.Logf("length of marshalled bytes = [%d]", len(by))
	stateDelta1 := NewStateDelta()
	stateDelta1.Unmarshal(by)

	testutils.AssertEquals(t, stateDelta1, stateDelta)
}

func TestStateDeltaCryptoHash(t *testing.T) {
	stateDelta := NewStateDelta()

	testutils.AssertNil(t, stateDelta.ComputeCryptoHash())

	stateDelta.Set("accountID1", "key2", []byte("value2"), nil)
	stateDelta.Set("accountID1", "key1", []byte("value1"), nil)
	stateDelta.Set("accountID2", "key2", []byte("value2"), nil)
	stateDelta.Set("accountID2", "key1", []byte("value1"), nil)
	testutils.AssertEquals(t, stateDelta.ComputeCryptoHash(), testutils.ComputeCryptoHash([]byte("accountID1key1value1key2value2accountID2key1value1key2value2")))

	stateDelta.Delete("accountID2", "key1", nil)
	testutils.AssertEquals(t, stateDelta.ComputeCryptoHash(), testutils.ComputeCryptoHash([]byte("accountID1key1value1key2value2accountID2key1key2value2")))
}

func TestStateDeltaEmptyArrayValue(t *testing.T) {
	stateDelta := NewStateDelta()
	stateDelta.Set("account1", "key1", []byte("value1"), nil)
	stateDelta.Set("account2", "key2", []byte{}, nil)
	stateDelta.Set("account3", "key3", nil, nil)
	stateDelta.Set("account4", "", []byte("value4"), nil)

	by := stateDelta.Marshal()
	t.Logf("length of marshalled bytes = [%d]", len(by))
	stateDelta1 := NewStateDelta()
	stateDelta1.Unmarshal(by)

	v := stateDelta1.Get("account2", "key2")
	if v.GetValue() == nil || len(v.GetValue()) > 0 {
		t.Fatalf("An empty array expected. found = %#v", v)
	}

	v = stateDelta1.Get("account3", "key3")
	if v.GetValue() != nil {
		t.Fatalf("Nil value expected. found = %#v", v)
	}

	v = stateDelta1.Get("account4", "")
	testutils.AssertEquals(t, v.GetValue(), []byte("value4"))
}

func TestApplyMerge(t *testing.T) {
	stateDelta := NewStateDelta()
	stateDelta.Set("account1", "key1", []byte("value1"), nil)
	stateDelta.Set("account2", "key2", []byte("value2"), nil)
	anotherStateDelta := NewStateDelta()
	anotherStateDelta.Set("account1", "key1", []byte("value3"), []byte("value1"))
	anotherStateDelta.Set("account2", "key2", []byte("value4"), []byte("value2"))
	anotherStateDelta.Set("account3", "key3", []byte("value6"), []byte("value5"))
	stateDelta.ApplyChanges(anotherStateDelta)
	testutils.AssertEquals(t, stateDelta.ComputeCryptoHash(), testutils.ComputeCryptoHash([]byte("account1key1value3account2key2value4account3key3value6")))
	testutils.AssertEquals(t, stateDelta.Get("account1", "key1").PreviousValue, nil)
	testutils.AssertEquals(t, stateDelta.Get("account2", "key2").PreviousValue, nil)
	testutils.AssertEquals(t, stateDelta.Get("account3", "key3").PreviousValue, []byte("value5"))
}
