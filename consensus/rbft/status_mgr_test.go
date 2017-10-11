//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusMgr(t *testing.T) {
	var status RbftStatus
	status.activeState(&status.byzantine)
	status.inActiveState(&status.isNewNode)
	ast := assert.New(t)
	ast.Equal(true, status.getState(&status.byzantine), "activeState failed")
	ast.Equal(false, status.getState(&status.isNewNode), "inActiveState failed")

	ast.Equal(true, status.checkStatesAnd(&status.byzantine), "checkStatesAnd failed")
	ast.Equal(false, status.checkStatesAnd(&status.byzantine, &status.isNewNode), "checkStatesAnd failed")

	ast.Equal(true, status.checkStatesOr(&status.byzantine), "checkStatesOr failed")
	ast.Equal(false, status.checkStatesOr(&status.isNewNode), "checkStatesOr failed")
	ast.Equal(true, status.checkStatesOr(&status.byzantine, &status.isNewNode), "checkStatesOr failed")
}
