// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package evm

import (
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"math/big"
)

var bigMaxUint64 = new(big.Int).SetUint64(^uint64(0))

// destinations stores one map per contract (keyed by hash of code).
// The maps contain an entry for each location of a JUMPDEST
// instruction.
type destinations map[common.Hash]map[uint64]struct{}

// has checks whether code has a JUMPDEST at dest.
func (d destinations) has(codehash common.Hash, code []byte, dest *big.Int) bool {
	// PC cannot go beyond len(code) and certainly can't be bigger than 64bits.
	// Don't bother checking for JUMPDEST in that case.
	if dest.Cmp(bigMaxUint64) > 0 {
		return false
	}
	m, analysed := d[codehash]
	if !analysed {
		m = jumpdests(code)
		d[codehash] = m
	}
	_, ok := m[dest.Uint64()]
	return ok
}

// jumpdests creates a map that contains an entry for each
// PC location that is a JUMPDEST instruction.
func jumpdests(code []byte) map[uint64]struct{} {
	m := make(map[uint64]struct{})
	for pc := uint64(0); pc < uint64(len(code)); pc++ {
		var op OpCode = OpCode(code[pc])
		switch op {
		case PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8, PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16, PUSH17, PUSH18, PUSH19, PUSH20, PUSH21, PUSH22, PUSH23, PUSH24, PUSH25, PUSH26, PUSH27, PUSH28, PUSH29, PUSH30, PUSH31, PUSH32:
			a := uint64(op) - uint64(PUSH1) + 1
			pc += a
		case JUMPDEST:
			m[pc] = struct{}{}
		}
	}
	return m
}

// Disassemble disassembles the byte code and returns the string
// representation (human readable opcodes).
func Disassemble(script []byte) (asm []string) {
	pc := new(big.Int)
	for {
		if pc.Cmp(big.NewInt(int64(len(script)))) >= 0 {
			return
		}

		// Get the memory location of pc
		val := script[pc.Int64()]
		// Get the opcode (it must be an opcode!)
		op := OpCode(val)

		asm = append(asm, fmt.Sprintf("%v", op))

		switch op {
		case PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8, PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16, PUSH17, PUSH18, PUSH19, PUSH20, PUSH21, PUSH22, PUSH23, PUSH24, PUSH25, PUSH26, PUSH27, PUSH28, PUSH29, PUSH30, PUSH31, PUSH32:
			pc.Add(pc, common.Big1)
			a := int64(op) - int64(PUSH1) + 1
			if int(pc.Int64()+a) > len(script) {
				return nil
			}

			data := script[pc.Int64() : pc.Int64()+a]
			if len(data) == 0 {
				data = []byte{0}
			}
			asm = append(asm, fmt.Sprintf("0x%x", data))

			pc.Add(pc, big.NewInt(a-1))
		}

		pc.Add(pc, common.Big1)
	}
}

// Parse parses all opcodes from the given code byte slice. This function
// performs no error checking and may return non-existing opcodes.
func Parse(code []byte) (opcodes []OpCode) {
	for pc := uint64(0); pc < uint64(len(code)); pc++ {
		op := OpCode(code[pc])

		switch op {
		case PUSH1, PUSH2, PUSH3, PUSH4, PUSH5, PUSH6, PUSH7, PUSH8, PUSH9, PUSH10, PUSH11, PUSH12, PUSH13, PUSH14, PUSH15, PUSH16, PUSH17, PUSH18, PUSH19, PUSH20, PUSH21, PUSH22, PUSH23, PUSH24, PUSH25, PUSH26, PUSH27, PUSH28, PUSH29, PUSH30, PUSH31, PUSH32:
			a := uint64(op) - uint64(PUSH1) + 1
			pc += a
			opcodes = append(opcodes, PUSH)
		case DUP1, DUP2, DUP3, DUP4, DUP5, DUP6, DUP7, DUP8, DUP9, DUP10, DUP11, DUP12, DUP13, DUP14, DUP15, DUP16:
			opcodes = append(opcodes, DUP)
		case SWAP1, SWAP2, SWAP3, SWAP4, SWAP5, SWAP6, SWAP7, SWAP8, SWAP9, SWAP10, SWAP11, SWAP12, SWAP13, SWAP14, SWAP15, SWAP16:
			opcodes = append(opcodes, SWAP)
		default:
			opcodes = append(opcodes, op)
		}
	}

	return opcodes
}

// MatchFn searcher for match in the given input and calls matcheFn if it finds
// an appropriate match. matcherFn yields the starting position in the input.
// MatchFn will continue to search for a match until it reaches the end of the
// buffer or if matcherFn return false.
func MatchFn(input, match []OpCode, matcherFn func(int) bool) {
	// short circuit if either input or match is empty or if the match is
	// greater than the input
	if len(input) == 0 || len(match) == 0 || len(match) > len(input) {
		return
	}

main:
	for i, op := range input[:len(input)+1-len(match)] {
		// match first opcode and continue search
		if op == match[0] {
			for j := 1; j < len(match); j++ {
				if input[i+j] != match[j] {
					continue main
				}
			}
			// check for abort instruction
			if !matcherFn(i) {
				return
			}
		}
	}
}
