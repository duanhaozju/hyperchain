// author: Lizhong kuang
// date: 2016-09-29

package utils

// Clone clones the passed slice
func Clone(src []byte) []byte {
	clone := make([]byte, len(src))
	copy(clone, src)

	return clone
}
