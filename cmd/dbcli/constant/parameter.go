package constant

type Parameter struct {
	Verbose bool
	TxIndex int
}

func (parameter *Parameter) GetVerbose() bool {
	return parameter.Verbose
}

func (parameter *Parameter) SetVerbose(verbose bool) {
	parameter.Verbose = verbose
}

func (parameter *Parameter) GetTxIndex() int {
	return parameter.TxIndex
}

func (parameter *Parameter) SetTxIndex(txIndex int) {
	parameter.TxIndex = txIndex
}
