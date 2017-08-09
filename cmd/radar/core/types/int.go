package types

type Int8OfSolidity struct {
	Base
}
func (int8OfSolidity *Int8OfSolidity) Decode() string {
	length := len(int8OfSolidity.Value)
	left := length - (int8OfSolidity.BitIndex+ int8OfSolidity.BitNum)/4
	right := length - int8OfSolidity.BitIndex / 4
	return int8OfSolidity.Value[left:right]
}

type Int16OfSolidity struct {
	Base
}
func (int16OfSolidity *Int16OfSolidity) Decode() string {
	length := len(int16OfSolidity.Value)
	left := length - (int16OfSolidity.BitIndex+ int16OfSolidity.BitNum)/4
	right := length - int16OfSolidity.BitIndex / 4
	return int16OfSolidity.Value[left:right]
}

type Int24OfSolidity struct {
	Base
}
func (int24OfSolidity *Int24OfSolidity) Decode() string {
	length := len(int24OfSolidity.Value)
	left := length - (int24OfSolidity.BitIndex+ int24OfSolidity.BitNum)/4
	right := length - int24OfSolidity.BitIndex / 4
	return int24OfSolidity.Value[left:right]
}

type Int32OfSolidity struct {
	Base
}
func (int32OfSolidity *Int32OfSolidity) Decode() string {
	length := len(int32OfSolidity.Value)
	left := length - (int32OfSolidity.BitIndex+ int32OfSolidity.BitNum)/4
	right := length - int32OfSolidity.BitIndex / 4
	return int32OfSolidity.Value[left:right]
}

type Int40OfSolidity struct {
	Base
}
func (int40OfSolidity *Int40OfSolidity) Decode() string {
	length := len(int40OfSolidity.Value)
	left := length - (int40OfSolidity.BitIndex+ int40OfSolidity.BitNum)/4
	right := length - int40OfSolidity.BitIndex / 4
	return int40OfSolidity.Value[left:right]
}

type Int48OfSolidity struct {
	Base
}
func (int48OfSolidity *Int48OfSolidity) Decode() string {
	length := len(int48OfSolidity.Value)
	left := length - (int48OfSolidity.BitIndex+ int48OfSolidity.BitNum)/4
	right := length - int48OfSolidity.BitIndex / 4
	return int48OfSolidity.Value[left:right]
}

type Int56OfSolidity struct {
	Base
}
func (int56OfSolidity *Int56OfSolidity) Decode() string {
	length := len(int56OfSolidity.Value)
	left := length - (int56OfSolidity.BitIndex+ int56OfSolidity.BitNum)/4
	right := length - int56OfSolidity.BitIndex / 4
	return int56OfSolidity.Value[left:right]
}

type Int64OfSolidity struct {
	Base
}
func (int64OfSolidity *Int64OfSolidity) Decode() string {
	length := len(int64OfSolidity.Value)
	left := length - (int64OfSolidity.BitIndex+ int64OfSolidity.BitNum)/4
	right := length - int64OfSolidity.BitIndex / 4
	return int64OfSolidity.Value[left:right]
}

type Int72OfSolidity struct {
	Base
}
func (int72OfSolidity *Int72OfSolidity) Decode() string {
	length := len(int72OfSolidity.Value)
	left := length - (int72OfSolidity.BitIndex+ int72OfSolidity.BitNum)/4
	right := length - int72OfSolidity.BitIndex / 4
	return int72OfSolidity.Value[left:right]
}

type Int80OfSolidity struct {
	Base
}
func (int80OfSolidity *Int80OfSolidity) Decode() string {
	length := len(int80OfSolidity.Value)
	left := length - (int80OfSolidity.BitIndex+ int80OfSolidity.BitNum)/4
	right := length - int80OfSolidity.BitIndex / 4
	return int80OfSolidity.Value[left:right]
}

type Int88OfSolidity struct {
	Base
}
func (int88OfSolidity *Int88OfSolidity) Decode() string {
	length := len(int88OfSolidity.Value)
	left := length - (int88OfSolidity.BitIndex+ int88OfSolidity.BitNum)/4
	right := length - int88OfSolidity.BitIndex / 4
	return int88OfSolidity.Value[left:right]
}

