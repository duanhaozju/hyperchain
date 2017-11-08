package gen

import (
	rcm "github.com/hyperchain/hyperchain/cmd/radar/core/common"
	"github.com/hyperchain/hyperchain/cmd/radar/core/types"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/ledger/state"
	"github.com/hyperchain/hyperchain/crypto"
	"github.com/syndtr/goleveldb/leveldb"
	"math"
	"strconv"
	"strings"
)

func GetResult(db *leveldb.DB, contractAddress string, contractVariables []*types.ContractVariable, contractContent *ContractContent, keyToOriMap map[string]map[string]string) []string {
	var result []string
	var variableNameSet map[string][]*types.ContractVariable
	variableNameSet = make(map[string][]*types.ContractVariable)
	for i := 0; i < len(contractVariables); i++ {
		contractVariable := contractVariables[i]
		if !contractVariable.Variable.GetIsAddress() {
			dbKey := state.CompositeStorageKey(common.Hex2Bytes(contractAddress), contractVariable.StartAddressOfSlot)
			temp, err := db.Get(dbKey, nil)
			if err != nil {
				if !strings.Contains(err.Error(), "leveldb: not found") {
					return result
				}
			}
			temp = rcm.LeftPaddingZeroByte(temp)
			str := common.Bytes2Hex(temp)
			contractVariable.Variable.SetValue(str)
			contractVariables[i] = contractVariable
		}
		variableNameSet[contractVariable.Name] = append(variableNameSet[contractVariable.Name], contractVariable)
	}

	for index := 0; index < len(contractVariables); index++ {
		name := contractVariables[index].Name
		if value, ok := variableNameSet[name]; ok {
			//for name, value := range variableNameSet {
			if strings.Contains(value[0].Key, "mapping") {
				var everyKeyMap map[string][]*types.ContractVariable
				everyKeyMap = make(map[string][]*types.ContractVariable)
				for i := 0; i < len(value); i++ {
					everyKeyMap[value[i].KeyOfMap] = append(everyKeyMap[value[i].KeyOfMap], value[i])
				}
				typeOfValue := value[0].Key
				var typeOfKeys []string
				for strings.Contains(typeOfValue, "mapping") {
					firIndex := strings.Index(typeOfValue, ">")
					lasIndex := strings.LastIndex(typeOfValue, ")")

					index1 := strings.Index(typeOfValue, "(")
					index2 := strings.Index(typeOfValue, "=")
					typeOfKeys = append(typeOfKeys, strings.TrimSpace(typeOfValue[index1+1:index2]))

					typeOfValue = typeOfValue[firIndex+1 : lasIndex]
				}
				for k, v := range everyKeyMap {
					flag := false
					for j := 0; j < len(v); j++ {
						if !v[j].Variable.GetIsAddress() {
							flag = true
						}
					}
					if flag {
						if len(k) != 0 {
							ks := strings.Split(k[1:], " ")
							var ksStr string
							for i := 0; i < len(ks); i++ {
								ksStr += "[" + humanDecodeOfKeyInMap(keyToOriMap[name][ks[i]], typeOfKeys[i]) + "]"
							}
							//result += name + ksStr + "=" + dealV(typeOfValue, v, "%s", contractContent, 0, keyToOriMap) + "\n";
							result = append(result, name+ksStr+"="+dealV(typeOfValue, v, "%s", contractContent, 0, keyToOriMap))
						}
					}
				}
			} else {
				//result += name + "=" + dealV(value[0].Key, value, "%s", contractContent, 0, keyToOriMap) + "\n";
				result = append(result, name+"="+dealV(value[0].Key, value, "%s", contractContent, 0, keyToOriMap))
			}
			delete(variableNameSet, name)
		}
	}
	return result
}

