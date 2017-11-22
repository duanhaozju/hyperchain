package common

import "testing"

func TestSortableUint64SliceFunctions(t *testing.T) {
	slice := SortableUint64Slice{1, 2, 3, 4, 5}
	if slice.Len() != 5 {
		t.Error("error slice.len != 5")
	}
	if slice.Less(2, 3) != true {
		t.Error("error slice[2] >= slice[3]")
	}
	if slice.Swap(2, 3); !(slice[2] == 4 && slice[3] == 3) {
		t.Error("error exchange slice[2], slice[3]")
	}
}
