package core

import (
	"sync"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"hyperchain/common"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"github.com/op/go-logging"
)

//-- --------------------- Balance ------------------------------------\
// Balance account balance
// The key is account public-key-hash
// The value is the account balance value
type BalanceMap map[common.Address][]byte

type stateType int32

const (
	closed stateType = iota // the instance is closed (be not)
	opened
)

// Balance is store all accounts' balance, it is a goroutine safe struct
type Balance struct {
	dbBalance    BalanceMap   // store in db, synchronization with block
	cacheBalance BalanceMap   // synchronization with transaction
	lock         sync.RWMutex // the lock for balance of reading of writing

	state        stateType    // the balance state, use for singleton
	stateLock    sync.Mutex   // the lock of get balance instance
}

// balance is a instance of Balance, balance.state is closed
// when initially, indicate the initial balance instance is not Initialization
var balance = &Balance{
	state: closed,
}

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("core")
}


// GetBalanceIns get balance singleton instance
// if there is no balance instance, it will create one. creating process:
// read dbBalance from db firstly, if exists, dbBalance and cacheBalance
// will be assigned dbBalance which get from db. if not exists,
// it will create empty dbBalance and cacheBalance
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


//-- PutCacheBalance put cacheBalance(just into memory not db)
func (self *Balance)PutCacheBalance(address common.Address, value []byte) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.cacheBalance[address] = value
}

//-- GetCacheBalance get cacheBalance by address
func (self *Balance)GetCacheBalance(address common.Address) []byte {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.cacheBalance[address]
}

//-- DeleteCacheBalance delete cacheBalance for given address
func (self *Balance)DeleteCacheBalance(address common.Address) {
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.cacheBalance, address)
}

//-- GetAllCacheBalance get all cacheBalance
func (self *Balance)GetAllCacheBalance() (BalanceMap) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	var bs = make(BalanceMap)
	for key, value := range self.cacheBalance {
		bs[key] = value
	}
	return bs
}

//-- PutDBBalance put dbbalance (just into memory not db)
func (self *Balance)PutDBBalance(address common.Address, value []byte) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.dbBalance[address] = value
}

//-- GetDBBalance get dbBalance by address
func (self *Balance)GetDBBalance(address common.Address) []byte {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return self.dbBalance[address]
}

//-- DeleteDBBalance deleate dbBalance for given address
func (self *Balance)DeleteDBBalance(address common.Address) {
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.dbBalance, address)
}

//-- GetAllDBBalance get all dbBalance
func (self *Balance)GetAllDBBalance() (BalanceMap) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	var bs = make(BalanceMap)
	for key, value := range self.dbBalance {
		bs[key] = value
	}
	return bs
}

// UpdateDBBalance update dbBalance require a latest block,
// after updating dbBalance, cacheBalance will be equal with dbBalance
func (self *Balance)UpdateDBBalance(block *types.Block) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err
	}

	//for test evm execute contract

		//ExecTransaction(*types.NewTestCreateTransaction())


	for _, trans := range block.Transactions {
		//ExecTransaction(*trans)


		if trans.To ==nil{
			ExecTransaction(*trans)
			continue
		}
		if _, ok := self.dbBalance[common.HexToAddress(string(trans.To))]; !ok {
			ExecTransaction(*trans)
			continue
		}

		var transValue big.Int
		transValue.SetString(string(trans.Value), 10)
		fromBalance := self.dbBalance[common.HexToAddress(string(trans.From))]
		toBalance := self.dbBalance[common.HexToAddress(string(trans.To))]
		var fromValue big.Int
		var toValue big.Int
		fromValue.SetString(string(fromBalance), 10)
		toValue.SetString(string(toBalance), 10)
		fromValue.Sub(&fromValue, &transValue)

		// Update Transaction.From account(sub the From account balance by value)
		self.dbBalance[common.HexToAddress(string(trans.From))] = []byte(fromValue.String())

		// Update Transaction.To account(add the To account balance by value)
		// if Transaction.To account not exist, it will be created, initial account balance is 0
		if _, ok := self.dbBalance[common.HexToAddress(string(trans.To))]; ok {
			toValue.Add(&toValue, &transValue)
			self.dbBalance[common.HexToAddress(string(trans.To))] = []byte(toValue.String())
		} else {
			self.dbBalance[common.HexToAddress(string(trans.To))] = []byte(transValue.String())
		}
	}
	// cacheBalance keep correspondence with dbBalance
	self.cacheBalance = self.dbBalance
	// put dbBalance into database
	err = PutDBBalance(db, self.dbBalance)
	if err != nil {
		return err
	}
	return nil
}

// UpdateCacheBalance　updates cacheBalance by transactions
func (self *Balance)UpdateCacheBalance(trans *types.Transaction) {
	self.lock.Lock()
	defer self.lock.Unlock()
	var transValue big.Int
	transValue.SetString(string(trans.Value), 10)
	fromBalance := self.cacheBalance[common.HexToAddress(string(trans.From))]
	toBalance := self.cacheBalance[common.HexToAddress(string(trans.To))]
	var fromValue big.Int
	var toValue big.Int
	fromValue.SetString(string(fromBalance), 10)
	toValue.SetString(string(toBalance), 10)

	// Update Transaction.From account(sub the From account balance by value)
	fromValue.Sub(&fromValue, &transValue)

	// Update Transaction.To account(add the To account balance by value)
	// if Transaction.To account not exist, it will be created, initial account balance is 0
	self.cacheBalance[common.HexToAddress(string(trans.From))] = []byte(fromValue.String())
	if _, ok := self.cacheBalance[common.HexToAddress(string(trans.To))]; ok {
		toValue.Add(&toValue, &transValue)
		self.cacheBalance[common.HexToAddress(string(trans.To))] = []byte(toValue.String())
	} else {
		self.cacheBalance[common.HexToAddress(string(trans.To))] = []byte(transValue.String())
	}
}


// VerifyTransaction is to verify balance of the tranaction
// If the balance is not enough, returns false
func VerifyBalance(tx *types.Transaction) bool {
	var balance big.Int
	var value big.Int
	balanceIns, err := GetBalanceIns()
	if err != nil {
		log.Fatalf("GetBalanceIns error, %v", err)
	}
	bal := balanceIns.GetCacheBalance(common.HexToAddress(string(tx.From)))
	//log.Println(common.Bytes2Hex(bal))
	balance.SetString(string(bal), 10)
	value.SetString(string(tx.Value), 10)

	//fmt.Println("value: ", value.String())
	//fmt.Println("balance: ", balance.String())
	if value.Cmp(&balance) == 1 {
		return false
	}

	return true
}


//-- --------------------- Balance END --------------------------------
