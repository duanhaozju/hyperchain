package version

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"hyperchain/cmd/dbcli/constant"
	"hyperchain/cmd/dbcli/database"
	"hyperchain/cmd/dbcli/utils"
	"hyperchain/cmd/dbcli/version/wrapper"
	"hyperchain/common"
	"strconv"
	"strings"
)

type Version struct {
	db database.Database
}

func NewVersion(db database.Database) *Version {
	return &Version{
		db: db,
	}
}

var (
	BlockNumPrefix           = []byte("blockNum-")
	BlockPrefix              = []byte("block-")
	TransactionPrefix        = []byte("transaction-")
	InvalidTransactionPrefix = []byte("invalidtransaction-")
	TxMetaSuffix             = []byte{0x01}
	ReceiptsPrefix           = []byte("receipts-")
	ChainKey                 = []byte("chain-key-")
)

/*-----------------------------------block--------------------------------------*/
func (self *Version) GetBlockByNumber(num uint64, parameter *constant.Parameter) (string, error) {
	keyNum := strconv.FormatUint(num, 10)
	v, err := self.db.Get(append(BlockNumPrefix, keyNum...))
	if err != nil {
		return "", err
	} else {
		return self.getBlockByHash(v, parameter)
	}
}

func (self *Version) getBlockByHash(hash []byte, parameter *constant.Parameter) (string, error) {
	key := append(BlockPrefix, hash...)
	data, err := self.db.Get(key)
	if err != nil {
		return "", err
	}
	var blockWrapper wrapper.BlockWrapper
	err = proto.Unmarshal(data, &blockWrapper)
	if err != nil {
		return "", err
	}
	result, err := NewResultFactory(constant.BLOCK, string(blockWrapper.BlockVersion), blockWrapper.Block, parameter)
	if err != nil {
		return "", err
	} else {
		return result, nil
	}

}

func (self *Version) GetBlockByHash(hash string, parameter *constant.Parameter) (string, error) {
	return self.getBlockByHash(common.Hex2Bytes(hash), parameter)
}

func (self *Version) GetBlockHashByNum(num uint64) (string, error) {
	keyNum := strconv.FormatUint(num, 10)
	v, err := self.db.Get(append(BlockNumPrefix, keyNum...))
	if err != nil {
		return "", err
	} else {
		str := fmt.Sprintf("{\n\t\"Number\": %v,\n\t\"BlockHash\": \"%v\"\n}", num, common.Bytes2Hex(v))
		return str, nil
	}
}

func (self *Version) GetBlockRange(path string, min, max uint64, parameter *constant.Parameter) {
	chainHeight, err := self.GetChainHeight()
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
	height, err := strconv.ParseUint(chainHeight, 10, 64)
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
	if height < max {
		max = height
	}
	if path != "" {
		utils.CreateOrAppend(path, "[")
	} else {
		fmt.Println(utils.Decorate("["))
	}
	for i := min; i <= max; i++ {
		result, err := self.GetBlockByNumber(i, parameter)
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
		} else {
			if i == max {
				result = strings.Replace("\t" + result, "\n", "\n\t", -1)
			} else {
				result = strings.Replace("\t" + result + ",", "\n", "\n\t", -1)
			}
			if path != "" {
				utils.CreateOrAppend(path, result)
			} else {
				fmt.Println(utils.Decorate(result))
			}
		}
	}
	if path != "" {
		utils.CreateOrAppend(path, "]")
	} else {
		fmt.Println(utils.Decorate("]"))
	}
}

func (self *Version) GetAllBlockSequential(path string, parameter *constant.Parameter) {
	height, err := self.GetChainHeight()
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
	max, err := strconv.ParseUint(height, 10, 64)
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
	self.GetBlockRange(path, 0, max, parameter)
}

/*-------------------------------------transaction--------------------------------------*/
func (self *Version) GetTransaction(hash string) (string, error) {
	var transactionWrapper wrapper.TransactionWrapper
	keyFact := append(TransactionPrefix, common.Hex2Bytes(hash)...)
	data, err := self.db.Get(keyFact)
	if err != nil {
		return "", err
	}
	err = proto.Unmarshal(data, &transactionWrapper)
	if err != nil {
		return "", err
	}
	result, err := NewResultFactory(constant.TRANSACTION, string(transactionWrapper.TransactionVersion), transactionWrapper.Transaction, nil)
	if err != nil {
		return "", err
	} else {
		return result, nil
	}
}