func dealV(remainType string, variableValues []*types.ContractVariable, format string, contractContent *ContractContent, start int, keyToOriMap map[string]map[string]string) string {
	if len(variableValues) < 0 {
		return "null"
	}
	index := strings.Index(remainType, "[")
	if key, ok := contractContent.StructContent[remainType]; ok {
		countS := strings.Count(format, "%s")
		structKeys := strings.Split(key, " ")
		for i := 0; i < countS; i++ {
			var result string
			for i := 0; i < len(structKeys); i = i + 2 {
				typeOfKey := structKeys[i]
				nameOfKey := structKeys[i+1]
				format1 := "%s"
				var everyValue []*types.ContractVariable
				everyValue = append(everyValue, variableValues[start])
				start++
				for j := start; j < len(variableValues); j++ {
					if !variableValues[j].Use {
						if strings.Split(variableValues[j].NowKey, " ")[0] == strings.Split(everyValue[0].NowKey, " ")[0] {
							everyValue = append(everyValue, variableValues[j])
							start++
						} else {
							break
						}
					}
				}
				for j := 0; j < len(everyValue); j++ {
					if everyValue[j].Variable.GetIsAddress() {
						for k := start; k < len(variableValues); k++ {
							if variableValues[k].BelongToFlag && variableValues[k].BelongTo == everyValue[j].Id && !variableValues[k].Use {
								everyValue = append(everyValue, variableValues[k])
							}
						}
					}
				}
				str := dealV(typeOfKey, everyValue, format1, contractContent, 0, keyToOriMap)
				result = result + nameOfKey + ":" + str + ","
			}
			format = strings.Replace(format, "%s", "{"+result[:len(result)-1]+"}", 1)
		}
		return format
	} else if _, ok := contractContent.EnumContent[remainType]; ok {
		for i := start; i < len(variableValues); i++ {
			if variableValues[i].Variable.GetValue() == "" {
				format = strings.Replace(format, "%s", humanDecode("00", remainType, contractContent.EnumContent), 1)
			} else {
				format = strings.Replace(format, "%s", humanDecode(variableValues[i].Variable.Decode(), remainType, contractContent.EnumContent), 1)
			}
			variableValues[i].Use = true
		}
		return format
	} else if index != -1 {
		firIndex := strings.Index(remainType, "[")
		lasIndex := strings.LastIndex(remainType, "[")
		pre := remainType[:firIndex]
		mid := remainType[firIndex:lasIndex]
		las := remainType[lasIndex:]
		if len(las) > 2 {
			length, _ := strconv.ParseUint(las[1:len(las)-1], 10, 64)
			var str string
			var i uint64
			for i = 0; i < length; i++ {
				str += "%s,"
			}
			str = "[" + str[:len(str)-1] + "]"
			format = strings.Replace(format, "%s", str, -1)
			return dealV(pre+mid, variableValues, format, contractContent, start, keyToOriMap)
		} else {
			countS := strings.Count(format, "%s")
			index := strings.Index(format, "%s")
			for i := 0; i < countS; i++ {
				length, _ := strconv.ParseUint(variableValues[start].Variable.GetValue(), 16, 64)
				var str string
				var i uint64
				for i = 0; i < length; i++ {
					str += "%s,"
				}
				if length == 0 {
					str = "[" + str + "]"
				} else {
					str = "[" + str[:len(str)-1] + "]"
				}
				format = format[:index] + strings.Replace(format[index:], "%s", str, 1)
				index = index + len(str) + 1
				variableValues[start].Use = true
				start++
			}
			return dealV(pre+mid, variableValues, format, contractContent, start, keyToOriMap)
		}
	} else if remainType == string_Type {
		for i := start; i < len(variableValues); i++ {
			if variableValues[i].Use == true {
				continue
			}
			value := variableValues[start].Variable.GetValue()
			temp, _ := strconv.ParseUint(value[len(value)-1:], 16, 64)
			if temp%2 == 0 {
				if variableValues[i].Variable.GetValue() == "0000000000000000000000000000000000000000000000000000000000000000" {
					format = strings.Replace(format, "%s", humanDecode("00", variableValues[i].Variable.GetKey(), contractContent.EnumContent), 1)
				} else {
					format = strings.Replace(format, "%s", humanDecode(variableValues[i].Variable.Decode(), variableValues[i].Variable.GetKey(), contractContent.EnumContent), 1)
				}
				variableValues[i].Use = true
			} else {
				strLength, _ := strconv.ParseInt(variableValues[start].Variable.GetValue(), 16, 64)
				variableValues[start].Use = true
				var result string
				tempLength := (float64(strLength) - 1) / 64
				slotSize := int64(math.Ceil(float64(tempLength)))
				kec256Hash := crypto.NewKeccak256Hash("keccak256")
				h := kec256Hash.ByteHash(variableValues[start].StartAddressOfSlot)
				var startOfPart int
				for j := start; j < len(variableValues); j++ {
					if common.Bytes2Hex(variableValues[j].StartAddressOfSlot) == common.Bytes2Hex(h[:]) {
						startOfPart = j
						break
					}
				}
				for j := 0; j < int(slotSize); j++ {
					result += variableValues[startOfPart].Variable.GetValue()
					variableValues[startOfPart].Use = true
					startOfPart++
				}
				format = strings.Replace(format, "%s", humanDecode(result[:strLength-1], string_Type, contractContent.EnumContent), 1)
			}
			start++
		}
		return format
	} else {
		for i := start; i < len(variableValues); i++ {
			if variableValues[i].Variable.GetValue() == "" {
				format = strings.Replace(format, "%s", humanDecode("00", variableValues[i].Variable.GetKey(), contractContent.EnumContent), 1)
			} else {
				format = strings.Replace(format, "%s", humanDecode(variableValues[i].Variable.Decode(), variableValues[i].Variable.GetKey(), contractContent.EnumContent), 1)
			}
			variableValues[i].Use = true
		}
		return format
	}
	return "0"
}

