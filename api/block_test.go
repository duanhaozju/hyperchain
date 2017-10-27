//author:zsx
//data:2016-11-10
package api

import (
	"fmt"
	//"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core"
	"github.com/hyperchain/hyperchain/hyperdb"
	"testing"
)

func Test_GetBlocks(t *testing.T) {
	//单例数据库状态设置为close
	hyperdb.Setclose()
	//初始化数据 正确的
	core.InitDB("./build/keystore", 8001)

	//获取db句柄
	db1, _ := hyperdb.GetDBDatabase()

	//单例数据库状态设置为close
	hyperdb.Setclose()
	//初始化数据库未使用的
	core.InitDB("./build/keystore1", 8002)

	//获取db句柄
	db3, _ := hyperdb.GetDBDatabase()

	publicBlockAPI1 := NewPublicBlockAPI(db1)

	publicBlockAPI3 := NewPublicBlockAPI(db3)

	arg1 := IntervalArgs{}

	from := BlockNumber(0)
	to := BlockNumber(1)

	arg2 := IntervalArgs{
		From: &from,
	}

	arg3 := IntervalArgs{
		To: &to,
	}

	arg4 := IntervalArgs{
		From: &from,
		To:   &to,
	}

	arg5 := IntervalArgs{
		From: &to,
		To:   &from,
	}

	ref, err := publicBlockAPI3.GetBlocks(arg1)
	if err.Error() != "leveldb: not found" {
		t.Errorf("publicBlockAPI3.GetBlocks fail 获取空数据库未返回错误")
	}

	ref, err = publicBlockAPI1.GetBlocks(arg1)
	if err != nil {
		t.Errorf("publicBlockAPI1.GetBlocks fail 未成功获取区块")
	}
	fmt.Println("区块哈希：")
	fmt.Println(ref[0].Hash)

	ref, err = publicBlockAPI1.GetBlocks(arg2)
	if err != nil {
		t.Errorf("publicBlockAPI1.GetBlocks fail 未成功获取区块")
	}

	ref, err = publicBlockAPI1.GetBlocks(arg3)
	if err != nil {
		t.Errorf("publicBlockAPI1.GetBlocks fail 未成功获取区块")
	}

	ref, err = publicBlockAPI1.GetBlocks(arg4)
	if err != nil {
		t.Errorf("publicBlockAPI1.GetBlocks fail 未成功获取区块")
	}

	ref, err = publicBlockAPI1.GetBlocks(arg4)
	if err != nil {
		t.Errorf("publicBlockAPI1.GetBlocks fail 未成功获取区块")
	}

	ref, err = publicBlockAPI1.GetBlocks(arg5)
	if err.Error() != "Invalid params" {
		t.Errorf("publicBlockAPI1.GetBlocks fail 错误参数返回值不对")
	}

	ref, err = publicBlockAPI3.GetBlocks(arg4)
	if err.Error() != "leveldb: not found" {
		t.Errorf("publicBlockAPI3.GetBlocks fail 错误参数返回值不对")
	}
}

func Test_LatestBlock(t *testing.T) {
	//单例数据库状态设置为close
	hyperdb.Setclose()
	//初始化数据 正确的
	core.InitDB("./build/keystore", 8002)

	//获取db句柄
	db1, err := hyperdb.GetDBDatabase()
	if err != nil {
		fmt.Println("GetDBDatabase err ：" + err.Error())
	}

	publicBlockAPI1 := NewPublicBlockAPI(db1)

	_, err = publicBlockAPI1.LatestBlock()
	if err != nil {
		fmt.Println("publicBlockAPI1.LatestBlock() fail err:" + err.Error())
		t.Errorf("publicBlockAPI1.LatestBlock() fail")
	}

}
func Test_GetBlockByHash(t *testing.T) {

	//获取db句柄
	db1, err := hyperdb.GetDBDatabase()

	if err != nil {
		fmt.Println("GetDBDatabase err ：" + err.Error())
	}

	publicBlockAPI1 := NewPublicBlockAPI(db1)

	hash := [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	hash1 := [32]byte{1, 2, 3, 4, 5}

	ref, err := publicBlockAPI1.GetBlockByHash(hash)
	if ref.Hash != hash {
		t.Errorf("publicBlockAPI1.GetBlockByHash(hash) fail 获得的哈希和要取的不相等")
	}

	ref, err = publicBlockAPI1.GetBlockByHash(hash1)
	if ref != nil {
		t.Errorf("publicBlockAPI1.GetBlockByHash(hash1) fail 不存在的哈希取到了值")
	}

}

func Test_GetBlockByNumber(t *testing.T) {
	//获取db句柄
	db1, err := hyperdb.GetDBDatabase()

	if err != nil {
		fmt.Println("GetDBDatabase err ：" + err.Error())
	}

	publicBlockAPI1 := NewPublicBlockAPI(db1)

	num := BlockNumber(0)
	num1 := BlockNumber(5)

	ref, err := publicBlockAPI1.GetBlockByNumber(num)
	if err != nil && ref.Number.ToUint64() != uint64(0) {
		t.Errorf("publicBlockAPI1.GetBlockByNumber fail ")
	}

	//第五块不存在
	ref, err = publicBlockAPI1.GetBlockByNumber(num1)
	if err.Error() != "leveldb: not found" {
		t.Errorf("publicBlockAPI1.GetBlockByNumber fail 错误区块返回错误为空")
	}
}

//提高覆盖率
func Test_QueryCommitAndBatchTime(t *testing.T) {
	//获取db句柄
	db1, err := hyperdb.GetDBDatabase()

	if err != nil {
		fmt.Println("GetDBDatabase err ：" + err.Error())
	}

	publicBlockAPI1 := NewPublicBlockAPI(db1)

	args := SendQueryArgs{
		From: Number(1),
		To:   Number(2),
	}

	publicBlockAPI1.QueryCommitAndBatchTime(args)

	publicBlockAPI1.QueryEvmAvgTime(args)
}
