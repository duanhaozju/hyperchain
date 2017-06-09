package common

func RetrieveSnapshotFileds() []string {
	return []string{
		// world state related
		"-account",
		"-storage",
		"-code",
		// bucket tree related
		"-bucket",
		"BucketNode",
	}
}
