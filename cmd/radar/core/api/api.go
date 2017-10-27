package api

import (
	rcm "github.com/hyperchain/hyperchain/cmd/radar/core/common"
	"github.com/hyperchain/hyperchain/cmd/radar/core/gen"
	"github.com/hyperchain/hyperchain/cmd/radar/core/types"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
)

func GetResult(contractFilePath string, db *leveldb.DB, contractAddress string, originKeyOfMaps map[string]map[string][][]string) (map[string][]string, error) {
	contractMap := gen.GetContentOfAllContracts(contractFilePath)
	keysOfMapContract, keyToOriMapContract := gen.ConvertKeyOfMap(originKeyOfMaps, contractMap)
	var analysisResult map[string][]string
	analysisResult = make(map[string][]string)
	for name, v := range contractMap {
		//if storage , ok := storages[name]; ok {
		contractVariables := gen.GetContractVariables(v, db, contractAddress)
		startAddress := big.NewInt(0)
		bit := 0
		for i := 0; i < len(contractVariables); i++ {
			var preContractVariable *types.ContractVariable
			if i == 0 {
				preContractVariable = contractVariables[i]
			} else {
				preContractVariable = contractVariables[i-1]
			}
			contractVariable := contractVariables[i]
			if bit == 0 || (!contractVariable.Variable.GetIsNewSlot() && !preContractVariable.Variable.GetIsNewDoneSlot() && contractVariable.Variable.GetBitNum()+bit <= 256) {
				contractVariable.SetStartAddressOfSlot((rcm.LeftPaddingZero(startAddress)))
				contractVariable.Variable.SetBitIndex(bit)
				bit += contractVariable.Variable.GetBitNum()
			} else {
				startAddress = startAddress.Add(startAddress, big.NewInt(1))
				contractVariable.SetStartAddressOfSlot((rcm.LeftPaddingZero(startAddress)))
				contractVariable.Variable.SetBitIndex(0)
				bit = contractVariable.Variable.GetBitNum()
			}
		}

		firMap := true
		var firstAppearMap map[string]int
		firstAppearMap = make(map[string]int)
		keysOfMap := keysOfMapContract[name]
		for i := 0; i < len(contractVariables); i++ {
			contractVariable := contractVariables[i]
			length := len(keysOfMap[contractVariable.Name])
			if length == 0 {
				length = 1
			}
			if contractVariable.Variable.GetIsAddress() {
				if _, ok := firstAppearMap[contractVariable.Name]; !ok {
					firstAppearMap[contractVariable.Name] = 0
					firMap = true
				}
				var err error
				contractVariables, err = gen.DealDynamic(contractVariable, contractVariables, v, contractVariable.Name, contractVariable.Key, contractVariable.NowKey, contractVariable.RemainType, false, false, false, db, contractAddress, keysOfMap, firMap, firstAppearMap[contractVariable.Name]%length, contractVariable.Id)
				if err != nil {
					return analysisResult, err
				}
				if firMap == false {
					firstAppearMap[contractVariable.Name]++
				}
				firMap = false
			}
		}

		result := gen.GetResult(db, contractAddress, contractVariables, v, keyToOriMapContract[name])
		analysisResult[name] = result
		//}
	}
	return analysisResult, nil
}
