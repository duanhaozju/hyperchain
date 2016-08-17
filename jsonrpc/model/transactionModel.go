package model//定义交易
import (
	"time"

	//"github.com/syndtr/goleveldb/leveldb"
	//"fmt"
	"encoding/json"
	//"fmt"
	"strconv"
	//"fmt"
	"crypto/sha256"
	"encoding/hex"
	"hyperchain-alpha/utils"
)

type Transaction struct {
	Hash string `json:"hash"`
	From      string    `json:"from"`
	To string      `json:"to"`
	Value  int    `json:"value"`
	TimeStamp  time.Time `json:"timestamp"`
}

type Transactions []Transaction

var currentKeyId int

var transactions Transactions

// 方便测试生成一些种子数据

//-- 初始化数据库
func InitDB(port int)  {
	utils.DBPATH.Path = strconv.Itoa(port)
	currentKeyId = getTransactionCountFromDB()
	//fmt.Println("currid: " , currentKeyId)
}

func SaveTransction(t Transaction) Transaction {
	key := "transation" + strconv.Itoa(currentKeyId)
	err := PutTransactionToDB(key, t)
	if err != nil {
		return t
	}
	currentKeyId += 1
	//t.Id = currentKeyId
	t.TimeStamp = time.Now()
	t.Value =t.Value
	transactions = append(transactions, t)
	return t
}

//-- 将交易信息t以Json的形式存储在数据库中
//-- 参数：key 交易信息对应的key
//-- 返回：error 执行时可能遇到的错误,主要来自连接数据库和对象序列化成json
//--            没有错误返回nil
func PutTransactionToDB(key string, t Transaction) error {
	//db, err := leveldb.OpenFile("hyperchain.cn/app/db", nil)

	if t.Hash == "" {
		hash := sha256.New()
		hash.Write([]byte(t.From + t.To + strconv.Itoa(t.Value) + t.TimeStamp.Format("2006-01-02 15:04:05")))
		md := hash.Sum(nil)
		mdStr := hex.EncodeToString(md)
		t.Hash = mdStr
	}
	db, err := utils.GetConnection()
	defer db.Close()
	if err != nil {
		return err
	}
	byteValue, err := json.Marshal(t)
	if err != nil {
		return err
	}
	err = db.Put([]byte(key), byteValue, nil)
	if err != nil {
		return err
	}
	return nil
}

//-- 根据key获取交易信息,当key不存在时,返回一个ErrNotFound的错误
//-- 当遇到错误时返回nil, err
//-- 否则返回t, nil
func GetTransactionFromDB(key string) (Transaction, error) {
	var t Transaction;
	db, err := utils.GetConnection()
	defer db.Close()
	if err != nil {
		return t, err
	}
	data, err := db.Get([]byte(key), nil)
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(data, &t)
	if err != nil {
		return t, err
	}
	return t, nil
}

//-- 从删除对应key的值
//-- 遇到错误返回,操作成功返回nil
func DeleteTransactionFromDB(key string) error {
	db, err := utils.GetConnection()
	defer db.Close()
	if err != nil {
		return  err
	}
	err = db.Delete([]byte(key), nil)
	if err != nil {
		return  err
	}
	return nil
}

//-- 更新数据库中key对应的值
//-- 参数：key 交易信息对应的key
//-- 返回：error 执行时可能遇到的错误,主要来自连接数据库和对象序列化成json
//--            没有错误返回nil
func UpdatatTransactionFromDB(key string, t Transaction) error {
	return PutTransactionToDB(key, t)
}

func GetAllTransaction() ([]Transaction, error) {
	var ts []Transaction
	db, err := utils.GetConnection()
	defer db.Close()
	if err != nil {
		return  ts, err
	}

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		var t Transaction
		value := iter.Value()
		err = json.Unmarshal(value, &t)
		ts = append(ts, t)
	}
	iter.Release()
	err = iter.Error()
	return ts, err
}

func getTransactionCountFromDB() int {
	db, err := utils.GetConnection()
	defer db.Close()
	if err != nil {
		return  -1
	}
	count := 0
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		count++
	}
	iter.Release()
	//err = iter.Error()
	return count
}

func GetTransactionByHash(hash string) (Transaction, error) {
	var result Transaction
	db, err := utils.GetConnection()
	defer db.Close()
	if err != nil {
		return result,err
	}
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		var t Transaction
		value := iter.Value()
		err = json.Unmarshal(value, &t)
		if t.Hash == hash {
			result = t
			break
		}
	}
	iter.Release()
	err = iter.Error()
	return result, err
}