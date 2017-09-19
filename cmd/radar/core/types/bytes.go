package types

type Bytes1OfSolidity struct {
	Base
}

func (bytes1OfSolidity *Bytes1OfSolidity) Decode() string {
	length := len(bytes1OfSolidity.Value)
	left := length - (bytes1OfSolidity.BitIndex+bytes1OfSolidity.BitNum)/4
	right := length - bytes1OfSolidity.BitIndex/4
	return bytes1OfSolidity.Value[left:right]
}

type Bytes2OfSolidity struct {
	Base
}

func (bytes2OfSolidity *Bytes2OfSolidity) Decode() string {
	length := len(bytes2OfSolidity.Value)
	left := length - (bytes2OfSolidity.BitIndex+bytes2OfSolidity.BitNum)/4
	right := length - bytes2OfSolidity.BitIndex/4
	return bytes2OfSolidity.Value[left:right]
}

type Bytes3OfSolidity struct {
	Base
}

func (bytes3OfSolidity *Bytes3OfSolidity) Decode() string {
	length := len(bytes3OfSolidity.Value)
	left := length - (bytes3OfSolidity.BitIndex+bytes3OfSolidity.BitNum)/4
	right := length - bytes3OfSolidity.BitIndex/4
	return bytes3OfSolidity.Value[left:right]
}

type Bytes4OfSolidity struct {
	Base
}

func (bytes4OfSolidity *Bytes4OfSolidity) Decode() string {
	length := len(bytes4OfSolidity.Value)
	left := length - (bytes4OfSolidity.BitIndex+bytes4OfSolidity.BitNum)/4
	right := length - bytes4OfSolidity.BitIndex/4
	return bytes4OfSolidity.Value[left:right]
}

type Bytes5OfSolidity struct {
	Base
}

func (bytes5OfSolidity *Bytes5OfSolidity) Decode() string {
	length := len(bytes5OfSolidity.Value)
	left := length - (bytes5OfSolidity.BitIndex+bytes5OfSolidity.BitNum)/4
	right := length - bytes5OfSolidity.BitIndex/4
	return bytes5OfSolidity.Value[left:right]
}

type Bytes6OfSolidity struct {
	Base
}

func (bytes6OfSolidity *Bytes6OfSolidity) Decode() string {
	length := len(bytes6OfSolidity.Value)
	left := length - (bytes6OfSolidity.BitIndex+bytes6OfSolidity.BitNum)/4
	right := length - bytes6OfSolidity.BitIndex/4
	return bytes6OfSolidity.Value[left:right]
}

type Bytes7OfSolidity struct {
	Base
}

func (bytes7OfSolidity *Bytes7OfSolidity) Decode() string {
	length := len(bytes7OfSolidity.Value)
	left := length - (bytes7OfSolidity.BitIndex+bytes7OfSolidity.BitNum)/4
	right := length - bytes7OfSolidity.BitIndex/4
	return bytes7OfSolidity.Value[left:right]
}

type Bytes8OfSolidity struct {
	Base
}

func (bytes8OfSolidity *Bytes8OfSolidity) Decode() string {
	length := len(bytes8OfSolidity.Value)
	left := length - (bytes8OfSolidity.BitIndex+bytes8OfSolidity.BitNum)/4
	right := length - bytes8OfSolidity.BitIndex/4
	return bytes8OfSolidity.Value[left:right]
}

type Bytes9OfSolidity struct {
	Base
}

func (bytes9OfSolidity *Bytes9OfSolidity) Decode() string {
	length := len(bytes9OfSolidity.Value)
	left := length - (bytes9OfSolidity.BitIndex+bytes9OfSolidity.BitNum)/4
	right := length - bytes9OfSolidity.BitIndex/4
	return bytes9OfSolidity.Value[left:right]
}

type Bytes10OfSolidity struct {
	Base
}

func (bytes10OfSolidity *Bytes10OfSolidity) Decode() string {
	length := len(bytes10OfSolidity.Value)
	left := length - (bytes10OfSolidity.BitIndex+bytes10OfSolidity.BitNum)/4
	right := length - bytes10OfSolidity.BitIndex/4
	return bytes10OfSolidity.Value[left:right]
}

type Bytes11OfSolidity struct {
	Base
}

