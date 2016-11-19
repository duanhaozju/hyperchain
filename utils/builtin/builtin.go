package builtin

import (
	"github.com/op/go-logging"
	"hyperchain/crypto"
	"math/rand"
	"time"
	"hyperchain/accounts"
)

var (
	logger         *logging.Logger // package-level logger
	encryption     = crypto.NewEcdsaEncrypto("ecdsa")
	kec256Hash     = crypto.NewKeccak256Hash("keccak256")
	normalTxPool   []string
	contractTxPool []string
	simulateTxPool []string
	genesisAccount []string
	contract       []string
	globalNodes    []string
	am             *accounts.AccountManager
	globalAccounts []string
	NHcontract     string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	logger = logging.MustGetLogger("builtin")
	am = accounts.NewAccountManager(keystore, encryption)
	am.UnlockAllAccount(keystore)
	_, globalAccounts = read(accountList, 5000)
}

const (
	keystore           = "./keystore"
	accountList        = "./keystore/addresses/address"
	normalTxStore      = "./testdata/normal_tx"
	contractTxStore    = "./testdata/contract_tx"
	contractStore      = "./testdata/contract"
	simulateStore      = "./testdata/simulate"
	transferUpperLimit = 100
	transferLowerLimit = 0
	defaultGas         = 10000
	defaultGasPrice    = 10000
	timestampRange     = 10000000000
	genesisPassword    = "123"

	normalTransactionNumber   = 1
	contractTransactionNumber = 1
	contractNumber            = 1
	simulateNumber            = 1
	accountNumber             = 50000
	payload                   = "0x60606040526000805463ffffffff1916815560ae908190601e90396000f3606060405260e060020a60003504633ad14af381146030578063569c5f6d14605e578063d09de08a146084575b6002565b346002576000805460e060020a60243560043563ffffffff8416010181020463ffffffff199091161790555b005b3460025760005463ffffffff166040805163ffffffff9092168252519081900360200190f35b34600257605c6000805460e060020a63ffffffff821660010181020463ffffffff1990911617905556"
	methodid                  = "0xd09de08a"
)
