package types

type Uint8OfSolidity struct {
	Base
}

func (uint8OfSolidity *Uint8OfSolidity) Decode() string {
	length := len(uint8OfSolidity.Value)
	left := length - (uint8OfSolidity.BitIndex+uint8OfSolidity.BitNum)/4
	right := length - uint8OfSolidity.BitIndex/4
	return uint8OfSolidity.Value[left:right]
}

type Uint16OfSolidity struct {
	Base
}

func (uint16OfSolidity *Uint16OfSolidity) Decode() string {
	length := len(uint16OfSolidity.Value)
	left := length - (uint16OfSolidity.BitIndex+uint16OfSolidity.BitNum)/4
	right := length - uint16OfSolidity.BitIndex/4
	return uint16OfSolidity.Value[left:right]
}

type Uint24OfSolidity struct {
	Base
}

func (uint24OfSolidity *Uint24OfSolidity) Decode() string {
	length := len(uint24OfSolidity.Value)
	left := length - (uint24OfSolidity.BitIndex+uint24OfSolidity.BitNum)/4
	right := length - uint24OfSolidity.BitIndex/4
	return uint24OfSolidity.Value[left:right]
}

type Uint32OfSolidity struct {
	Base
}

func (uint32OfSolidity *Uint32OfSolidity) Decode() string {
	length := len(uint32OfSolidity.Value)
	left := length - (uint32OfSolidity.BitIndex+uint32OfSolidity.BitNum)/4
	right := length - uint32OfSolidity.BitIndex/4
	return uint32OfSolidity.Value[left:right]
}

type Uint40OfSolidity struct {
	Base
}

func (uint40OfSolidity *Uint40OfSolidity) Decode() string {
	length := len(uint40OfSolidity.Value)
	left := length - (uint40OfSolidity.BitIndex+uint40OfSolidity.BitNum)/4
	right := length - uint40OfSolidity.BitIndex/4
	return uint40OfSolidity.Value[left:right]
}

type Uint48OfSolidity struct {
	Base
}

func (uint48OfSolidity *Uint48OfSolidity) Decode() string {
	length := len(uint48OfSolidity.Value)
	left := length - (uint48OfSolidity.BitIndex+uint48OfSolidity.BitNum)/4
	right := length - uint48OfSolidity.BitIndex/4
	return uint48OfSolidity.Value[left:right]
}

type Uint56OfSolidity struct {
	Base
}

func (uint56OfSolidity *Uint56OfSolidity) Decode() string {
	length := len(uint56OfSolidity.Value)
	left := length - (uint56OfSolidity.BitIndex+uint56OfSolidity.BitNum)/4
	right := length - uint56OfSolidity.BitIndex/4
	return uint56OfSolidity.Value[left:right]
}

type Uint64OfSolidity struct {
	Base
}

func (uint64OfSolidity *Uint64OfSolidity) Decode() string {
	length := len(uint64OfSolidity.Value)
	left := length - (uint64OfSolidity.BitIndex+uint64OfSolidity.BitNum)/4
	right := length - uint64OfSolidity.BitIndex/4
	return uint64OfSolidity.Value[left:right]
}

type Uint72OfSolidity struct {
	Base
}

func (uint72OfSolidity *Uint72OfSolidity) Decode() string {
	length := len(uint72OfSolidity.Value)
	left := length - (uint72OfSolidity.BitIndex+uint72OfSolidity.BitNum)/4
	right := length - uint72OfSolidity.BitIndex/4
	return uint72OfSolidity.Value[left:right]
}

type Uint80OfSolidity struct {
	Base
}

func (uint80OfSolidity *Uint80OfSolidity) Decode() string {
	length := len(uint80OfSolidity.Value)
	left := length - (uint80OfSolidity.BitIndex+uint80OfSolidity.BitNum)/4
	right := length - uint80OfSolidity.BitIndex/4
	return uint80OfSolidity.Value[left:right]
}

type Uint88OfSolidity struct {
	Base
}

func (uint88OfSolidity *Uint88OfSolidity) Decode() string {
	length := len(uint88OfSolidity.Value)
	left := length - (uint88OfSolidity.BitIndex+uint88OfSolidity.BitNum)/4
	right := length - uint88OfSolidity.BitIndex/4
	return uint88OfSolidity.Value[left:right]
}

