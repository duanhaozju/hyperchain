// Hash interface defined
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25
package crypto

import "hyperchain/common"

// hash interface
type CommonHash interface {
     Hash(x interface{}) (h common.Hash)
}
