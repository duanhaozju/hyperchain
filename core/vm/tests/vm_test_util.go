package tests

import (
	"math/big"
	"hyperchain/common"
	"hyperchain/core/state"
	"hyperchain/core"
	"hyperchain/core/vm"
	"hyperchain/hyperdb"
	"github.com/op/go-logging"
	"fmt"
	"time"
	"hyperchain/core/types"
)
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p")
}
var sourcecode = `
contract mortal {
     /* Define variable owner of the type address*/
     address owner;

     /* this function is executed at initialization and sets the owner of the contract */
     function mortal() {
         owner = msg.sender;
     }

     /* Function to recover the funds on the contract */
     function kill() {
         if (msg.sender == owner)
             selfdestruct(owner);
     }
 }


 contract greeter is mortal {
     /* define variable greeting of the type string */
     string greeting;
    uint32 sum;
     /* this runs when the contract is executed */
     function greeter(string _greeting) public {
         greeting = _greeting;
     }

     /* main function */
     function greet() constant returns (string) {
         return greeting;
     }
    function add(uint32 num1,uint32 num2) {
        sum = sum+num1+num2;
    }
 }`
type bconf struct {
	name    string
	precomp bool
	jit     bool
}

func RunVmTest(p string, skipTests []string) error {
	tests := make(map[string]VmTest)
	err := readJsonFile(p, &tests)
	if err != nil {
		return err
	}

	if err := runVmTests(tests, skipTests); err != nil {
		return err
	}

	return nil
}

func runVmTests(tests map[string]VmTest, skipTests []string) error {
	skipTest := make(map[string]bool, len(skipTests))
	for _, name := range skipTests {
		skipTest[name] = true
	}

	for name, test := range tests {
		if err := runVmTest(test); err != nil {
			return fmt.Errorf("%s %s", name, err.Error())
		}
		//fmt.Println("tests----",test.Exec)
		return nil
	}
	return nil
}

func runVmTest(test VmTest) error {
	db, _ := hyperdb.NewMemDatabase()
	statedb, _ := state.New(common.Hash{},db)
	for addr, account := range test.Pre {
		obj := StateObjectFromAccount(db, addr, account)
		statedb.SetStateObject(obj)
		for a, v := range account.Storage {
			obj.SetState(common.HexToHash(a), common.HexToHash(v))
		}
	}
	RunVm(statedb, test.Exec)
	return nil
}

func RunVm(state *state.StateDB, exec map[string]string) ([]byte, vm.Logs, *big.Int, error) {
	// init the parameters
	var (
		addr common.Address
		receipt *types.Receipt
		err error
		ret []byte
		testNum = 1
	)
	//vm.Precompiled = vm.PrecompiledContracts()
	vmenv := core.GetVMEnv()
	//core.InitTestEnv()
	state = vmenv.State()
	state.CreateAccount(common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec3"))
	state.AddBalance(common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec3"),big.NewInt(100000))

	// create a new contract
	log.Debug("create the contract--------------------------")
	now_time := time.Now()
	for i := 0;i<testNum;i++{
		receipt,ret,addr,err =core.ExecTransaction(*types.NewTestCreateTransaction(),vmenv)
	}
	log.Debug("the create contract time we used is ",time.Now().Sub(now_time))
	log.Debug("----------addr",common.ToHex(addr.Bytes()))
	log.Debug("the nonce of account is ",state.GetNonce(common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec6")))
	log.Debug("receipt",receipt.Ret)

	log.Debug("create the contract--------------------------")
	now_time = time.Now()
	for i := 0;i<testNum;i++{
		tx := types.NewTestCallTransaction()
		tx.To = addr.Bytes()
		receipt,ret,_,_ = core.ExecTransaction(*tx,vmenv)
	}
	//state.GetAccount(addr).PrintStorages()
	log.Debug("the call contract time we used is ",time.Now().Sub(now_time))

	return ret, vmenv.State().Logs(), vmenv.Gas, err
}