func humanDecode(decodeResult string, typeOfVariable string, enumContent map[string]string) string {
	if v, ok := enumContent[typeOfVariable]; ok {
		index, _ := strconv.ParseInt(decodeResult, 16, 0)
		enums := strings.Split(v, ",")
		return enums[index]
	} else if strings.Contains(typeOfVariable, "byte") || strings.Contains(typeOfVariable, "string") {
		decodeResultCopy := decodeResult
		for {
			if len(decodeResultCopy) < 2 {
				break
			}
			temp := decodeResultCopy[len(decodeResultCopy)-2:]
			if temp == "00" {
				decodeResultCopy = decodeResultCopy[:len(decodeResultCopy)-2]
			} else {
				break
			}
		}
		return "\"" + string(common.Hex2Bytes(decodeResultCopy)) + "\""
	} else if strings.Contains(typeOfVariable, "uint") {
		value, _ := strconv.ParseUint(decodeResult, 16, 64)
		res := strconv.FormatUint(value, 10)
		return res
	} else if strings.Contains(typeOfVariable, "int") {
		value, _ := strconv.ParseInt(decodeResult, 16, 0)
		res := strconv.Itoa(int(value))
		return res
	} else if strings.Contains(typeOfVariable, "address") {
		return common.HexToAddress(decodeResult).Hex()
	} else if strings.Contains(typeOfVariable, "bool") {
		value, _ := strconv.ParseUint(decodeResult, 16, 64)
		var res string = "false"
		if value != 0 {
			res = "true"
		}
		return res
	} else {
		return "not support type1."
	}
}

func humanDecodeOfKeyInMap(decodeResult string, typeOfVariable string) string {
	if strings.Contains(typeOfVariable, "byte") || strings.Contains(typeOfVariable, "string") {
		decodeResultCopy := decodeResult
		return "\"" + decodeResultCopy + "\""
	} else if strings.Contains(typeOfVariable, "uint") || strings.Contains(typeOfVariable, "int") {
		return decodeResult
	} else if strings.Contains(typeOfVariable, "address") {
		return common.HexToAddress(decodeResult).Hex()
	} else {
		return "not support type2."
	}
}
