//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSubChs(t *testing.T) {
	ctx1 := context.Background()
	ctx2 := context.Background()
	sub1 := GetSubChs(ctx1)
	sub2 := GetSubChs(ctx2)
	assert.Equal(t, sub1, sub2, "These two SubChs should be equal")
}
