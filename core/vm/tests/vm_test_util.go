package tests

import (
	"fmt"
	"math/big"
	"hyperchain/common"
	"hyperchain/core/state"
	"hyperchain/core/vm"
	"hyperchain/core/vm/api"
	"hyperchain/hyperdb"
	"github.com/op/go-logging"
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
	statedb, _ := state.New(db)
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
	var (
		to    = common.HexToAddress(exec["address"])
		from  = common.HexToAddress(exec["caller"])
		data  = common.FromHex(exec["code"])
		gas = common.Big(exec["gasLimit"])
		price = common.Big(exec["gasPrice"])
		value = common.Big(exec["value"])
		data2 = common.FromHex(exec["data"])
	)

	vmenv := api.GetVMEnv()
	state = vmenv.State()
	ret,err := api.Exec(&from,nil,([]byte)(sourcecode), gas, price, value)
	//ret,err = api.Exec(&from,nil, data, gas, price, value)
	//ret,err = api.Exec(&from,nil, data, gas, price, value)
	//ret,err = api.Exec(&from,nil, data, gas, price, value)
	//ret,err = api.Exec(&from,nil, data, gas, price, value)
	//ret,err = api.Exec(&from,nil, data, gas, price, value)

	for k,v := range state.GetAccounts(){
		fmt.Println("k:",k,"---------,v:",v.Balance())
	}

	addr := state.GetLeastAccount().Address()
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"----",v)
	}
	fmt.Println("addr---------------",addr)
	fmt.Println("ret--------",ret)

	ret,err = api.Exec(&from, &to, data, gas, price, value)
	fmt.Println("ret--------",ret)
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"++++++",v)
	}

	to = addr
	ret,err = api.Exec(&from, &to, data2, gas, price, value)
	fmt.Println("ret--------",ret)
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"++++++",v)
	}

	ret,err = api.Exec(&from, &to, data2, gas, price, value)
	fmt.Println("ret--------",ret)
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"++++++",v)
	}

	ret,err = api.Exec(&from, &to, data2, gas, price, value)
	fmt.Println("ret--------",ret)
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"++++++",v)
	}
	if err != nil{
		log.Error("VM call err:",err)
	}

	return ret, vmenv.State().Logs(), vmenv.Gas, err
}