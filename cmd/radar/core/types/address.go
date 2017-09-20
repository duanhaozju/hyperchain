package types

type AddressOfSolidity struct {
	Base
}

func (addressOfSolidity *AddressOfSolidity) Decode() string {
	length := len(addressOfSolidity.Value)
	left := length - (addressOfSolidity.BitIndex+addressOfSolidity.BitNum)/4
	right := length - addressOfSolidity.BitIndex/4
	return addressOfSolidity.Value[left:right]
}
