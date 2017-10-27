package gen

import (
	"github.com/hyperchain/hyperchain/common"
	"math/big"
	"strings"
)

func ConvertKeyOfMap(originKeyOfMaps map[string]map[string][][]string, contractMap map[string]*ContractContent) (map[string]map[string][][]string, map[string]map[string]map[string]string) {
	var keysOfMapContract map[string]map[string][][]string
	var keyToOriMapContract map[string]map[string]map[string]string

	keysOfMapContract = make(map[string]map[string][][]string)
	keyToOriMapContract = make(map[string]map[string]map[string]string)

	for name, v := range contractMap {
		if originKeyOfMap, ok := originKeyOfMaps[name]; ok {
			a, b := convertKeyOfMap(originKeyOfMap, v.MapContent)
			keysOfMapContract[name] = a
			keyToOriMapContract[name] = b
		}
	}
	return keysOfMapContract, keyToOriMapContract
}

func convertKeyOfMap(originKeyOfMap map[string][][]string, mapContent map[string]string) (map[string][][]string, map[string]map[string]string) {
	var keysOfMap map[string][][]string
	var keyTypesOfMap map[string][]string
	var keyToOriMap map[string]map[string]string

	keysOfMap = make(map[string][][]string)
	keyTypesOfMap = make(map[string][]string)
	keyToOriMap = make(map[string]map[string]string)

	for key, value := range mapContent {
		var typesOfKey []string
		for {
			firIndex := strings.Index(value, "(")
			secIndex := strings.Index(value, "=")
			if firIndex == -1 || secIndex == -1 {
				break
			}
			typesOfKey = append(typesOfKey, value[firIndex+1:secIndex])
			value = value[secIndex+1:]
		}
		keyTypesOfMap[key] = typesOfKey
	}

	for key, value := range originKeyOfMap {
		var res1 [][]string
		var keyToOri map[string]string
		keyToOri = make(map[string]string)
		if _, ok := mapContent[key]; ok {
			for i := 0; i < len(value); i++ {
				var res2 []string
				str := value[i]
				for j := 0; j < len(str); j++ {
					res2 = append(res2, getConvertKey(str[j], keyTypesOfMap[key][j]))
					keyToOri[getConvertKey(str[j], keyTypesOfMap[key][j])] = str[j]
				}
				res1 = append(res1, res2)
			}
			keysOfMap[key] = res1
			keyToOriMap[key] = keyToOri
		}
	}
	return keysOfMap, keyToOriMap
}

