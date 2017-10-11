package vm

// Vm is the basic interface for an implementation of the VM.
type Vm interface {
	// Run should execute the given contract with the given input
	// and return the contract execution return bytes or an error if it
	// failed.
	Run(c VmContext, in []byte) ([]byte, error)
	Finalize()
}

// Just a symbolic interface
// virtual machine use type assert to convert to real type.
type VmContext interface{}