func (self *Version) GetAllTransaction(path string) {
	iter := self.db.NewIterator(TransactionPrefix)
	if path != "" {
		utils.CreateOrAppend(path, "[")
	} else {
		fmt.Println(utils.Decorate("["))
	}
	var result1 string
	for iter.Next() {
		var transactionWrapper wrapper.TransactionWrapper
		value := iter.Value()
		err := proto.Unmarshal(value, &transactionWrapper)
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
			break
		}
		result, err := NewResultFactory(constant.TRANSACTION, string(transactionWrapper.TransactionVersion), transactionWrapper.Transaction, nil)
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
			break
		} else if result1 != "" {
			result1 = strings.Replace("\t" + result1 + ",", "\n", "\n\t", -1)
			if path != "" {
				utils.CreateOrAppend(path, result1)
			} else {
				fmt.Println(utils.Decorate(result1))
			}
		}
		result1 = result
	}
	if result1 != "" {
		result1 = strings.Replace("\t" + result1, "\n", "\n\t", -1)
		if path != "" {
			utils.CreateOrAppend(path, result1)
		} else {
			fmt.Println(utils.Decorate(result1))
		}
	}
	if path != "" {
		utils.CreateOrAppend(path, "]")
	} else {
		fmt.Println(utils.Decorate("]"))
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
}

func (self *Version) GetTxWithBlock(hash string) (string, error) {
	dataMeta, err := self.db.Get(append(common.Hex2Bytes(hash), TxMetaSuffix...))
	if err != nil {
		return "", err
	}
	result, err := NewResultFactory(constant.TRANSACTIONMETA, constant.VERSIONFINAL, dataMeta, nil)
	if err != nil {
		return "", err
	} else {
		return result, nil
	}
}

func (self *Version) GetDiscardTransaction(hash string) (string, error) {
	data, err := self.db.Get(append(InvalidTransactionPrefix, common.Hex2Bytes(hash)...))
	if err != nil {
		return "", err
	}
	//todo version1.1
	result, err := NewResultFactory(constant.INVAILDTRANSACTION, constant.VERSIONFINAL, data, nil)
	if err != nil {
		return "", err
	} else {
		return result, nil
	}
}

func (self *Version) GetAllDiscardTransaction(path string) {
	iter := self.db.NewIterator(InvalidTransactionPrefix)
	if path != "" {
		utils.CreateOrAppend(path, "[")
	} else {
		fmt.Println(utils.Decorate("["))
	}
	var result1 string
	for iter.Next() {
		data := iter.Value()
		//todo version1.1
		result, err := NewResultFactory(constant.INVAILDTRANSACTION, constant.VERSIONFINAL, data, nil)
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
			break
		} else if result1 != "" {
			result1 = strings.Replace("\t" + result1 + ",", "\n", "\n\t", -1)
			if path != "" {
				utils.CreateOrAppend(path, result)
			} else {
				fmt.Println(utils.Decorate(result))
			}
		}
		result1 = result
	}
	if result1 != "" {
		result1 = strings.Replace("\t" + result1, "\n", "\n\t", -1)
		if path != "" {
			utils.CreateOrAppend(path, result1)
		} else {
			fmt.Println(utils.Decorate(result1))
		}
	}
	if path != "" {
		utils.CreateOrAppend(path, "]")
	} else {
		fmt.Println(utils.Decorate("]"))
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
}

/*-------------------------------------receipt--------------------------------------*/
func (self *Version) GetReceipt(hash string) (string, error) {
	var receiptWrapper wrapper.ReceiptWrapper
	keyFact := append(ReceiptsPrefix, common.Hex2Bytes(hash)...)
	data, err := self.db.Get(keyFact)
	if err != nil {
		return "", err
	}
	err = proto.Unmarshal(data, &receiptWrapper)
	if err != nil {
		return "", err
	}
	result, err := NewResultFactory(constant.RECEIPT, string(receiptWrapper.ReceiptVersion), receiptWrapper.Receipt, nil)
	if err != nil {
		return "", err
	} else {
		return result, nil
	}
}

/*-------------------------------------chain--------------------------------------*/
func (self *Version) GetChain() (string, error) {
	var result string
	var err error
	data, err := self.db.Get(ChainKey)
	if err != nil {
		return "", err
	}
	for _, version := range constant.VERSIONS {
		result, err = NewResultFactory(constant.CHAIN, version, data, nil)
		if !(err != nil && strings.Contains(err.Error(), constant.PROTOERR)) {
			break
		}
	}
	if err != nil {
		return "", err
	} else {
		return result, nil
	}
}

func (self *Version) GetChainHeight() (string, error) {
	var result string
	var err error
	data, err := self.db.Get(ChainKey)
	if err != nil {
		return "", err
	}
	for _, version := range constant.VERSIONS {
		result, err = NewResultFactory(constant.CHAINHEIGHT, version, data, nil)
		if !(err != nil && strings.Contains(err.Error(), constant.PROTOERR)) {
			break
		}
	}
	if err != nil {
		return "", err
	} else {
		return result, nil
	}
}

/*-------------------------------------self--------------------------------------*/
func (self *Version) GetDB() database.Database {
	return self.db
}