type Int96OfSolidity struct {
	Base
}
func (int96OfSolidity *Int96OfSolidity) Decode() string {
	length := len(int96OfSolidity.Value)
	left := length - (int96OfSolidity.BitIndex+ int96OfSolidity.BitNum)/4
	right := length - int96OfSolidity.BitIndex / 4
	return int96OfSolidity.Value[left:right]
}

type Int104OfSolidity struct {
	Base
}
func (int104OfSolidity *Int104OfSolidity) Decode() string {
	length := len(int104OfSolidity.Value)
	left := length - (int104OfSolidity.BitIndex+ int104OfSolidity.BitNum)/4
	right := length - int104OfSolidity.BitIndex / 4
	return int104OfSolidity.Value[left:right]
}

type Int112OfSolidity struct {
	Base
}
func (int112OfSolidity *Int112OfSolidity) Decode() string {
	length := len(int112OfSolidity.Value)
	left := length - (int112OfSolidity.BitIndex+ int112OfSolidity.BitNum)/4
	right := length - int112OfSolidity.BitIndex / 4
	return int112OfSolidity.Value[left:right]
}

type Int120OfSolidity struct {
	Base
}
func (int120OfSolidity *Int120OfSolidity) Decode() string {
	length := len(int120OfSolidity.Value)
	left := length - (int120OfSolidity.BitIndex+ int120OfSolidity.BitNum)/4
	right := length - int120OfSolidity.BitIndex / 4
	return int120OfSolidity.Value[left:right]
}

type Int128OfSolidity struct {
	Base
}
func (int128OfSolidity *Int128OfSolidity) Decode() string {
	length := len(int128OfSolidity.Value)
	left := length - (int128OfSolidity.BitIndex+ int128OfSolidity.BitNum)/4
	right := length - int128OfSolidity.BitIndex / 4
	return int128OfSolidity.Value[left:right]
}

type Int136OfSolidity struct {
	Base
}
func (int136OfSolidity *Int136OfSolidity) Decode() string {
	length := len(int136OfSolidity.Value)
	left := length - (int136OfSolidity.BitIndex+ int136OfSolidity.BitNum)/4
	right := length - int136OfSolidity.BitIndex / 4
	return int136OfSolidity.Value[left:right]
}

type Int144OfSolidity struct {
	Base
}
func (int144OfSolidity *Int144OfSolidity) Decode() string {
	length := len(int144OfSolidity.Value)
	left := length - (int144OfSolidity.BitIndex+ int144OfSolidity.BitNum)/4
	right := length - int144OfSolidity.BitIndex / 4
	return int144OfSolidity.Value[left:right]
}

type Int152OfSolidity struct {
	Base
}
func (int152OfSolidity *Int152OfSolidity) Decode() string {
	length := len(int152OfSolidity.Value)
	left := length - (int152OfSolidity.BitIndex+ int152OfSolidity.BitNum)/4
	right := length - int152OfSolidity.BitIndex / 4
	return int152OfSolidity.Value[left:right]
}

type Int160OfSolidity struct {
	Base
}
func (int160OfSolidity *Int160OfSolidity) Decode() string {
	length := len(int160OfSolidity.Value)
	left := length - (int160OfSolidity.BitIndex+ int160OfSolidity.BitNum)/4
	right := length - int160OfSolidity.BitIndex / 4
	return int160OfSolidity.Value[left:right]
}

type Int168OfSolidity struct {
	Base
}
func (int168OfSolidity *Int168OfSolidity) Decode() string {
	length := len(int168OfSolidity.Value)
	left := length - (int168OfSolidity.BitIndex+ int168OfSolidity.BitNum)/4
	right := length - int168OfSolidity.BitIndex / 4
	return int168OfSolidity.Value[left:right]
}

