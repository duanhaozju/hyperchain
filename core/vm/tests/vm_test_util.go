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
	"hyperchain/core/vm/params"
)
var (
	log *logging.Logger // package-level logger
	env	= make(map[string]string)
	vmenv	= (*core.Env)(nil)
)
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

func RunVm(statedb *state.StateDB, exec map[string]string) ([]byte, vm.Logs, *big.Int, error) {
	// init the parameters
	var (
		addr common.Address
		receipt *types.Receipt
		err error
		ret []byte
		testCreateNum = 0
		testCallNum = 0
		testTransferNum = 0
	)
	log = logging.MustGetLogger("p2p")
	db,_ := hyperdb.GetLDBDatabase()
	statedb,_ = state.New(common.Hash{},db)
	env["currentNumber"] = "1"
	env["currentGasLimit"] = "10000000"
	//vm.Precompiled = vm.PrecompiledContracts()
	vmenv = core.NewEnvFromMap(core.RuleSet{params.MainNetHomesteadBlock,params.MainNetDAOForkBlock,true},statedb,env)
	//core.InitTestEnv()
	statedb.CreateAccount(common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec3"))
	statedb.AddBalance(common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec3"),big.NewInt(100000))

	// create a new contract
	log.Debug("Create the contract-------------------------------------------------------------------------------")
	receipt,ret,addr,err =core.ExecTransaction(*types.NewTestCreateTransaction(),vmenv)
	now_time := time.Now()
	for i := 0;i<testCreateNum;i++{
		receipt,ret,addr,err =core.ExecTransaction(*types.NewTestCreateTransaction(),vmenv)
		log.Debug("Create**********************************",i)
		log.Debug("----------addr",common.ToHex(addr.Bytes()))
		log.Debug("receipt",receipt.Ret)
	}
	log.Infof("the create contract time we used is ",time.Now().Sub(now_time))

	log.Debug("Call the contract-------------------------------------------------------------------------------")
	for i := 0;i<1;i++{
		tx := types.NewTestCallTransaction()
		tx.To = addr.Bytes()
		receipt,ret,addr,err =core.ExecTransaction(*tx,vmenv)
		log.Debug("----------addr",common.ToHex(addr.Bytes()))
		log.Debug("the nonce of account is ",statedb.GetNonce(common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec3")))
		log.Debug("receipt",receipt.Ret)
	}
	log.Debug("the create contract time we used is ",time.Now().Sub(now_time))
	log.Debug("create the contract--------------------------")

	now_time = time.Now()
	for i := 0;i<testCallNum;i++{
		log.Debug("Call**********************************",i)
		tx := types.NewTestCallTransaction()
		tx.To = addr.Bytes()
		//receipt,ret,_,_ = core.ExecTransaction(*tx,*vmenv)
	}
	log.Infof("the call contract time we used is ",time.Now().Sub(now_time))
	log.Debug("receipt",receipt.Ret)

	log.Debug("Transfer the Balance-------------------------------------------------------------------------------")
	now_time = time.Now()
	for i := 0;i<testTransferNum;i++{
		tx := types.NewTestCallTransaction()
		tx.Value = nil
		tx.To = addr.Bytes()
		receipt,ret,addr,err =core.ExecTransaction(*tx,vmenv)
		log.Debug("Transfer**********************************",i)
	}
	log.Infof("the Transfer Balance time we used is ",time.Now().Sub(now_time))
	return ret, statedb.Logs(), nil, err
}