type Uint96OfSolidity struct {
	Base
}

func (uint96OfSolidity *Uint96OfSolidity) Decode() string {
	length := len(uint96OfSolidity.Value)
	left := length - (uint96OfSolidity.BitIndex+uint96OfSolidity.BitNum)/4
	right := length - uint96OfSolidity.BitIndex/4
	return uint96OfSolidity.Value[left:right]
}

type Uint104OfSolidity struct {
	Base
}

func (uint104OfSolidity *Uint104OfSolidity) Decode() string {
	length := len(uint104OfSolidity.Value)
	left := length - (uint104OfSolidity.BitIndex+uint104OfSolidity.BitNum)/4
	right := length - uint104OfSolidity.BitIndex/4
	return uint104OfSolidity.Value[left:right]
}

type Uint112OfSolidity struct {
	Base
}

func (uint112OfSolidity *Uint112OfSolidity) Decode() string {
	length := len(uint112OfSolidity.Value)
	left := length - (uint112OfSolidity.BitIndex+uint112OfSolidity.BitNum)/4
	right := length - uint112OfSolidity.BitIndex/4
	return uint112OfSolidity.Value[left:right]
}

type Uint120OfSolidity struct {
	Base
}

func (uint120OfSolidity *Uint120OfSolidity) Decode() string {
	length := len(uint120OfSolidity.Value)
	left := length - (uint120OfSolidity.BitIndex+uint120OfSolidity.BitNum)/4
	right := length - uint120OfSolidity.BitIndex/4
	return uint120OfSolidity.Value[left:right]
}

type Uint128OfSolidity struct {
	Base
}

func (uint128OfSolidity *Uint128OfSolidity) Decode() string {
	length := len(uint128OfSolidity.Value)
	left := length - (uint128OfSolidity.BitIndex+uint128OfSolidity.BitNum)/4
	right := length - uint128OfSolidity.BitIndex/4
	return uint128OfSolidity.Value[left:right]
}

type Uint136OfSolidity struct {
	Base
}

func (uint136OfSolidity *Uint136OfSolidity) Decode() string {
	length := len(uint136OfSolidity.Value)
	left := length - (uint136OfSolidity.BitIndex+uint136OfSolidity.BitNum)/4
	right := length - uint136OfSolidity.BitIndex/4
	return uint136OfSolidity.Value[left:right]
}

type Uint144OfSolidity struct {
	Base
}

func (uint144OfSolidity *Uint144OfSolidity) Decode() string {
	length := len(uint144OfSolidity.Value)
	left := length - (uint144OfSolidity.BitIndex+uint144OfSolidity.BitNum)/4
	right := length - uint144OfSolidity.BitIndex/4
	return uint144OfSolidity.Value[left:right]
}

type Uint152OfSolidity struct {
	Base
}

func (uint152OfSolidity *Uint152OfSolidity) Decode() string {
	length := len(uint152OfSolidity.Value)
	left := length - (uint152OfSolidity.BitIndex+uint152OfSolidity.BitNum)/4
	right := length - uint152OfSolidity.BitIndex/4
	return uint152OfSolidity.Value[left:right]
}

type Uint160OfSolidity struct {
	Base
}

func (uint160OfSolidity *Uint160OfSolidity) Decode() string {
	length := len(uint160OfSolidity.Value)
	left := length - (uint160OfSolidity.BitIndex+uint160OfSolidity.BitNum)/4
	right := length - uint160OfSolidity.BitIndex/4
	return uint160OfSolidity.Value[left:right]
}

type Uint168OfSolidity struct {
	Base
}

func (uint168OfSolidity *Uint168OfSolidity) Decode() string {
	length := len(uint168OfSolidity.Value)
	left := length - (uint168OfSolidity.BitIndex+uint168OfSolidity.BitNum)/4
	right := length - uint168OfSolidity.BitIndex/4
	return uint168OfSolidity.Value[left:right]
}

type Uint176OfSolidity struct {
	Base
}

