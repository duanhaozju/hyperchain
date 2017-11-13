package version

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperchain/hyperchain/cmd/dbcli/constant"
	"github.com/hyperchain/hyperchain/cmd/dbcli/database"
	"github.com/hyperchain/hyperchain/cmd/dbcli/utils"
	"github.com/hyperchain/hyperchain/cmd/dbcli/version/versionAccount"
	"github.com/hyperchain/hyperchain/cmd/dbcli/version/wrapper"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/ledger/state"
	"github.com/hyperchain/hyperchain/hyperdb"
	"os"
	"regexp"
	"sort"
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
	InvalidTransactionPrefix = []byte("invalidtransaction-")
	TxMetaSuffix             = []byte{0x01}
	ReceiptsPrefix           = []byte("receipts-")
	ChainKey                 = []byte("chain-key-")

	accountIdentifier = "-account"
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
	var file *os.File
	defer utils.Close(file)
	if path != "" {
		file = utils.CreateOrAppend(path, "[")
	} else {
		fmt.Println(utils.Decorate("["))
	}
	for i := min; i <= max; i++ {
		result, err := self.GetBlockByNumber(i, parameter)
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
		} else {
			if i == max {
				result = strings.Replace("\t"+result, "\n", "\n\t", -1)
			} else {
				result = strings.Replace("\t"+result+",", "\n", "\n\t", -1)
			}
			if path != "" {
				utils.Append(file, result)
			} else {
				fmt.Println(utils.Decorate(result))
			}
		}
	}
	if path != "" {
		utils.Append(file, "]")
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
func (self *Version) GetTransactionByHash(hash string) (string, error) {
	txBlockStr, err := self.GetTxWithBlock(hash)
	if err != nil {
		return "", err
	}
	blockIndex := 0
	txIndex := 0
	blockIndexReg, err := regexp.Compile("\"BlockIndex\": [0-9]+")
	if err != nil {
		return "", err
	}
	if blockIndexReg.MatchString(txBlockStr) {
		temp := blockIndexReg.FindString(txBlockStr)[len("\"BlockIndex\": "):]
		index := strings.Index(",", temp)
		if index == -1 {
			blockIndex, err = strconv.Atoi(temp)
			if err != nil {
				return "", err
			}
		} else {
			blockIndex, err = strconv.Atoi(temp[:index])
			if err != nil {
				return "", err
			}
		}
	}
	txIndexReg, err := regexp.Compile("\"Index\": [0-9]+")
	if err != nil {
		return "", err
	}
	if txIndexReg.MatchString(txBlockStr) {
		temp := txIndexReg.FindString(txBlockStr)[len("\"Index\": "):]
		index := strings.Index(",", temp)
		if index == -1 {
			txIndex, err = strconv.Atoi(temp)
			if err != nil {
				return "", err
			}
		} else {
			txIndex, err = strconv.Atoi(temp[:index])
			if err != nil {
				return "", err
			}
		}
	}
	keyNum := strconv.FormatInt(int64(blockIndex), 10)
	blockHash, err := self.db.Get(append(BlockNumPrefix, keyNum...))
	if err != nil {
		return "", err
	}
	data, err := self.db.Get(append(BlockPrefix, blockHash...))
	if err != nil {
		return "", err
	}
	var blockWrapper wrapper.BlockWrapper
	err = proto.Unmarshal(data, &blockWrapper)
	if err != nil {
		return "", err
	}
	parameter := &constant.Parameter{
		TxIndex: txIndex,
	}
	result, err := NewResultFactory(constant.TRANSACTION, string(blockWrapper.BlockVersion), blockWrapper.Block, parameter)
	if err != nil {
		return "", err
	} else {
		return result, nil
	}
}

func (self *Version) GetAllTransactionSequential(path string, parameter *constant.Parameter) {
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
	var file *os.File
	defer utils.Close(file)
	if path != "" {
		file = utils.CreateOrAppend(path, "[")
	} else {
		fmt.Println(utils.Decorate("["))
	}
	var result1 string
	for i := uint64(0); i <= max; i++ {
		block, err := self.GetBlockByNumber(i, parameter)
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
		} else {
			reg, err := regexp.Compile("(\"Transactions\": \\[){1}.*(\\]){1}")
			if err != nil {
				fmt.Println(constant.ErrQuery.Error(), err.Error())
			}
			result := reg.FindString(strings.Replace(block, "\n", "||", -1))
			if result != "" {
				result = strings.Replace(result, "||", "\n", -1)
				reg2, err := regexp.Compile(`{[^}]*}`)
				if err != nil {
					fmt.Println(constant.ErrQuery.Error(), err.Error())
				}
				txs := reg2.FindAllString(result, -1)
				for j := 0; j < len(txs); j++ {
					if result1 != "" {
						result1 = strings.Replace("\t"+strings.TrimSpace(result1)+",", "\t\t", "\t", -1)
						if path != "" {
							utils.Append(file, result1)
						} else {
							fmt.Println(utils.Decorate(result1))
						}
					}
					result1 = txs[j]
				}
			}
		}
	}
	if result1 != "" {
		result1 = strings.Replace("\t"+strings.TrimSpace(result1), "\t\t", "\t", -1)
		if path != "" {
			utils.Append(file, result1)
		} else {
			fmt.Println(utils.Decorate(result1))
		}
	}
	if path != "" {
		utils.Append(file, "]")
	} else {
		fmt.Println(utils.Decorate("]"))
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
	var file *os.File
	defer utils.Close(file)
	if path != "" {
		file = utils.CreateOrAppend(path, "[")
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
			result1 = strings.Replace("\t"+result1+",", "\n", "\n\t", -1)
			if path != "" {
				utils.Append(file, result)
			} else {
				fmt.Println(utils.Decorate(result))
			}
		}
		result1 = result
	}
	if result1 != "" {
		result1 = strings.Replace("\t"+result1, "\n", "\n\t", -1)
		if path != "" {
			utils.Append(file, result1)
		} else {
			fmt.Println(utils.Decorate(result1))
		}
	}
	if path != "" {
		utils.Append(file, "]")
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

/*-------------------------------------account--------------------------------------*/
func (self *Version) GetAccountByAddress(address, path string, parameter *constant.Parameter) {
	data, err := self.db.Get(state.CompositeAccountKey(common.Hex2Bytes(address)))
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
	var file *os.File
	defer utils.Close(file)
	if path != "" {
		file = utils.Open(path)
	}
	var account versionAccount.Account
	err = json.Unmarshal(data, &account)
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
	if parameter.GetVerbose() {
		code, _ := self.db.Get(state.CompositeCodeHash(common.Hex2Bytes(address), account.CodeHash))
		prefixRes := account.EncodeVerbose(address, common.Bytes2Hex(code))
		if path != "" {
			utils.Append(file, prefixRes)
		} else {
			fmt.Println(utils.Decorate(prefixRes))
		}
		storageIt := self.db.NewIterator(state.GetStorageKeyPrefix(common.Hex2Bytes(address)))
		var storageRes string
		for storageIt.Next() {
			storageKey, _ := state.SplitCompositeStorageKey(common.Hex2Bytes(address), storageIt.Key())
			if storageRes != "" {
				storageRes += ","
				if path != "" {
					utils.Append(file, storageRes)
				} else {
					fmt.Println(utils.Decorate(storageRes))
				}
			}
			storageRes = "\t\t\t" + "\"" + common.Bytes2Hex(storageKey) + "\"" + ":" + " " + "\"" + common.Bytes2Hex(storageIt.Value()) + "\""
		}
		if storageRes != "" {
			if path != "" {
				utils.Append(file, storageRes)
				utils.Append(file, "\t\t}")
				utils.Append(file, "\t}")
				utils.Append(file, "}")
			} else {
				fmt.Println(utils.Decorate(storageRes))
				fmt.Println(utils.Decorate("\t\t}"))
				fmt.Println(utils.Decorate("\t}"))
				fmt.Println(utils.Decorate("}"))
			}
		}
	} else {
		res := account.Encode(address)
		if path != "" {
			utils.Append(file, res)
		} else {
			fmt.Println(utils.Decorate(res))
		}
	}
}

func (self *Version) GetAllAccount(path string, parameter *constant.Parameter) {
	var file *os.File
	defer utils.Close(file)
	if path != "" {
		file = utils.CreateOrAppend(path, "{")
	} else {
		fmt.Println(utils.Decorate("{"))
	}

	var accountAddress []string
	iter := self.db.NewIterator([]byte(accountIdentifier))
	for iter.Next() {
		address, ok := state.SplitCompositeAccountKey(iter.Key())
		if ok == false {
			fmt.Println(constant.ErrQuery.Error(), "split account address error.")
			continue
		}
		accountAddress = append(accountAddress, common.Bytes2Hex(address))
	}
	sort.Strings(accountAddress)

	var result1 string
	for index := 0; index < len(accountAddress); index++ {
		data, err := self.db.Get(state.CompositeAccountKey(common.Hex2Bytes(accountAddress[index])))
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
			break
		}
		var account versionAccount.Account
		err = json.Unmarshal(data, &account)
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
			break
		}
		if parameter.GetVerbose() {
			if result1 != "" {
				result1 = "\t" + result1 + ","
				if path != "" {
					utils.Append(file, result1)
				} else {
					fmt.Println(utils.Decorate(result1))
				}
			}
			code, _ := self.db.Get(state.CompositeCodeHash(common.Hex2Bytes(accountAddress[index]), account.CodeHash))
			prefixRes := account.EncodeVerbose(accountAddress[index], common.Bytes2Hex(code))
			prefixRes = strings.Replace("\t"+prefixRes, "\n", "\n\t", -1)
			if path != "" {
				utils.Append(file, prefixRes)
			} else {
				fmt.Println(utils.Decorate(prefixRes))
			}
			storageIt := self.db.NewIterator(state.GetStorageKeyPrefix(common.Hex2Bytes(accountAddress[index])))
			var storageRes string
			for storageIt.Next() {
				storageKey, _ := state.SplitCompositeStorageKey(common.Hex2Bytes(accountAddress[index]), storageIt.Key())
				if storageRes != "" {
					storageRes += ","
					if path != "" {
						utils.Append(file, storageRes)
					} else {
						fmt.Println(utils.Decorate(storageRes))
					}
				}
				var value []byte = make([]byte, len(storageIt.Value()))
				if len(storageIt.Value()) <= common.HashLength {
					copy(value, storageIt.Value())
					value = common.BytesToHash(value).Bytes()
				} else {
					copy(value, storageIt.Value())
				}

				storageRes = "\t\t\t\t" + "\"" + common.Bytes2Hex(storageKey) + "\"" + ":" + " " + "\"" + common.Bytes2Hex(value) + "\""
			}
			if path != "" {
				if storageRes != "" {
					utils.Append(file, storageRes)
				}
				utils.Append(file, "\t\t\t}")
				utils.Append(file, "\t\t}")
			} else {
				if storageRes != "" {
					fmt.Println(utils.Decorate(storageRes))
				}
				fmt.Println(utils.Decorate("\t\t\t}"))
				fmt.Println(utils.Decorate("\t\t}"))
			}
			result1 = "}"
		} else {
			if result1 != "" {
				result1 = strings.Replace("\t"+result1+",", "\n", "\n\t", -1)
				if path != "" {
					utils.Append(file, result1)
				} else {
					fmt.Println(utils.Decorate(result1))
				}
			}
			result1 = account.Encode(accountAddress[index])
		}
	}

	if result1 != "" {
		result1 = strings.Replace("\t"+result1, "\n", "\n\t", -1)
		if path != "" {
			utils.Append(file, result1)
		} else {
			fmt.Println(utils.Decorate(result1))
		}
	}
	if path != "" {
		utils.Append(file, "}")
	} else {
		fmt.Println(utils.Decorate("}"))
	}
}

