package main

import (
	res "hyperchain/core/executor/restore"
	"hyperchain/common"
	"fmt"
	"hyperchain/hyperdb"
)

func restore(conf *common.Config, sid string, namespace string) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		fmt.Println("[RESTORE] init db failed.")
		fmt.Println("[RESTORE] detail reason: ", err.Error())
		return
	}
	handler := res.NewRestorer(conf, db, namespace)
	if err := handler.Restore(sid); err != nil {
		fmt.Printf("[RESTORE] restore from snapshot %s failed.\n", sid)
		fmt.Println("[RESTORE] detail reason: ", err.Error())
		return
	}
	fmt.Printf("[RESTORE] restore from snapshot %s success.\n", sid)
}
