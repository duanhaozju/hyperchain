package core

import (
	"sync"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"log"
	"hyperchain/common"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"

)

//-- --------------------- Balance ------------------------------------\
// Balance account balance
// The key is account public-key-hash
// The value is the account balance value
type BalanceMap map[common.Address][]byte

type stateType int32

const (
	closed stateType = iota
	opened
)

type Balance struct {
	dbBalance    BalanceMap   // store in db, synchronization with block
	cacheBalance BalanceMap   // synchronization with transaction
	lock         sync.RWMutex // the lock for balance of reading of writing
	state        stateType    // the balance state, use for singleton
	stateLock    sync.Mutex   // the lock of get balance instance
}

var balance = &Balance{
	state: closed,
}

// GetBalanceIns 获取Balance单例
// 如果没有创建实例，册创建，创建过程：
// 先从数据库中读，如果读到，则将dbBalance，cacheBalance都赋值为读到的Map
// 如果数据库中没有，则创建一个空的dbBalance和cacheBalance
func GetBalanceIns() (*Balance, error) {
	balance.stateLock.Lock()
	defer balance.stateLock.Unlock()

	if balance.state == closed {
		balance.cacheBalance = make(BalanceMap)
		balance.dbBalance = make(BalanceMap)
		balance.state = opened

		db, err := hyperdb.GetLDBDatabase()
		if err != nil {
			log.Fatal(err)
			return balance, err
		}
		balance_db, err := GetDBBalance(db)
		if err == nil {
			balance.cacheBalance = balance_db
			balance.dbBalance = balance_db
			return balance, nil
		}else if err == leveldb.ErrNotFound {
			return balance, nil
		}else{
			return balance, err
		}
	}
	return balance, nil
}

//var balanceInstance = newBalanceIns()

//-- 将Balance存入内存
func (self *Balance)PutCacheBalance(accountPublicKeyHash common.Address, value []byte) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.cacheBalance[accountPublicKeyHash] = value
}

//-- 在MEM中 根据Key获取的Balance
func (self *Balance)GetCacheBalance(accountPublicKeyHash common.Address) []byte {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.cacheBalance[accountPublicKeyHash]
}

//-- 从MEM中删除Balance
func (self *Balance)DeleteCacheBalance(accountPublicKeyHash common.Address) {
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.cacheBalance, accountPublicKeyHash)
}

//-- 从MEM中获取所有Balance
func (self *Balance)GetAllCacheBalance() (BalanceMap) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	var bs = make(BalanceMap)
	for key, value := range self.cacheBalance {
		bs[key] = value
	}
	return bs
}

//-- 将Balance存入内存
func (self *Balance)PutDBBalance(accountPublicKeyHash common.Address, value []byte) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.dbBalance[accountPublicKeyHash] = value
}

//-- 在MEM中 根据Key获取的Balance
func (self *Balance)GetDBBalance(accountPublicKeyHash common.Address) []byte {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.dbBalance[accountPublicKeyHash]
}

//-- 从MEM中删除Balance
func (self *Balance)DeleteDBBalance(accountPublicKeyHash common.Address) {
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.dbBalance, accountPublicKeyHash)
}

//-- 从MEM中获取所有Balance
func (self *Balance)GetAllDBBalance() (BalanceMap) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	var bs = make(BalanceMap)
	for key, value := range self.dbBalance {
		bs[key] = value
	}
	return bs
}

//-- 更新balance表 需要一个新的区块
func (self *Balance)UpdateDBBalance(block *types.Block) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err
	}
	for _, trans := range block.Transactions {
		//-- 将交易里的From(账户的publickey)余额减去value
		//-- 如果余额表中没有这个From(实际上不可能，因为余额表中没有这个From，不可能会验证通过
		//-- 但是上帝用户例外，因为上帝用户可能会出现负数)，则新建一个
		//-- 如果余额表中有这个From，则覆盖publickey(覆盖的Publickey是一样的，实际上没改)
		var transValue big.Int
		transValue.SetString(string(trans.Value), 10)
		fromBalance := self.cacheBalance[common.BytesToAddress(trans.From)]
		toBalance := self.cacheBalance[common.BytesToAddress(trans.To)]
		var fromValue big.Int
		var toValue big.Int

		fromValue.SetString(string(fromBalance), 10)
		toValue.SetString(string(toBalance), 10)
		//b -= value
		fromValue.Sub(&fromValue, &transValue)
		self.cacheBalance[common.BytesToAddress(trans.From)] = []byte(fromValue.String())
		//-- 将交易中的To(账户中的publickey)余额加上value
		//-- 如果余额表中没有这个To(就是所有publickey不含有To)
		//-- 新建一个balance，将交易的value直接赋给balance.value
		//-- 如果余额表中有这个To,则直接加上交易中的value
		if _, ok := self.cacheBalance[common.BytesToAddress(trans.To)]; ok {
			//fromBalance = self.cacheBalance[trans.To]
			//fromBalance += value
			toValue.Add(&toValue, &transValue)
			self.cacheBalance[common.BytesToAddress(trans.To)] = []byte(toValue.String())
		} else {
			//fromBalance = value
			self.cacheBalance[common.BytesToAddress(trans.To)] = []byte(transValue.String())
		}
	}
	//-- 此时balance_cache与balance_db一致
	self.cacheBalance = self.dbBalance
	//-- 将balance_db更新到数据库中
	err = PutDBBalance(db, self.dbBalance)
	if err != nil {
		return err
	}
	return nil
}

func (self *Balance)UpdateCacheBalance(trans *types.Transaction) {
	self.lock.Lock()
	defer self.lock.Unlock()
	var transValue big.Int
	transValue.SetString(string(trans.Value), 10)
	fromBalance := self.cacheBalance[common.BytesToAddress(trans.From)]
	toBalance := self.cacheBalance[common.BytesToAddress(trans.To)]
	var fromValue big.Int
	var toValue big.Int

	fromValue.SetString(string(fromBalance), 10)
	toValue.SetString(string(toBalance), 10)
	//b -= value
	fromValue.Sub(&fromValue, &transValue)
	self.cacheBalance[common.BytesToAddress(trans.From)] = []byte(fromValue.String())
	if _, ok := self.cacheBalance[common.BytesToAddress(trans.To)]; ok {
		//fromBalance = self.cacheBalance[trans.To]
		//fromBalance += value
		toValue.Add(&toValue, &transValue)
		self.cacheBalance[common.BytesToAddress(trans.To)] = []byte(toValue.String())
	} else {
		//fromBalance = value
		self.cacheBalance[common.BytesToAddress(trans.To)] = []byte(transValue.String())
	}
}


// VerifyTransaction is to verify balance of the tranaction
// If the balance is not enough, returns false
func VerifyBalance(tx types.Transaction) bool{
	var balance big.Int
	var value big.Int

	balanceIns, err := GetBalanceIns()

	if err != nil {
		log.Fatalf("GetBalanceIns error, %v", err)
	}

	bal := balanceIns.GetCacheBalance(common.BytesToAddress(tx.From))

	balance.SetString(string(bal), 10)
	value.SetString(string(tx.Value), 10)

	if value.Cmp(&balance) == 1 {
		return false
	}

	return true
}


//-- --------------------- Balance END --------------------------------