/*-------------------------------------self--------------------------------------*/
func (self *Version) GetDB() database.Database {
	return self.db
}

func (self *Version) RevertDB(ns, globalConf, path string, number uint64, parameter *constant.Parameter) {
	str, err := self.GetChainHeight()
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
	height, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
	maxBlockStr, err := self.GetBlockByNumber(uint64(height), parameter)
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
	reg, err := regexp.Compile("\"MerkleRoot\": \"[a-zA-Z0-9]+\",")
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
	rootStr := reg.FindString(maxBlockStr)
	root := common.HexToHash(rootStr[len("\"MerkleRoot\": \"") : len(rootStr)-2])

	var targetRoot [][]byte
	for i := int64(height - 1); i >= int64(number); i-- {
		minBlockStr, err := self.GetBlockByNumber(uint64(i), parameter)
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
			return
		}
		needRootStr := reg.FindString(minBlockStr)
		needBlockRoot := common.Hex2Bytes(needRootStr[len("\"MerkleRoot\": \"") : len(rootStr)-2])
		targetRoot = append(targetRoot, needBlockRoot)
	}
	var config *common.Config
	if globalConf == "" {
		config = DefaultConfig()
	} else {
		config = common.NewConfig(globalConf)
	}
	err = common.InitHyperLogger(ns, config)
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
	}

	self.GetDB().Close()

	db, err := hyperdb.NewDatabase(nil, path, "leveldb", ns)
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
	stateDb, err := state.New(root, db, nil, config, uint64(height))
	if err != nil {
		fmt.Println(constant.ErrQuery.Error(), err.Error())
		return
	}
	for i := int64(height - 1); i >= int64(number); i-- {
		batch := db.NewBatch()
		err = stateDb.RevertToJournal(uint64(i), uint64(height), targetRoot[0], batch)
		if err != nil {
			fmt.Println(constant.ErrQuery.Error(), err.Error())
			return
		}
		batch.Write()
		height--
		if len(targetRoot) > 1 {
			targetRoot = targetRoot[1:]
		}
	}
	db.Close()
}

func DefaultConfig() *common.Config {
	config := common.NewRawConfig()
	config.Set(state.StateCapacity, 1000003)
	config.Set(state.StateAggreation, 5)
	config.Set(state.StateMerkleCacheSize, 100000)
	config.Set(state.StateBucketCacheSize, 100000)

	config.Set(state.StateObjectCapacity, 1000003)
	config.Set(state.StateObjectAggreation, 5)
	config.Set(state.StateObjectMerkleCacheSize, 100000)
	config.Set(state.StateObjectBucketCacheSize, 100000)

	config.Set(common.LOG_BASE_LOG_LEVEL, "NOTICE")
	return config
}
