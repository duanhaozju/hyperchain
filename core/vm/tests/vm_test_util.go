package tests

import (
	"fmt"
	"math/big"
	"strconv"

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
	env := make(map[string]string)
	env["currentGasLimit"] = test.Env.CurrentGasLimit
	env["currentNumber"] = test.Env.CurrentNumber
	if n, ok := test.Env.CurrentTimestamp.(float64); ok {
		env["currentTimestamp"] = strconv.Itoa(int(n))
	} else {
		env["currentTimestamp"] = test.Env.CurrentTimestamp.(string)
	}
	RunVm(statedb, env, test.Exec)
	return nil
}

func RunVm(state *state.StateDB, env, exec map[string]string) ([]byte, vm.Logs, *big.Int, error) {
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
	ret,err := api.Exec(vmenv,&from,nil, data, gas, price, value)
	addr := state.GetLeastAccount().Address()
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"----",v)
	}
	fmt.Println("addr---------------",addr)
	fmt.Println("ret--------",ret)

	ret,err = api.Exec(vmenv,&from, &to, data, gas, price, value)
	fmt.Println("ret--------",ret)
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"++++++",v)
	}

	to = addr
	ret,err = api.Exec(vmenv,&from, &to, data2, gas, price, value)
	fmt.Println("ret--------",ret)
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"++++++",v)
	}

	ret,err = api.Exec(vmenv,&from, &to, data2, gas, price, value)
	fmt.Println("ret--------",ret)
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"++++++",v)
	}

	ret,err = api.Exec(vmenv,&from, &to, data2, gas, price, value)
	fmt.Println("ret--------",ret)
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"++++++",v)
	}
	if err != nil{
		log.Error("VM call err:",err)
	}

	return ret, vmenv.State().Logs(), vmenv.Gas, err
}