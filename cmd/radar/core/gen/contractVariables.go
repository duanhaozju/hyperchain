package gen

import (
	"errors"
	"fmt"
	rcm "github.com/hyperchain/hyperchain/cmd/radar/core/common"
	"github.com/hyperchain/hyperchain/cmd/radar/core/types"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/state"
	"github.com/hyperchain/hyperchain/crypto"
	"github.com/syndtr/goleveldb/leveldb"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type Status struct {
	variableKey    string
	variableName   string
	contractKey    string
	contractNowKey string
	remainType     string
	isNewSlot      bool
	isAddress      bool
	isNewDoneSlot  bool
	keyOfMap       string
	BelongTo       uint
	BelongToFlag   bool
}

func GetContractVariables(contractContent *ContractContent, db *leveldb.DB, contractAddress string) []*types.ContractVariable {
	var contractVariables []*types.ContractVariable
	for i := 0; i < len(contractContent.ContractVariablesContent); i++ {
		str := strings.Split(contractContent.ContractVariablesContent[i], " ")
		status := Status{
			variableKey:    str[0],
			variableName:   contractContent.ContractVariablesContent[i][len(str[0])+1:],
			contractKey:    str[0],
			contractNowKey: "",
			remainType:     str[0],
			isNewSlot:      false,
			isAddress:      false,
			isNewDoneSlot:  false,
			keyOfMap:       "",
		}
		contractVariables = deal(contractVariables, contractContent, status, db, contractAddress)
	}
	return contractVariables
}

func deal(contractVariables []*types.ContractVariable, contractContent *ContractContent, status Status, db *leveldb.DB, contractAddress string) []*types.ContractVariable {
	fixedSizeArrayReg, err := regexp.Compile("^([^\\[\\]])+([\\[][\\d]+[\\]])+([\\[][\\]]){0,0}$")
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	dynamicArrayReg, err := regexp.Compile("[\\[\\]]+")
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	if strings.Contains(status.variableKey, "mapping") {
		status.isNewSlot = true
		status.isAddress = true
		status.isNewDoneSlot = true
		var id uint = 0
		if len(contractVariables) != 0 {
			id = contractVariables[len(contractVariables)-1].Id + 1
		}
		if !status.isAddress {
			status.remainType = ""
		}
		status.variableKey = uint256_Type
		variable := GenerateContractVariableFactory(status, id)
		variable.RemainType = status.remainType
		contractVariables = append(contractVariables, variable)
	} else if contentValue, ok := contractContent.StructContent[status.variableKey]; ok == true {
		str := strings.Split(contentValue, " ")
		for i := 0; i < len(str)/2; i++ {
			var temp = status
			if i == 0 {
				temp.isNewSlot = true
			} else {
				temp.isNewSlot = false
			}
			if i == len(str)/2-1 {
				temp.isNewDoneSlot = true
			} else {
				temp.isNewDoneSlot = false
			}
			temp.variableKey = str[i*2]
			temp.contractNowKey = str[i*2+1]
			temp.remainType = str[i*2]
			contractVariables = deal(contractVariables, contractContent, temp, db, contractAddress)
		}
	} else if contentValue, ok := contractContent.EnumContent[status.variableKey]; ok == true {
		str := strings.Split(contentValue, ",")
		len := len(str)
		count := 0
		for len != 0 {
			len /= 2
			count++
		}
		for count%8 != 0 {
			count++
		}
		status.variableKey = "uint" + strconv.Itoa(count)
		status.remainType = ""
		contractVariables = deal(contractVariables, contractContent, status, db, contractAddress)
	} else {
		if fixedSizeArrayReg.MatchString(status.remainType) {
			firIndex := strings.Index(status.remainType, "[")
			lasIndex := strings.LastIndex(status.remainType, "[")
			var pre, use, remain string
			if firIndex != -1 && lasIndex != -1 {
				pre = status.remainType[:firIndex]
				use = status.remainType[lasIndex:]
				remain = status.remainType[firIndex:lasIndex]
			} else {
				pre = status.remainType
				use = ""
				remain = ""
			}
			length, _ := strconv.ParseUint(use[1:len(use)-1], 0, 64)
			var i uint64
			for i = 0; i < length; i++ {
				var temp = status
				if i == 0 {
					temp.isNewSlot = true
				} else {
					temp.isNewSlot = false
				}
				if i == length-1 {
					temp.isNewDoneSlot = true
				} else {
					temp.isNewDoneSlot = false
				}
				temp.variableKey = pre
				temp.contractNowKey = status.contractNowKey + " " + pre + use + strconv.Itoa(int(i))
				temp.remainType = pre + remain
				contractVariables = deal(contractVariables, contractContent, temp, db, contractAddress)
			}
		} else if dynamicArrayReg.MatchString(status.variableKey) {
			status.isAddress = true
			status.isNewSlot = true
			status.isNewDoneSlot = true
			firIndex := strings.Index(status.variableKey, "[")
			lasIndex := strings.LastIndex(status.variableKey, "[")
			pre := status.variableKey[:firIndex]
			remain := status.variableKey[firIndex:lasIndex]
			if len(status.variableKey[lasIndex:]) > 2 {
				temp := status.variableKey[lasIndex:]
				length, _ := strconv.ParseUint(temp[1:len(temp)-1], 0, 64)
				var i uint64
				for i = 0; i < length; i++ {
					temp := status
					temp.variableKey = uint256_Type
					temp.contractNowKey = status.contractNowKey + " " + pre + status.variableKey[lasIndex:] + strconv.Itoa(int(i))
					temp.remainType = pre + remain
					contractVariables = deal(contractVariables, contractContent, temp, db, contractAddress)
				}
			} else {
				var id uint = 0
				if len(contractVariables) != 0 {
					id = contractVariables[len(contractVariables)-1].Id + 1
				}
				if !status.isAddress {
					status.remainType = ""
				}
				status.contractNowKey = status.contractNowKey + " " + pre + status.variableKey[lasIndex:]
				status.remainType = pre + remain
				status.variableKey = uint256_Type
				variable := GenerateContractVariableFactory(status, id)
				contractVariables = append(contractVariables, variable)
			}
		} else {
			var id uint = 0
			if len(contractVariables) != 0 {
				id = contractVariables[len(contractVariables)-1].Id + 1
			}
			if !status.isAddress {
				status.remainType = ""
			}
			if status.variableKey == string_Type {
				status.isAddress = true
			}
			variable := GenerateContractVariableFactory(status, id)
			contractVariables = append(contractVariables, variable)
		}
	}
	return contractVariables
}

func DealDynamic(contractVariable *types.ContractVariable, contractVariables []*types.ContractVariable, contractContent *ContractContent, variableName, contractType, contractNowType, contractRemainType string, isNewSlot, isAddress, isNewDoneSlot bool, db *leveldb.DB, contractAddress string, keysOfMap map[string][][]string, firMap bool, index int, belongTo uint) ([]*types.ContractVariable, error) {
	firIndex := strings.Index(contractRemainType, "[")
	lasIndex := strings.LastIndex(contractRemainType, "[")
	var pre, use, remain string
	if firIndex != -1 && lasIndex != -1 {
		pre = contractRemainType[:firIndex]
		use = contractRemainType[lasIndex:]
		remain = contractRemainType[firIndex:lasIndex]
	} else {
		pre = contractRemainType
		use = ""
		remain = ""
	}
	dbKey := state.CompositeStorageKey(common.Hex2Bytes(contractAddress), contractVariable.StartAddressOfSlot)
	temp, err := db.Get(dbKey, nil)
	if err != nil && !strings.Contains(err.Error(), "leveldb: not found") {
		return contractVariables, err
	}
	temp = rcm.LeftPaddingZeroByte(temp)
	str := common.Bytes2Hex(temp)
	if len(str) == 0 {
		str = "0000000000000000000000000000000000000000000000000000000000000000"
	}
	contractVariable.Variable.SetValue(str)
	if contractVariable.Variable.GetKey() == string_Type {
		value := str
		tmp, _ := strconv.ParseUint(value[len(value)-1:], 16, 64)
		if tmp%2 != 0 {
			length, _ := strconv.ParseInt(value, 16, 64)
			tempLength := (float64(length) - 1) / 64
			length = int64(math.Ceil(float64(tempLength)))
			kec256Hash := crypto.NewKeccak256Hash("keccak256")
			hexAddress := kec256Hash.ByteHash(contractVariable.StartAddressOfSlot).Hex()
			startAddress := big.NewInt(0)
			startAddress.SetString(hexAddress, 0)
			isNewSlot = true
			isNewDoneSlot = true
			isAddress = false
			preLength := len(contractVariables)
			for i := 0; i < int(length); i++ {
				status := Status{
					variableKey:    string_Type,
					variableName:   variableName,
					contractKey:    contractType,
					contractNowKey: contractNowType,
					remainType:     "",
					isNewSlot:      isNewSlot,
					isAddress:      isAddress,
					isNewDoneSlot:  isNewDoneSlot,
					keyOfMap:       "",
					BelongTo:       belongTo,
					BelongToFlag:   true,
				}
				contractVariables = deal(contractVariables, contractContent, status, db, contractAddress)
			}
			nowLength := len(contractVariables)
			for j := preLength; j < nowLength; j++ {
				variable := contractVariables[j]
				variable.Variable.SetIsAddress(false)
				variable.SetStartAddressOfSlot((rcm.LeftPaddingZero(startAddress)))
				startAddress = startAddress.Add(startAddress, big.NewInt(1))
			}
		}
	} else if strings.Contains(contractVariable.RemainType, "mapping") {
		if len(keysOfMap[contractVariable.Name]) > 0 {
			var length int
			if firMap {
				length = len(keysOfMap[contractVariable.Name])
			} else {
				length = 1
			}
			for i := 0; i < length; i++ {
				keys := keysOfMap[contractVariable.Name][index+i]
				if len(keys) == 0 {
					return contractVariables, errors.New("key of map does not match contract!")
				}
				key := keys[0]
				keysOfMap[contractVariable.Name][index+i] = keys[1:]
				kec256Hash := crypto.NewKeccak256Hash("keccak256")
				hexAddress := kec256Hash.ByteHash([]byte(common.Hex2Bytes(key + common.Bytes2Hex(contractVariable.StartAddressOfSlot)))).Hex()
				startAddress := big.NewInt(0)
				startAddress.SetString(hexAddress, 0)
				index1 := strings.Index(contractVariable.RemainType, ">")
				index2 := strings.LastIndex(contractVariable.RemainType, ")")
				remainTypeOfMap := contractVariable.RemainType[index1+1 : index2]
				bit := 0
				preLength := len(contractVariables)
				status := Status{
					variableKey:    remainTypeOfMap,
					variableName:   variableName,
					contractKey:    contractVariable.Key,
					contractNowKey: "",
					remainType:     remainTypeOfMap,
					isNewSlot:      false,
					isAddress:      false,
					isNewDoneSlot:  false,
					keyOfMap:       contractVariable.KeyOfMap + " " + key,
					BelongTo:       belongTo,
					BelongToFlag:   true,
				}
				contractVariables = deal(contractVariables, contractContent, status, db, contractAddress)
				nowLength := len(contractVariables)
				for j := preLength; j < nowLength; j++ {
					variable := contractVariables[j]
					preContractVariable := contractVariables[j-1]
					if bit == 0 || (!variable.Variable.GetIsNewSlot() && !preContractVariable.Variable.GetIsNewDoneSlot() && variable.Variable.GetBitNum()+bit <= 256) {
						variable.SetStartAddressOfSlot((rcm.LeftPaddingZero(startAddress)))
						variable.Variable.SetBitIndex(bit)
						bit += variable.Variable.GetBitNum()
					} else {
						startAddress = startAddress.Add(startAddress, big.NewInt(1))
						variable.SetStartAddressOfSlot((rcm.LeftPaddingZero(startAddress)))
						variable.Variable.SetBitIndex(0)
						bit = variable.Variable.GetBitNum()
					}
				}
			}
		}
	} else if firIndex != -1 && len(contractRemainType[lasIndex:]) > 2 { //[][6]
		length, _ := strconv.ParseUint(str, 16, 64)
		kec256Hash := crypto.NewKeccak256Hash("keccak256")
		hexAddress := kec256Hash.ByteHash(contractVariable.StartAddressOfSlot).Hex()
		startAddress := big.NewInt(0)
		startAddress.SetString(hexAddress, 0)
		var i uint64 = 0
		bit := 0
		preLength := len(contractVariables)
		for ; i < length; i++ {
			status := Status{
				variableKey:    contractRemainType,
				variableName:   variableName,
				contractKey:    contractType,
				contractNowKey: contractNowType + " " + pre + use + strconv.Itoa(int(i)),
				remainType:     contractRemainType,
				isNewSlot:      isNewSlot,
				isAddress:      isAddress,
				isNewDoneSlot:  isNewDoneSlot,
				keyOfMap:       "",
				BelongTo:       belongTo,
				BelongToFlag:   true,
			}
			contractVariables = deal(contractVariables, contractContent, status, db, contractAddress)
		}
		nowLength := len(contractVariables)
		for j := preLength; j < nowLength; j++ {
			variable := contractVariables[j]
			preContractVariable := contractVariables[j-1]
			if bit == 0 || (!variable.Variable.GetIsNewSlot() && !preContractVariable.Variable.GetIsNewDoneSlot() && variable.Variable.GetBitNum()+bit <= 256) {
				variable.SetStartAddressOfSlot((rcm.LeftPaddingZero(startAddress)))
				variable.Variable.SetBitIndex(bit)
				bit += variable.Variable.GetBitNum()
			} else {
				startAddress = startAddress.Add(startAddress, big.NewInt(1))
				variable.SetStartAddressOfSlot((rcm.LeftPaddingZero(startAddress)))
				variable.Variable.SetBitIndex(0)
				bit = variable.Variable.GetBitNum()
			}
		}
	} else { //[][]
		length, _ := strconv.ParseUint(str, 16, 64)
		kec256Hash := crypto.NewKeccak256Hash("keccak256")
		hexAddress := kec256Hash.ByteHash(contractVariable.StartAddressOfSlot).Hex()
		startAddress := big.NewInt(0)
		startAddress.SetString(hexAddress, 0)
		var i uint64 = 0
		bit := 0
		for ; i < length; i++ {
			var id uint = 0
			if len(contractVariables) != 0 {
				id = contractVariables[len(contractVariables)-1].Id + 1
			}
			isAddress = true
			remainType := pre + remain
			if firIndex == -1 {
				remainType = remain
				isAddress = false
			}
			if isAddress {
				isNewSlot = true
				isNewDoneSlot = true
				status := Status{
					variableKey:    uint256_Type,
					variableName:   variableName,
					contractKey:    contractType,
					contractNowKey: contractNowType + " " + pre + use + strconv.Itoa(int(i)),
					remainType:     remainType,
					isNewSlot:      isNewSlot,
					isAddress:      isAddress,
					isNewDoneSlot:  isNewDoneSlot,
					keyOfMap:       contractVariable.KeyOfMap,
					BelongTo:       belongTo,
					BelongToFlag:   true,
				}
				variable := GenerateContractVariableFactory(status, id)
				variable.SetStartAddressOfSlot((rcm.LeftPaddingZero(startAddress)))
				variable.Variable.SetBitIndex(0)
				startAddress = startAddress.Add(startAddress, big.NewInt(1))
				contractVariables = append(contractVariables, variable)
			} else {
				if i == 0 {
					isNewSlot = true
				} else {
					isNewSlot = false
				}
				if i == length-1 {
					isNewDoneSlot = true
				} else {
					isNewDoneSlot = false
				}
				preLength := len(contractVariables)
				status := Status{
					variableKey:    pre,
					variableName:   variableName,
					contractKey:    contractType,
					contractNowKey: contractNowType + " " + pre + use + strconv.Itoa(int(i)),
					remainType:     remainType,
					isNewSlot:      isNewSlot,
					isAddress:      isAddress,
					isNewDoneSlot:  isNewDoneSlot,
					keyOfMap:       contractVariable.KeyOfMap,
					BelongTo:       belongTo,
					BelongToFlag:   true,
				}
				contractVariables = deal(contractVariables, contractContent, status, db, contractAddress)
				nowLength := len(contractVariables)
				for j := preLength; j < nowLength; j++ {
					variable := contractVariables[j]
					preContractVariable := contractVariables[j-1]
					if bit == 0 || (!variable.Variable.GetIsNewSlot() && !preContractVariable.Variable.GetIsNewDoneSlot() && variable.Variable.GetBitNum()+bit <= 256) {
						variable.SetStartAddressOfSlot((rcm.LeftPaddingZero(startAddress)))
						variable.Variable.SetBitIndex(bit)
						bit += variable.Variable.GetBitNum()
					} else {
						startAddress = startAddress.Add(startAddress, big.NewInt(1))
						variable.SetStartAddressOfSlot((rcm.LeftPaddingZero(startAddress)))
						variable.Variable.SetBitIndex(0)
						bit = variable.Variable.GetBitNum()
					}
				}
			}
		}
	}
	return contractVariables, nil
}
