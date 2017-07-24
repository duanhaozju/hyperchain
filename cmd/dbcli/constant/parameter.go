package constant

type Parameter struct {
	Verbose bool
}

func (parameter *Parameter) GetVerbose() bool  {
	return parameter.Verbose
}

func (parameter *Parameter) SetVerbose(verbose bool) {
	parameter.Verbose = verbose
}