func (bytes11OfSolidity *Bytes11OfSolidity) Decode() string {
	length := len(bytes11OfSolidity.Value)
	left := length - (bytes11OfSolidity.BitIndex+bytes11OfSolidity.BitNum)/4
	right := length - bytes11OfSolidity.BitIndex/4
	return bytes11OfSolidity.Value[left:right]
}

type Bytes12OfSolidity struct {
	Base
}

func (bytes12OfSolidity *Bytes12OfSolidity) Decode() string {
	length := len(bytes12OfSolidity.Value)
	left := length - (bytes12OfSolidity.BitIndex+bytes12OfSolidity.BitNum)/4
	right := length - bytes12OfSolidity.BitIndex/4
	return bytes12OfSolidity.Value[left:right]
}

type Bytes13OfSolidity struct {
	Base
}

func (bytes13OfSolidity *Bytes13OfSolidity) Decode() string {
	length := len(bytes13OfSolidity.Value)
	left := length - (bytes13OfSolidity.BitIndex+bytes13OfSolidity.BitNum)/4
	right := length - bytes13OfSolidity.BitIndex/4
	return bytes13OfSolidity.Value[left:right]
}

type Bytes14OfSolidity struct {
	Base
}

func (bytes14OfSolidity *Bytes14OfSolidity) Decode() string {
	length := len(bytes14OfSolidity.Value)
	left := length - (bytes14OfSolidity.BitIndex+bytes14OfSolidity.BitNum)/4
	right := length - bytes14OfSolidity.BitIndex/4
	return bytes14OfSolidity.Value[left:right]
}

type Bytes15OfSolidity struct {
	Base
}

func (bytes15OfSolidity *Bytes15OfSolidity) Decode() string {
	length := len(bytes15OfSolidity.Value)
	left := length - (bytes15OfSolidity.BitIndex+bytes15OfSolidity.BitNum)/4
	right := length - bytes15OfSolidity.BitIndex/4
	return bytes15OfSolidity.Value[left:right]
}

type Bytes16OfSolidity struct {
	Base
}

func (bytes16OfSolidity *Bytes16OfSolidity) Decode() string {
	length := len(bytes16OfSolidity.Value)
	left := length - (bytes16OfSolidity.BitIndex+bytes16OfSolidity.BitNum)/4
	right := length - bytes16OfSolidity.BitIndex/4
	return bytes16OfSolidity.Value[left:right]
}

type Bytes17OfSolidity struct {
	Base
}

func (bytes17OfSolidity *Bytes17OfSolidity) Decode() string {
	length := len(bytes17OfSolidity.Value)
	left := length - (bytes17OfSolidity.BitIndex+bytes17OfSolidity.BitNum)/4
	right := length - bytes17OfSolidity.BitIndex/4
	return bytes17OfSolidity.Value[left:right]
}

type Bytes18OfSolidity struct {
	Base
}

func (bytes18OfSolidity *Bytes18OfSolidity) Decode() string {
	length := len(bytes18OfSolidity.Value)
	left := length - (bytes18OfSolidity.BitIndex+bytes18OfSolidity.BitNum)/4
	right := length - bytes18OfSolidity.BitIndex/4
	return bytes18OfSolidity.Value[left:right]
}

type Bytes19OfSolidity struct {
	Base
}

func (bytes19OfSolidity *Bytes19OfSolidity) Decode() string {
	length := len(bytes19OfSolidity.Value)
	left := length - (bytes19OfSolidity.BitIndex+bytes19OfSolidity.BitNum)/4
	right := length - bytes19OfSolidity.BitIndex/4
	return bytes19OfSolidity.Value[left:right]
}

type Bytes20OfSolidity struct {
	Base
}

func (bytes20OfSolidity *Bytes20OfSolidity) Decode() string {
	length := len(bytes20OfSolidity.Value)
	left := length - (bytes20OfSolidity.BitIndex+bytes20OfSolidity.BitNum)/4
	right := length - bytes20OfSolidity.BitIndex/4
	return bytes20OfSolidity.Value[left:right]
}

type Bytes21OfSolidity struct {
	Base
}

func (bytes21OfSolidity *Bytes21OfSolidity) Decode() string {
	length := len(bytes21OfSolidity.Value)
	left := length - (bytes21OfSolidity.BitIndex+bytes21OfSolidity.BitNum)/4
	right := length - bytes21OfSolidity.BitIndex/4
	return bytes21OfSolidity.Value[left:right]
}

type Bytes22OfSolidity struct {
	Base
}

