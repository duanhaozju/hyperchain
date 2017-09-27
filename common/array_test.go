package common

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestIntArrayEquals(t *testing.T) {
	array1 := []int{1,2,3}
	array2 := []int{1,2,3}
	array3 := []int{3,2,1}
	array4 := []int{1,2,3,3}
	assert.Equal(t, true, IntArrayEquals(array1,array2), "they should be equal")
	assert.NotEqual(t, true, IntArrayEquals(array1,array3), "they should not be equal")
	assert.NotEqual(t, true, IntArrayEquals(array1,array4), "they should not be equal")
}