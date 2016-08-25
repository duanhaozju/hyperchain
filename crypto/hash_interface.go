package crypto

import "hyperchain-alpha/common"

type CommonHash interface {
     Hash(x interface{}) (h common.Hash)
}
