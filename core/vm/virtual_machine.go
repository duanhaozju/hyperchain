package vm

// VirtualMachine is an EVM interface
type VirtualMachine interface {
	Run(*Contract, []byte) ([]byte, error)
}
