package runtime_test

import (
	"fmt"

	"hyperchain/common"
	"hyperchain/core/vm/evm/runtime"
	"hyperchain/hyperdb/mdb"
)

func ExampleExecute() {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	ret, _, err := runtime.Execute(db, common.Hex2Bytes("6060604052600a8060106000396000f360606040526008565b00"), nil, nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ret)
	// Output:
	// [96 96 96 64 82 96 8 86 91 0]
}
