package common

type SortableUint64Slice []uint64

func (a SortableUint64Slice) Len() int {
	return len(a)
}
func (a SortableUint64Slice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a SortableUint64Slice) Less(i, j int) bool {
	return a[i] < a[j]
}
