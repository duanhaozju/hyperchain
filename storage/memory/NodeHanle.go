package memory

import (
	"hyperchain-alpha/core/node"
	"fmt"
)

func (n node.Node) Save(){
	fmt.Println("保存到内存了")
}