func (uint176OfSolidity *Uint176OfSolidity) Decode() string {
	length := len(uint176OfSolidity.Value)
	left := length - (uint176OfSolidity.BitIndex+uint176OfSolidity.BitNum)/4
	right := length - uint176OfSolidity.BitIndex/4
	return uint176OfSolidity.Value[left:right]
}

type Uint184OfSolidity struct {
	Base
}

func (uint184OfSolidity *Uint184OfSolidity) Decode() string {
	length := len(uint184OfSolidity.Value)
	left := length - (uint184OfSolidity.BitIndex+uint184OfSolidity.BitNum)/4
	right := length - uint184OfSolidity.BitIndex/4
	return uint184OfSolidity.Value[left:right]
}

type Uint192OfSolidity struct {
	Base
}

func (uint192OfSolidity *Uint192OfSolidity) Decode() string {
	length := len(uint192OfSolidity.Value)
	left := length - (uint192OfSolidity.BitIndex+uint192OfSolidity.BitNum)/4
	right := length - uint192OfSolidity.BitIndex/4
	return uint192OfSolidity.Value[left:right]
}

type Uint200OfSolidity struct {
	Base
}

func (uint200OfSolidity *Uint200OfSolidity) Decode() string {
	length := len(uint200OfSolidity.Value)
	left := length - (uint200OfSolidity.BitIndex+uint200OfSolidity.BitNum)/4
	right := length - uint200OfSolidity.BitIndex/4
	return uint200OfSolidity.Value[left:right]
}

type Uint208OfSolidity struct {
	Base
}

func (uint208OfSolidity *Uint208OfSolidity) Decode() string {
	length := len(uint208OfSolidity.Value)
	left := length - (uint208OfSolidity.BitIndex+uint208OfSolidity.BitNum)/4
	right := length - uint208OfSolidity.BitIndex/4
	return uint208OfSolidity.Value[left:right]
}

type Uint216OfSolidity struct {
	Base
}

func (uint216OfSolidity *Uint216OfSolidity) Decode() string {
	length := len(uint216OfSolidity.Value)
	left := length - (uint216OfSolidity.BitIndex+uint216OfSolidity.BitNum)/4
	right := length - uint216OfSolidity.BitIndex/4
	return uint216OfSolidity.Value[left:right]
}

type Uint224OfSolidity struct {
	Base
}

func (uint224OfSolidity *Uint224OfSolidity) Decode() string {
	length := len(uint224OfSolidity.Value)
	left := length - (uint224OfSolidity.BitIndex+uint224OfSolidity.BitNum)/4
	right := length - uint224OfSolidity.BitIndex/4
	return uint224OfSolidity.Value[left:right]
}

type Uint232OfSolidity struct {
	Base
}

func (uint232OfSolidity *Uint232OfSolidity) Decode() string {
	length := len(uint232OfSolidity.Value)
	left := length - (uint232OfSolidity.BitIndex+uint232OfSolidity.BitNum)/4
	right := length - uint232OfSolidity.BitIndex/4
	return uint232OfSolidity.Value[left:right]
}

type Uint240OfSolidity struct {
	Base
}

func (uint240OfSolidity *Uint240OfSolidity) Decode() string {
	length := len(uint240OfSolidity.Value)
	left := length - (uint240OfSolidity.BitIndex+uint240OfSolidity.BitNum)/4
	right := length - uint240OfSolidity.BitIndex/4
	return uint240OfSolidity.Value[left:right]
}

type Uint248OfSolidity struct {
	Base
}

func (uint248OfSolidity *Uint248OfSolidity) Decode() string {
	length := len(uint248OfSolidity.Value)
	left := length - (uint248OfSolidity.BitIndex+uint248OfSolidity.BitNum)/4
	right := length - uint248OfSolidity.BitIndex/4
	return uint248OfSolidity.Value[left:right]
}

type Uint256OfSolidity struct {
	Base
}

func (uint256OfSolidity *Uint256OfSolidity) Decode() string {
	length := len(uint256OfSolidity.Value)
	left := length - (uint256OfSolidity.BitIndex+uint256OfSolidity.BitNum)/4
	right := length - uint256OfSolidity.BitIndex/4
	return uint256OfSolidity.Value[left:right]
}
