package runtime

import (
	"math/big"
	"testing"

	"hyperchain/common"
	"hyperchain/core/state"
	"hyperchain/core/evm"
	"hyperchain/hyperdb"
)

func TestDefaults(t *testing.T) {
	cfg := new(Config)
	setDefaults(cfg)

	if cfg.Difficulty == nil {
		t.Error("expected difficulty to be non nil")
	}

	if cfg.Time == nil {
		t.Error("expected time to be non nil")
	}
	if cfg.GasLimit == nil {
		t.Error("expected time to be non nil")
	}
	if cfg.GasPrice == nil {
		t.Error("expected time to be non nil")
	}
	if cfg.Value == nil {
		t.Error("expected time to be non nil")
	}
	if cfg.GetHashFn == nil {
		t.Error("expected time to be non nil")
	}
	if cfg.BlockNumber == nil {
		t.Error("expected block number to be non nil")
	}
}

func TestEnvironment(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("crashed with: %v", r)
		}
	}()

	Execute([]byte{
		byte(evm.DIFFICULTY),
		byte(evm.TIMESTAMP),
		byte(evm.GASLIMIT),
		byte(evm.PUSH1),
		byte(evm.ORIGIN),
		byte(evm.BLOCKHASH),
		byte(evm.COINBASE),
	}, nil, nil)
}

func TestExecute(t *testing.T) {
	ret, _, err := Execute([]byte{
		byte(evm.PUSH1), 10,
		byte(evm.PUSH1), 0,
		byte(evm.MSTORE),
		byte(evm.PUSH1), 32,
		byte(evm.PUSH1), 0,
		byte(evm.RETURN),
	}, nil, nil)
	if err != nil {
		t.Fatal("didn't expect error", err)
	}

	num := common.BytesToBig(ret)
	if num.Cmp(big.NewInt(10)) != 0 {
		t.Error("Expected 10, got", num)
	}
}

func TestCall(t *testing.T) {
	db, _ := hyperdb.NewMemDatabase()
	state, _ := state.New(common.Hash{}, db)
	address := common.HexToAddress("0x0a")
	state.SetCode(address, []byte{
		byte(evm.PUSH1), 10,
		byte(evm.PUSH1), 0,
		byte(evm.MSTORE),
		byte(evm.PUSH1), 32,
		byte(evm.PUSH1), 0,
		byte(evm.RETURN),
	})

	ret, err := Call(address, nil, &Config{State: state})
	if err != nil {
		t.Fatal("didn't expect error", err)
	}

	num := common.BytesToBig(ret)
	if num.Cmp(big.NewInt(10)) != 0 {
		t.Error("Expected 10, got", num)
	}
}
