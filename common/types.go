//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"math/big"
	"math/rand"
	"reflect"
	"strings"
)

const (
	HashLength    = 32
	AddressLength = 20
)

var hashJsonLengthErr = errors.New("common: unmarshalJSON failed: hash must be exactly 32 bytes")

type (
	Hash    [HashLength]byte
	Address [AddressLength]byte
)

func StringToHex(s string) string {
	if len(s) >= 2 && s[:2] == "0x" {
		return s
	} else {
		return "0x" + s
	}
}

func HexToString(s string) string {
	if len(s) >= 2 && s[:2] == "0x" {
		return s[2:]
	} else {
		return s
	}
}

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}
func StringToHash(s string) Hash { return BytesToHash([]byte(s)) }
func BigToHash(b *big.Int) Hash  { return BytesToHash(b.Bytes()) }
func HexToHash(s string) Hash    { return BytesToHash(FromHex(s)) }

// Don't use the default 'String' method in case we want to overwrite

// Get the string representation of the underlying hash
func (h Hash) Str() string   { return string(h[:]) }
func (h Hash) Bytes() []byte { return h[:] }
func (h Hash) Big() *big.Int { return Bytes2Big(h[:]) }
func (h Hash) Hex() string   { return "0x" + Bytes2Hex(h[:]) }

// UnmarshalJSON parses a hash in its hex from to a hash.
func (h *Hash) UnmarshalJSON(input []byte) error {
	length := len(input)
	if length >= 2 && input[0] == '"' && input[length-1] == '"' {
		input = input[1 : length-1]
	}
	// strip "0x" for length check
	if len(input) > 1 && strings.ToLower(string(input[:2])) == "0x" {
		input = input[2:]
	}

	// validate the length of the input hash
	if len(input) != HashLength*2 {
		return hashJsonLengthErr
	}
	h.SetBytes(FromHex(string(input)))
	return nil
}

// Serialize given hash to JSON
func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Hex())
}

// Sets the hash to the value of b. If b is larger than len(h) it will panic
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

// Set string `s` to h. If s is larger than len(h) it will panic
func (h *Hash) SetString(s string) { h.SetBytes([]byte(s)) }

// Sets h to other
func (h *Hash) Set(other Hash) {
	for i, v := range other {
		h[i] = v
	}
}

// Generate implements testing/quick.Generator.
func (h Hash) Generate(rand *rand.Rand, size int) reflect.Value {
	m := rand.Intn(len(h))
	for i := len(h) - 1; i > m; i-- {
		h[i] = byte(rand.Uint32())
	}
	return reflect.ValueOf(h)
}

func EmptyHash(h Hash) bool {
	return h == Hash{}
}

func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}
func StringToAddress(s string) Address { return BytesToAddress([]byte(s)) }
func BigToAddress(b *big.Int) Address  { return BytesToAddress(b.Bytes()) }
func HexToAddress(s string) Address    { return BytesToAddress(FromHex(s)) }

// IsHexAddress verifies whether a string can represent a valid hex-encoded
// Ethereum address or not.
func IsHexAddress(s string) bool {
	if len(s) == 2+2*AddressLength && IsHex(s) {
		return true
	}
	if len(s) == 2*AddressLength && IsHex("0x"+s) {
		return true
	}
	return false
}

// Get the string representation of the underlying address
func (a Address) Str() string   { return string(a[:]) }
func (a Address) Bytes() []byte { return a[:] }
func (a Address) Big() *big.Int { return Bytes2Big(a[:]) }
func (a Address) Hash() Hash    { return BytesToHash(a[:]) }
func (a Address) Hex() string   { return "0x" + Bytes2Hex(a[:]) }

// Sets the address to the value of b. If b is larger than len(a) it will panic
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

// Set string `s` to a. If s is larger than len(a) it will panic
func (a *Address) SetString(s string) { a.SetBytes([]byte(s)) }

// Sets a to other
func (a *Address) Set(other Address) {
	for i, v := range other {
		a[i] = v
	}
}

// Serialize given address to JSON
func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Hex())
}

// Parse address from raw json data
func (a *Address) UnmarshalJSON(data []byte) error {
	if len(data) > 2 && data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}

	if len(data) > 2 && data[0] == '0' && data[1] == 'x' {
		data = data[2:]
	}

	if len(data) != 2*AddressLength {
		return fmt.Errorf("Invalid address length, expected %d got %d bytes", 2*AddressLength, len(data))
	}

	n, err := hex.Decode(a[:], data)
	if err != nil {
		return err
	}

	if n != AddressLength {
		return fmt.Errorf("Invalid address")
	}

	a.Set(HexToAddress(string(data)))
	return nil
}

// PP Pretty Prints a byte slice in the following format:
// 	hex(value[:4])...(hex[len(value)-4:])
func PP(value []byte) string {
	if len(value) <= 8 {
		return Bytes2Hex(value)
	}

	return fmt.Sprintf("%x...%x", value[:4], value[len(value)-4])
}

// API describes the set of methods offered over the RPC interface
//type API struct {
//	Srvname string      // srvname under which the rpc methods of Service are exposed
//	Version string      // api version
//	Service interface{} // receiver instance which holds the methods
//	Public  bool        // indication if the methods must be considered safe for public use
//}
//
//var Apis map[string]*API

// rpcRequest represents a raw incoming RPC request
type RPCRequest struct {
	Service   string
	Method    string
	Namespace string
	Id        interface{}
	IsPubSub  bool
	Params    interface{}
	Ctx       context.Context
}

// rpcResponse represents a raw incoming RPC request
type RPCResponse struct {
	Namespace string
	Id        interface{}
	//Reply []reflect.Value
	Reply interface{}
	Error RPCError
	IsPubSub  bool
	IsUnsub   bool
}

type RPCNotification struct {
	Namespace string
	SubId     ID
	Service   string
	Result    interface{}
}