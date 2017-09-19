package types

type BoolOfSolidity struct {
	Base
}

func (boolOfSolidity *BoolOfSolidity) Decode() string {
	length := len(boolOfSolidity.Value)
	left := length - (boolOfSolidity.BitIndex+boolOfSolidity.BitNum)/4
	right := length - boolOfSolidity.BitIndex/4
	return boolOfSolidity.Value[left:right]
}
