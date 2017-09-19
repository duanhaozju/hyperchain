package types

import (
	"strconv"
)

type StringOfSolidity struct {
	Base
}

func (stringOfSolidity *StringOfSolidity) Decode() string {
	value := stringOfSolidity.getValue()
	tmp, _ := strconv.ParseUint(value[len(value)-2:], 16, 64)
	if tmp%2 == 0 {
		length := tmp
		res := value[0:length]
		return res
	} else {
		return ""
	}
}
