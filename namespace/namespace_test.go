//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"testing"
	"sync"
	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	ast := assert.New(t)
	status := &Status{
		state: newed,
		desc:  "newed",
		lock:  new(sync.RWMutex),
	}
	ast.Equal(newed, status.getState(), "getState should get the current state of status")
	status.setState(running)
	ast.Equal(running, status.getState(), "After setState, state should be changed to running")
	ast.Equal("running", status.desc, "After setState, desc should be changed to 'running'")
}