type Int176OfSolidity struct {
	Base
}
func (int176OfSolidity *Int176OfSolidity) Decode() string {
	length := len(int176OfSolidity.Value)
	left := length - (int176OfSolidity.BitIndex + int176OfSolidity.BitNum) / 4
	right := length - int176OfSolidity.BitIndex / 4
	return int176OfSolidity.Value[left:right]
}

type Int184OfSolidity struct {
	Base
}
func (int184OfSolidity *Int184OfSolidity) Decode() string {
	length := len(int184OfSolidity.Value)
	left := length - (int184OfSolidity.BitIndex+ int184OfSolidity.BitNum)/4
	right := length - int184OfSolidity.BitIndex / 4
	return int184OfSolidity.Value[left:right]
}

type Int192OfSolidity struct {
	Base
}
func (int192OfSolidity *Int192OfSolidity) Decode() string {
	length := len(int192OfSolidity.Value)
	left := length - (int192OfSolidity.BitIndex+ int192OfSolidity.BitNum)/4
	right := length - int192OfSolidity.BitIndex / 4
	return int192OfSolidity.Value[left:right]
}

type Int200OfSolidity struct {
	Base
}
func (int200OfSolidity *Int200OfSolidity) Decode() string {
	length := len(int200OfSolidity.Value)
	left := length - (int200OfSolidity.BitIndex+ int200OfSolidity.BitNum)/4
	right := length - int200OfSolidity.BitIndex / 4
	return int200OfSolidity.Value[left:right]
}

type Int208OfSolidity struct {
	Base
}
func (int208OfSolidity *Int208OfSolidity) Decode() string {
	length := len(int208OfSolidity.Value)
	left := length - (int208OfSolidity.BitIndex+ int208OfSolidity.BitNum)/4
	right := length - int208OfSolidity.BitIndex / 4
	return int208OfSolidity.Value[left:right]
}

type Int216OfSolidity struct {
	Base
}
func (int216OfSolidity *Int216OfSolidity) Decode() string {
	length := len(int216OfSolidity.Value)
	left := length - (int216OfSolidity.BitIndex+ int216OfSolidity.BitNum)/4
	right := length - int216OfSolidity.BitIndex / 4
	return int216OfSolidity.Value[left:right]
}

type Int224OfSolidity struct {
	Base
}
func (int224OfSolidity *Int224OfSolidity) Decode() string {
	length := len(int224OfSolidity.Value)
	left := length - (int224OfSolidity.BitIndex+ int224OfSolidity.BitNum)/4
	right := length - int224OfSolidity.BitIndex / 4
	return int224OfSolidity.Value[left:right]
}

type Int232OfSolidity struct {
	Base
}
func (int232OfSolidity *Int232OfSolidity) Decode() string {
	length := len(int232OfSolidity.Value)
	left := length - (int232OfSolidity.BitIndex+ int232OfSolidity.BitNum)/4
	right := length - int232OfSolidity.BitIndex / 4
	return int232OfSolidity.Value[left:right]
}

type Int240OfSolidity struct {
	Base
}
func (int240OfSolidity *Int240OfSolidity) Decode() string {
	length := len(int240OfSolidity.Value)
	left := length - (int240OfSolidity.BitIndex+ int240OfSolidity.BitNum)/4
	right := length - int240OfSolidity.BitIndex / 4
	return int240OfSolidity.Value[left:right]
}

type Int248OfSolidity struct {
	Base
}
func (int248OfSolidity *Int248OfSolidity) Decode() string {
	length := len(int248OfSolidity.Value)
	left := length - (int248OfSolidity.BitIndex+ int248OfSolidity.BitNum)/4
	right := length - int248OfSolidity.BitIndex / 4
	return int248OfSolidity.Value[left:right]
}

type Int256OfSolidity struct {
	Base
}
func (int256OfSolidity *Int256OfSolidity) Decode() string {
	length := len(int256OfSolidity.Value)
	left := length - (int256OfSolidity.BitIndex+ int256OfSolidity.BitNum)/4
	right := length - int256OfSolidity.BitIndex / 4
	return int256OfSolidity.Value[left:right]
}