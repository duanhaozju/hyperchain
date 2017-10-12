//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIntArrayEquals(t *testing.T) {
	array1 := []int{1, 2, 3}
	array2 := []int{1, 2, 3}
	array3 := []int{3, 2, 1}
	array4 := []int{1, 2, 3, 3}
	assert.Equal(t, true, IntArrayEquals(array1, array2), "Array1 should be equal to array2")
	assert.NotEqual(t, true, IntArrayEquals(array1, array3), "Array1 should not be equal to array3")
	assert.NotEqual(t, true, IntArrayEquals(array1, array4), "Array1 should not be equal to array4")
}

func TestClone(t *testing.T) {
	slice := []byte{'a', 'b', 'c', 'd'}
	slice2 := Clone(slice)
	assert.Equal(t, slice, slice2, "The slice should be equal to the cloned slice")
}

func TestTransportEncodeAndDecode(t *testing.T) {
	str := "Hello World!"
	encodeStr := TransportEncode(str)
	decodeStr := TransportDecode(encodeStr)
	assert.Equal(t, str, decodeStr, "The origin string should be equal to the decoded string")
}
