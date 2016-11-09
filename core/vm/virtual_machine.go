//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package vm

// VirtualMachine is an EVM interface
type VirtualMachine interface {
	Run(*Contract, []byte) ([]byte, error)
}
