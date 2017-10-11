package runtime_test

import (
	"fmt"
	"hyperchain/common"
	"hyperchain/core/vm/evm/runtime"
	"hyperchain/hyperdb/mdb"
)

func ExampleExecute() {
	conf := common.NewRawConfig()
	common.InitHyperLogger(common.DEFAULT_NAMESPACE, conf)
	common.GetLogger(common.DEFAULT_NAMESPACE, "state")
	common.SetLogLevel(common.DEFAULT_NAMESPACE, "state", "NOTICE")
	common.GetLogger(common.DEFAULT_NAMESPACE, "buckettree")
	common.SetLogLevel(common.DEFAULT_NAMESPACE, "buckettree", "NOTICE")

	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	ret, _, _, err := runtime.Execute(db, common.Hex2Bytes("6060604052600a8060106000396000f360606040526008565b00"), nil, nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ret)
	// Output:
	// [96 96 96 64 82 96 8 86 91 0]
}