func (bytes22OfSolidity *Bytes22OfSolidity) Decode() string {
	length := len(bytes22OfSolidity.Value)
	left := length - (bytes22OfSolidity.BitIndex+bytes22OfSolidity.BitNum)/4
	right := length - bytes22OfSolidity.BitIndex/4
	return bytes22OfSolidity.Value[left:right]
}

type Bytes23OfSolidity struct {
	Base
}

func (bytes23OfSolidity *Bytes23OfSolidity) Decode() string {
	length := len(bytes23OfSolidity.Value)
	left := length - (bytes23OfSolidity.BitIndex+bytes23OfSolidity.BitNum)/4
	right := length - bytes23OfSolidity.BitIndex/4
	return bytes23OfSolidity.Value[left:right]
}

type Bytes24OfSolidity struct {
	Base
}

func (bytes24OfSolidity *Bytes24OfSolidity) Decode() string {
	length := len(bytes24OfSolidity.Value)
	left := length - (bytes24OfSolidity.BitIndex+bytes24OfSolidity.BitNum)/4
	right := length - bytes24OfSolidity.BitIndex/4
	return bytes24OfSolidity.Value[left:right]
}

type Bytes25OfSolidity struct {
	Base
}

func (bytes25OfSolidity *Bytes25OfSolidity) Decode() string {
	length := len(bytes25OfSolidity.Value)
	left := length - (bytes25OfSolidity.BitIndex+bytes25OfSolidity.BitNum)/4
	right := length - bytes25OfSolidity.BitIndex/4
	return bytes25OfSolidity.Value[left:right]
}

type Bytes26OfSolidity struct {
	Base
}

func (bytes26OfSolidity *Bytes26OfSolidity) Decode() string {
	length := len(bytes26OfSolidity.Value)
	left := length - (bytes26OfSolidity.BitIndex+bytes26OfSolidity.BitNum)/4
	right := length - bytes26OfSolidity.BitIndex/4
	return bytes26OfSolidity.Value[left:right]
}

type Bytes27OfSolidity struct {
	Base
}

func (bytes27OfSolidity *Bytes27OfSolidity) Decode() string {
	length := len(bytes27OfSolidity.Value)
	left := length - (bytes27OfSolidity.BitIndex+bytes27OfSolidity.BitNum)/4
	right := length - bytes27OfSolidity.BitIndex/4
	return bytes27OfSolidity.Value[left:right]
}

type Bytes28OfSolidity struct {
	Base
}

func (bytes28OfSolidity *Bytes28OfSolidity) Decode() string {
	length := len(bytes28OfSolidity.Value)
	left := length - (bytes28OfSolidity.BitIndex+bytes28OfSolidity.BitNum)/4
	right := length - bytes28OfSolidity.BitIndex/4
	return bytes28OfSolidity.Value[left:right]
}

type Bytes29OfSolidity struct {
	Base
}

func (bytes29OfSolidity *Bytes29OfSolidity) Decode() string {
	length := len(bytes29OfSolidity.Value)
	left := length - (bytes29OfSolidity.BitIndex+bytes29OfSolidity.BitNum)/4
	right := length - bytes29OfSolidity.BitIndex/4
	return bytes29OfSolidity.Value[left:right]
}

type Bytes30OfSolidity struct {
	Base
}

func (bytes30OfSolidity *Bytes30OfSolidity) Decode() string {
	length := len(bytes30OfSolidity.Value)
	left := length - (bytes30OfSolidity.BitIndex+bytes30OfSolidity.BitNum)/4
	right := length - bytes30OfSolidity.BitIndex/4
	return bytes30OfSolidity.Value[left:right]
}

type Bytes31OfSolidity struct {
	Base
}

func (bytes31OfSolidity *Bytes31OfSolidity) Decode() string {
	length := len(bytes31OfSolidity.Value)
	left := length - (bytes31OfSolidity.BitIndex+bytes31OfSolidity.BitNum)/4
	right := length - bytes31OfSolidity.BitIndex/4
	return bytes31OfSolidity.Value[left:right]
}

type Bytes32OfSolidity struct {
	Base
}

func (bytes32OfSolidity *Bytes32OfSolidity) Decode() string {
	length := len(bytes32OfSolidity.Value)
	left := length - (bytes32OfSolidity.BitIndex+bytes32OfSolidity.BitNum)/4
	right := length - bytes32OfSolidity.BitIndex/4
	return bytes32OfSolidity.Value[left:right]
}