func getConvertKey(str string, typeOfKey string) string {
	var res string
	switch typeOfKey {
	case byte_Type:
		fallthrough
	case bytes1_Type:
		fallthrough
	case bytes2_Type:
		fallthrough
	case bytes3_Type:
		fallthrough
	case bytes4_Type:
		fallthrough
	case bytes5_Type:
		fallthrough
	case bytes6_Type:
		fallthrough
	case bytes7_Type:
		fallthrough
	case bytes8_Type:
		fallthrough
	case bytes9_Type:
		fallthrough
	case bytes10_Type:
		fallthrough
	case bytes11_Type:
		fallthrough
	case bytes12_Type:
		fallthrough
	case bytes13_Type:
		fallthrough
	case bytes14_Type:
		fallthrough
	case bytes15_Type:
		fallthrough
	case bytes16_Type:
		fallthrough
	case bytes17_Type:
		fallthrough
	case bytes18_Type:
		fallthrough
	case bytes19_Type:
		fallthrough
	case bytes20_Type:
		fallthrough
	case bytes21_Type:
		fallthrough
	case bytes22_Type:
		fallthrough
	case bytes23_Type:
		fallthrough
	case bytes24_Type:
		fallthrough
	case bytes25_Type:
		fallthrough
	case bytes26_Type:
		fallthrough
	case bytes27_Type:
		fallthrough
	case bytes28_Type:
		fallthrough
	case bytes29_Type:
		fallthrough
	case bytes30_Type:
		fallthrough
	case bytes31_Type:
		fallthrough
	case bytes32_Type:
		res = common.Bytes2Hex(common.RightPadBytes([]byte(str), 32))
	case uint_Type:
		fallthrough
	case uint8_Type:
		fallthrough
	case uint16_Type:
		fallthrough
	case uint24_Type:
		fallthrough
	case uint32_Type:
		fallthrough
	case uint40_Type:
		fallthrough
	case uint48_Type:
		fallthrough
	case uint56_Type:
		fallthrough
	case uint64_Type:
		fallthrough
	case uint72_Type:
		fallthrough
	case uint80_Type:
		fallthrough
	case uint88_Type:
		fallthrough
	case uint96_Type:
		fallthrough
	case uint104_Type:
		fallthrough
	case uint112_Type:
		fallthrough
	case uint120_Type:
		fallthrough
	case uint128_Type:
		fallthrough
	case uint136_Type:
		fallthrough
	case uint144_Type:
		fallthrough
	case uint152_Type:
		fallthrough
	case uint160_Type:
		fallthrough
	case uint168_Type:
		fallthrough
	case uint176_Type:
		fallthrough
	case uint184_Type:
		fallthrough
	case uint192_Type:
		fallthrough
	case uint200_Type:
		fallthrough
	case uint208_Type:
		fallthrough
	case uint216_Type:
		fallthrough
	case uint224_Type:
		fallthrough
	case uint232_Type:
		fallthrough
	case uint240_Type:
		fallthrough
	case uint248_Type:
		fallthrough
	case uint256_Type:
		aa1 := big.NewInt(0)
		aa1.SetString(str, 0)
		res = common.Bytes2Hex(common.LeftPadBytes(aa1.Bytes(), 32))
	case address_Type:
		str16 := str
		if len(str) >= 2 && str[:2] == "0x" {
			str16 = str[2:]
		}
		if len(str16)%2 != 0 {
			str16 = "0" + str16
		}
		temp := common.Hex2Bytes(str16)
		res = common.Bytes2Hex(common.LeftPadBytes(temp, 32))
	case int_Type:
		fallthrough
	case int8_Type:
		fallthrough
	case int16_Type:
		fallthrough
	case int24_Type:
		fallthrough
	case int32_Type:
		fallthrough
	case int40_Type:
		fallthrough
	case int48_Type:
		fallthrough
	case int56_Type:
		fallthrough
	case int64_Type:
		fallthrough
	case int72_Type:
		fallthrough
	case int80_Type:
		fallthrough
	case int88_Type:
		fallthrough
	case int96_Type:
		fallthrough
	case int104_Type:
		fallthrough
	case int112_Type:
		fallthrough
	case int120_Type:
		fallthrough
	case int128_Type:
		fallthrough
	case int136_Type:
		fallthrough
	case int144_Type:
		fallthrough
	case int152_Type:
		fallthrough
	case int160_Type:
		fallthrough
	case int168_Type:
		fallthrough
	case int176_Type:
		fallthrough
	case int184_Type:
		fallthrough
	case int192_Type:
		fallthrough
	case int200_Type:
		fallthrough
	case int208_Type:
		fallthrough
	case int216_Type:
		fallthrough
	case int224_Type:
		fallthrough
	case int232_Type:
		fallthrough
	case int240_Type:
		fallthrough
	case int248_Type:
		fallthrough
	case int256_Type:
		aa1 := big.NewInt(0)
		aa1.SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)

		aa2 := big.NewInt(0)
		aa2.SetString("-1", 10)

		aa3 := big.NewInt(0)
		aa3.SetString(str, 0)
		aa1.Sub(aa1, aa2.Sub(aa2, aa3))

		res = common.Bytes2Hex(aa1.Bytes()[len(aa1.Bytes())-32:])
	}

	return res
}
