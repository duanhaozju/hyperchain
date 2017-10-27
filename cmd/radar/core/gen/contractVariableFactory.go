package gen

import (
	"fmt"
	"github.com/hyperchain/hyperchain/cmd/radar/core/types"
)

func GenerateContractVariableFactory(status Status, id uint) *types.ContractVariable {
	var contractVariable types.ContractVariable
	contractVariable.Name = status.variableName
	contractVariable.Id = id
	contractVariable.Key = status.contractKey
	contractVariable.NowKey = status.contractNowKey
	contractVariable.RemainType = status.remainType
	contractVariable.KeyOfMap = status.keyOfMap
	contractVariable.BelongTo = status.BelongTo
	contractVariable.BelongToFlag = status.BelongToFlag
	switch status.variableKey {
	case uint8_Type:
		integerBase := types.Base{Key: uint8_Type, BitNum: bitNum_8, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint8OfSolidity{integerBase}
	case uint16_Type:
		integerBase := types.Base{Key: uint16_Type, BitNum: bitNum_16, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint16OfSolidity{integerBase}
	case uint24_Type:
		integerBase := types.Base{Key: uint24_Type, BitNum: bitNum_24, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint24OfSolidity{integerBase}
	case uint32_Type:
		integerBase := types.Base{Key: uint32_Type, BitNum: bitNum_32, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint32OfSolidity{integerBase}
	case uint40_Type:
		integerBase := types.Base{Key: uint40_Type, BitNum: bitNum_40, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint40OfSolidity{integerBase}
	case uint48_Type:
		integerBase := types.Base{Key: uint48_Type, BitNum: bitNum_48, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint48OfSolidity{integerBase}
	case uint56_Type:
		integerBase := types.Base{Key: uint56_Type, BitNum: bitNum_56, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint56OfSolidity{integerBase}
	case uint64_Type:
		integerBase := types.Base{Key: uint64_Type, BitNum: bitNum_64, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint64OfSolidity{integerBase}
	case uint72_Type:
		integerBase := types.Base{Key: uint72_Type, BitNum: bitNum_72, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint72OfSolidity{integerBase}
	case uint80_Type:
		integerBase := types.Base{Key: uint80_Type, BitNum: bitNum_80, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint80OfSolidity{integerBase}
	case uint88_Type:
		integerBase := types.Base{Key: uint88_Type, BitNum: bitNum_88, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint88OfSolidity{integerBase}
	case uint96_Type:
		integerBase := types.Base{Key: uint96_Type, BitNum: bitNum_96, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint96OfSolidity{integerBase}
	case uint104_Type:
		integerBase := types.Base{Key: uint104_Type, BitNum: bitNum_104, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint104OfSolidity{integerBase}
	case uint112_Type:
		integerBase := types.Base{Key: uint112_Type, BitNum: bitNum_112, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint112OfSolidity{integerBase}
	case uint120_Type:
		integerBase := types.Base{Key: uint120_Type, BitNum: bitNum_120, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint120OfSolidity{integerBase}
	case uint128_Type:
		integerBase := types.Base{Key: uint128_Type, BitNum: bitNum_128, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint128OfSolidity{integerBase}
	case uint136_Type:
		integerBase := types.Base{Key: uint136_Type, BitNum: bitNum_136, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint136OfSolidity{integerBase}
	case uint144_Type:
		integerBase := types.Base{Key: uint144_Type, BitNum: bitNum_144, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint144OfSolidity{integerBase}
	case uint152_Type:
		integerBase := types.Base{Key: uint152_Type, BitNum: bitNum_152, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint152OfSolidity{integerBase}
	case uint160_Type:
		integerBase := types.Base{Key: uint160_Type, BitNum: bitNum_160, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint160OfSolidity{integerBase}
	case uint168_Type:
		integerBase := types.Base{Key: uint168_Type, BitNum: bitNum_168, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint168OfSolidity{integerBase}
	case uint176_Type:
		integerBase := types.Base{Key: uint176_Type, BitNum: bitNum_176, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint176OfSolidity{integerBase}
	case uint184_Type:
		integerBase := types.Base{Key: uint184_Type, BitNum: bitNum_184, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint184OfSolidity{integerBase}
	case uint192_Type:
		integerBase := types.Base{Key: uint192_Type, BitNum: bitNum_192, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint192OfSolidity{integerBase}
	case uint200_Type:
		integerBase := types.Base{Key: uint200_Type, BitNum: bitNum_200, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint200OfSolidity{integerBase}
	case uint208_Type:
		integerBase := types.Base{Key: uint208_Type, BitNum: bitNum_208, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint208OfSolidity{integerBase}
	case uint216_Type:
		integerBase := types.Base{Key: uint216_Type, BitNum: bitNum_216, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint216OfSolidity{integerBase}
	case uint224_Type:
		integerBase := types.Base{Key: uint224_Type, BitNum: bitNum_224, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint224OfSolidity{integerBase}
	case uint232_Type:
		integerBase := types.Base{Key: uint232_Type, BitNum: bitNum_232, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint232OfSolidity{integerBase}
	case uint240_Type:
		integerBase := types.Base{Key: uint240_Type, BitNum: bitNum_240, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint240OfSolidity{integerBase}
	case uint248_Type:
		integerBase := types.Base{Key: uint248_Type, BitNum: bitNum_248, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint248OfSolidity{integerBase}
	case uint256_Type:
		integerBase := types.Base{Key: uint256_Type, BitNum: bitNum_256, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint256OfSolidity{integerBase}
	case uint_Type:
		integerBase := types.Base{Key: uint256_Type, BitNum: bitNum_256, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Uint256OfSolidity{integerBase}
	case int8_Type:
		integerBase := types.Base{Key: int8_Type, BitNum: bitNum_8, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int8OfSolidity{integerBase}
	case int16_Type:
		integerBase := types.Base{Key: int16_Type, BitNum: bitNum_16, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int16OfSolidity{integerBase}
	case int24_Type:
		integerBase := types.Base{Key: int24_Type, BitNum: bitNum_24, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int24OfSolidity{integerBase}
	case int32_Type:
		integerBase := types.Base{Key: int32_Type, BitNum: bitNum_32, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int32OfSolidity{integerBase}
	case int40_Type:
		integerBase := types.Base{Key: int40_Type, BitNum: bitNum_40, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int40OfSolidity{integerBase}
	case int48_Type:
		integerBase := types.Base{Key: int48_Type, BitNum: bitNum_48, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int48OfSolidity{integerBase}
	case int56_Type:
		integerBase := types.Base{Key: int56_Type, BitNum: bitNum_56, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int56OfSolidity{integerBase}
	case int64_Type:
		integerBase := types.Base{Key: int64_Type, BitNum: bitNum_64, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int64OfSolidity{integerBase}
	case int72_Type:
		integerBase := types.Base{Key: int72_Type, BitNum: bitNum_72, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int72OfSolidity{integerBase}
	case int80_Type:
		integerBase := types.Base{Key: int80_Type, BitNum: bitNum_80, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int80OfSolidity{integerBase}
	case int88_Type:
		integerBase := types.Base{Key: int88_Type, BitNum: bitNum_88, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int88OfSolidity{integerBase}
	case int96_Type:
		integerBase := types.Base{Key: int96_Type, BitNum: bitNum_96, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int96OfSolidity{integerBase}
	case int104_Type:
		integerBase := types.Base{Key: int104_Type, BitNum: bitNum_104, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int104OfSolidity{integerBase}
	case int112_Type:
		integerBase := types.Base{Key: int112_Type, BitNum: bitNum_112, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int112OfSolidity{integerBase}
	case int120_Type:
		integerBase := types.Base{Key: int120_Type, BitNum: bitNum_120, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int120OfSolidity{integerBase}
	case int128_Type:
		integerBase := types.Base{Key: int128_Type, BitNum: bitNum_128, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int128OfSolidity{integerBase}
	case int136_Type:
		integerBase := types.Base{Key: int136_Type, BitNum: bitNum_136, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int136OfSolidity{integerBase}
	case int144_Type:
		integerBase := types.Base{Key: int144_Type, BitNum: bitNum_144, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int144OfSolidity{integerBase}
	case int152_Type:
		integerBase := types.Base{Key: int152_Type, BitNum: bitNum_152, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int152OfSolidity{integerBase}
	case int160_Type:
		integerBase := types.Base{Key: int160_Type, BitNum: bitNum_160, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int160OfSolidity{integerBase}
	case int168_Type:
		integerBase := types.Base{Key: int168_Type, BitNum: bitNum_168, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int168OfSolidity{integerBase}
	case int176_Type:
		integerBase := types.Base{Key: int176_Type, BitNum: bitNum_176, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int176OfSolidity{integerBase}
	case int184_Type:
		integerBase := types.Base{Key: int184_Type, BitNum: bitNum_184, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int184OfSolidity{integerBase}
	case int192_Type:
		integerBase := types.Base{Key: int192_Type, BitNum: bitNum_192, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int192OfSolidity{integerBase}
	case int200_Type:
		integerBase := types.Base{Key: int200_Type, BitNum: bitNum_200, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int200OfSolidity{integerBase}
	case int208_Type:
		integerBase := types.Base{Key: int208_Type, BitNum: bitNum_208, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int208OfSolidity{integerBase}
	case int216_Type:
		integerBase := types.Base{Key: int216_Type, BitNum: bitNum_216, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int216OfSolidity{integerBase}
	case int224_Type:
		integerBase := types.Base{Key: int224_Type, BitNum: bitNum_224, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int224OfSolidity{integerBase}
	case int232_Type:
		integerBase := types.Base{Key: int232_Type, BitNum: bitNum_232, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int232OfSolidity{integerBase}
	case int240_Type:
		integerBase := types.Base{Key: int240_Type, BitNum: bitNum_240, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int240OfSolidity{integerBase}
	case int248_Type:
		integerBase := types.Base{Key: int248_Type, BitNum: bitNum_248, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int248OfSolidity{integerBase}
	case int256_Type:
		integerBase := types.Base{Key: int256_Type, BitNum: bitNum_256, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int256OfSolidity{integerBase}
	case int_Type:
		integerBase := types.Base{Key: int256_Type, BitNum: bitNum_256, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Int256OfSolidity{integerBase}
	case bool_Type:
		integerBase := types.Base{Key: bool_Type, BitNum: bitNum_8, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.BoolOfSolidity{integerBase}
	case address_Type:
		integerBase := types.Base{Key: address_Type, BitNum: bitNum_160, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.AddressOfSolidity{integerBase}
	case bytes1_Type:
		integerBase := types.Base{Key: bytes1_Type, BitNum: bitNum_8, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes1OfSolidity{integerBase}
	case bytes2_Type:
		integerBase := types.Base{Key: bytes2_Type, BitNum: bitNum_16, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes2OfSolidity{integerBase}
	case bytes3_Type:
		integerBase := types.Base{Key: bytes3_Type, BitNum: bitNum_24, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes3OfSolidity{integerBase}
	case bytes4_Type:
		integerBase := types.Base{Key: bytes4_Type, BitNum: bitNum_32, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes4OfSolidity{integerBase}
	case bytes5_Type:
		integerBase := types.Base{Key: bytes5_Type, BitNum: bitNum_40, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes5OfSolidity{integerBase}
	case bytes6_Type:
		integerBase := types.Base{Key: bytes6_Type, BitNum: bitNum_48, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes6OfSolidity{integerBase}
	case bytes7_Type:
		integerBase := types.Base{Key: bytes7_Type, BitNum: bitNum_56, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes7OfSolidity{integerBase}
	case bytes8_Type:
		integerBase := types.Base{Key: bytes8_Type, BitNum: bitNum_64, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes8OfSolidity{integerBase}
	case bytes9_Type:
		integerBase := types.Base{Key: bytes9_Type, BitNum: bitNum_72, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes9OfSolidity{integerBase}
	case bytes10_Type:
		integerBase := types.Base{Key: bytes10_Type, BitNum: bitNum_80, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes10OfSolidity{integerBase}
	case bytes11_Type:
		integerBase := types.Base{Key: bytes11_Type, BitNum: bitNum_88, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes11OfSolidity{integerBase}
	case bytes12_Type:
		integerBase := types.Base{Key: bytes12_Type, BitNum: bitNum_96, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes12OfSolidity{integerBase}
	case bytes13_Type:
		integerBase := types.Base{Key: bytes13_Type, BitNum: bitNum_104, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes13OfSolidity{integerBase}
	case bytes14_Type:
		integerBase := types.Base{Key: bytes14_Type, BitNum: bitNum_112, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes14OfSolidity{integerBase}
	case bytes15_Type:
		integerBase := types.Base{Key: bytes15_Type, BitNum: bitNum_120, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes15OfSolidity{integerBase}
	case bytes16_Type:
		integerBase := types.Base{Key: bytes16_Type, BitNum: bitNum_128, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes16OfSolidity{integerBase}
	case bytes17_Type:
		integerBase := types.Base{Key: bytes17_Type, BitNum: bitNum_136, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes17OfSolidity{integerBase}
	case bytes18_Type:
		integerBase := types.Base{Key: bytes18_Type, BitNum: bitNum_144, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes18OfSolidity{integerBase}
	case bytes19_Type:
		integerBase := types.Base{Key: bytes19_Type, BitNum: bitNum_152, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes19OfSolidity{integerBase}
	case bytes20_Type:
		integerBase := types.Base{Key: bytes20_Type, BitNum: bitNum_160, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes20OfSolidity{integerBase}
	case bytes21_Type:
		integerBase := types.Base{Key: bytes21_Type, BitNum: bitNum_168, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes21OfSolidity{integerBase}
	case bytes22_Type:
		integerBase := types.Base{Key: bytes22_Type, BitNum: bitNum_176, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes22OfSolidity{integerBase}
	case bytes23_Type:
		integerBase := types.Base{Key: bytes23_Type, BitNum: bitNum_184, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes23OfSolidity{integerBase}
	case bytes24_Type:
		integerBase := types.Base{Key: bytes24_Type, BitNum: bitNum_192, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes24OfSolidity{integerBase}
	case bytes25_Type:
		integerBase := types.Base{Key: bytes25_Type, BitNum: bitNum_200, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes25OfSolidity{integerBase}
	case bytes26_Type:
		integerBase := types.Base{Key: bytes26_Type, BitNum: bitNum_208, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes26OfSolidity{integerBase}
	case bytes27_Type:
		integerBase := types.Base{Key: bytes27_Type, BitNum: bitNum_216, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes27OfSolidity{integerBase}
	case bytes28_Type:
		integerBase := types.Base{Key: bytes28_Type, BitNum: bitNum_224, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes28OfSolidity{integerBase}
	case bytes29_Type:
		integerBase := types.Base{Key: bytes29_Type, BitNum: bitNum_232, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes29OfSolidity{integerBase}
	case bytes30_Type:
		integerBase := types.Base{Key: bytes30_Type, BitNum: bitNum_240, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes30OfSolidity{integerBase}
	case bytes31_Type:
		integerBase := types.Base{Key: bytes31_Type, BitNum: bitNum_248, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes31OfSolidity{integerBase}
	case bytes32_Type:
		integerBase := types.Base{Key: bytes32_Type, BitNum: bitNum_256, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes32OfSolidity{integerBase}
	case byte_Type:
		integerBase := types.Base{Key: bytes1_Type, BitNum: bitNum_8, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.Bytes1OfSolidity{integerBase}
	case string_Type:
		integerBase := types.Base{Key: string_Type, BitNum: bitNum_256, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.StringOfSolidity{integerBase}
	case bytes_Type:
		integerBase := types.Base{Key: string_Type, BitNum: bitNum_256, AddressNum: AddressNum_1, IsNewSlot: status.isNewSlot, IsAddress: status.isAddress, IsNewDoneSlot: status.isNewDoneSlot}
		contractVariable.Variable = &types.StringOfSolidity{integerBase}
	default:
		fmt.Println("err type: ", status.variableKey)
	}

	return &contractVariable
}
