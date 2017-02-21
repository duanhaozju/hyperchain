package peermessage

// Len is the number of elements in the collection.
func (routers Routers) Len() int {
	return len(routers.Routers)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (routers Routers) Less(i, j int) bool {
	id1 := routers.Routers[i].ID
	id2 := routers.Routers[j].ID
	return id1 < id2

}

// Swap swaps the elements with indexes i and j.
func (routers Routers) Swap(i, j int) {
	temp := routers.Routers[i]
	routers.Routers[i] = routers.Routers[j]
	routers.Routers[j] = temp
}
