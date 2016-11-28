//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package utils

// Clone clones the passed slice
func Clone(src []byte) []byte {
	clone := make([]byte, len(src))
	copy(clone, src)

	return clone
}
