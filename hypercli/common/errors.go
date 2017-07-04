//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import "errors"

var (
	ErrInvalidParams  = errors.New("hypercli/cmd: invalid params")
	ErrInvalidArgsNum = errors.New("hypercli/cmd: invalid args num")
	ErrInvalidCmd     = errors.New("hypercli/cmd: invalid command")
	ErrEmptyHeader    = errors.New("Empty authorization header.")
)
