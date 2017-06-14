//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package evm

// VirtualMachine is an EVM interface
type VirtualMachine interface {
	Run(*Contract, []byte) ([]byte, error)
}
