package vm
// Vm is the basic interface for an implementation of the EVM.
type Vm interface {
	// Run should execute the given contract with the input given in in
	// and return the contract execution return bytes or an error if it
	// failed.
	Run(c VmContext, in []byte) ([]byte, error)
}
// Just a symbolic interface
// virtual machine use type assert to convert to real type.
type VmContext interface {}
