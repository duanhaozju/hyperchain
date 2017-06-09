package common

func RetrieveSnapshotFileds() []string {
	return []string{
		// world state related
		"-storage",
		"-account",
		"-code",
		// bucket tree related
		"-bucket",
		"BucketNode",
	}
}
