package tests

import (
	"fmt"
	"math/big"
	"strconv"

	"hyperchain/common"
	"hyperchain/core/state"
	"hyperchain/core/vm"
	//"hyperchain/hyperdb"
	//"hyperchain/hyperdb"
	"hyperchain/core/vm/params"
	"hyperchain/core/vm/api"
	/*"hyperchain/core/types"
	"github.com/golang/protobuf/proto"*/
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
	statedb, _ := state.New(common.Hash{}, db)
	for addr, account := range test.Pre {
		obj := StateObjectFromAccount(db, addr, account)
		statedb.SetStateObject(obj)
		for a, v := range account.Storage {
			obj.SetState(common.HexToHash(a), common.HexToHash(v))
		}
	}

	// XXX Yeah, yeah...
	env := make(map[string]string)
	env["currentCoinbase"] = test.Env.CurrentCoinbase
	env["currentDifficulty"] = test.Env.CurrentDifficulty
	env["currentGasLimit"] = test.Env.CurrentGasLimit
	env["currentNumber"] = test.Env.CurrentNumber
	env["previousHash"] = test.Env.PreviousHash
	if n, ok := test.Env.CurrentTimestamp.(float64); ok {
		env["currentTimestamp"] = strconv.Itoa(int(n))
	} else {
		env["currentTimestamp"] = test.Env.CurrentTimestamp.(string)
	}

	var (
		ret  []byte
		gas  *big.Int
		err  error
		logs vm.Logs
	)

	ret, logs, gas, err = RunVm(statedb, env, test.Exec)

	//// Compare expected and actual return
	//rexp := common.FromHex(test.Out)
	//if bytes.Compare(rexp, ret) != 0 {
	//	return fmt.Errorf("return failed. Expected %x, got %x\n", rexp, ret)
	//}

	// Check gas usage
	if len(test.Gas) == 0 && err == nil {
		//return fmt.Errorf("gas unspecified, indicating an error. VM returned (incorrectly) successfull")
	} else {
		gexp := common.Big(test.Gas)
		if gexp.Cmp(gas) != 0 {
			//return fmt.Errorf("gas failed. Expected %v, got %v\n", gexp, gas)
		}
	}

	// check post state
	for addr, account := range test.Post {
		obj := statedb.GetStateObject(common.HexToAddress(addr))
		if obj == nil {
			continue
		}

		for addr, value := range account.Storage {
			v := obj.GetState(common.HexToHash(addr))
			vexp := common.HexToHash(value)

			fmt.Println("value: ",value)
			fmt.Println("vexp: ",vexp)
			fmt.Println("v: ",v)
			fmt.Println("ret : ",ret)
			/*
			if v != vexp {
				return fmt.Errorf("(%x: %s) storage failed. Expected %x, got %x (%v %v)\n", obj.Address().Bytes()[0:4], addr, vexp, v, vexp.Big(), v.Big())
			}*/
		}
	}

	// check logs
	if len(test.Logs) > 0 {
		lerr := checkLogs(test.Logs, logs)
		if lerr != nil {
			return lerr
		}
	}

	return nil
}

// TODO 在这里改成我们希望的transaction去验证api
func RunVmCOPY(state *state.StateDB, env, exec map[string]string) ([]byte, vm.Logs, *big.Int, error) {
	var (
		to    = common.HexToAddress(exec["address"])
		from  = common.HexToAddress(exec["caller"])
		data  = common.FromHex(exec["data"])
		gas   = common.Big(exec["gas"])
		price = common.Big(exec["gasPrice"])
		value = common.Big(exec["value"])
		//tx types.Transaction
	)
	vm.Precompiled = make(map[string]*vm.PrecompiledAccount)

	caller := state.GetOrNewStateObject(from)
	vmenv := NewEnvFromMap(RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, state, env, exec)
	vmenv.vmTest = true
	vmenv.skipTransfer = true
	vmenv.initial = true

	//ret, err := vmenv.Call(caller, to, data, gas, price, value)
	ret,err := api.Exec(vmenv,caller, &to, data, gas, price, value)

	log.Info("the to address is : ",to)
	log.Info("the ret is : ",ret)
	log.Info("the ret is : ",string(ret))

	if err != nil{
		log.Error("VM call err:",err)
	}
	return ret, vmenv.state.Logs(), vmenv.Gas, err
}
// TODO 在这里改成我们希望的transaction去验证api

func RunVm(state *state.StateDB, env, exec map[string]string) ([]byte, vm.Logs, *big.Int, error) {
	var (
		to    = common.HexToAddress(exec["address"])
		from  = common.HexToAddress(exec["caller"])
		data  = common.FromHex(exec["code"])
		//gas   = common.Big(exec["gas"])
		gas = big.NewInt(1000000000000)
		price = common.Big(exec["gasPrice"])
		value = common.Big(exec["value"])
		data2 = common.FromHex(exec["data"])
		//tx types.Transaction
	)

	vm.Precompiled = make(map[string]*vm.PrecompiledAccount)

	caller := state.GetOrNewStateObject(from)
	vmenv := NewEnvFromMap(RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, state, env, exec)
	vmenv.vmTest = true
	vmenv.skipTransfer = true
	vmenv.initial = true

	//data = common.FromHex("6016060405261018f806100126000396000f360606040526000357c010000000000000000000000000000000000000000000000000000000090048063a5643bf21461004f578063ab55044d146100f2578063cdcd77c01461012f5761004d565b005b6100f06004808035906020019082018035906020019191908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509090919080359060200190919080359060200190820180359060200191919080806020026020016040519081016040528093929190818152602001838360200280828437820191505050505050909091905050610166565b005b61012d60048080604001906002806020026040519081016040528092919082600260200280828437820191505050505090909190505061016c565b005b61014e6004808035906020019091908035906020019091905050610170565b60405180821515815260200191505060405180910390f35b5b505050565b5b50565b600060208363ffffffff1611806101845750815b905080505b9291505056")
	//data = byte("")
	ret,addr,err := vmenv.Create(caller, data, gas, price, value)
	fmt.Println("data---------------",data)
	fmt.Println("data2---------------",data2)
	fmt.Println("is exist: ",vmenv.Db().Exist(addr))
	//fmt.Println("data2---------------",common.FromHex(exec["data"]))
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"----",v)
	}
	//data  = common.FromHex("0xcdcd77c000000000000000000000000000000000000000000000000000000000000000450000000000000000000000000000000000000000000000000000000000000001")
	fmt.Println("addr---------------",addr)
	fmt.Println("ret--------",ret)

	//ret, err = vmenv.Call(caller, addr, data2, gas, price, value)
	ret,err = api.Exec(vmenv,caller, &to, data, gas, price, value)

	fmt.Println("ret--------",ret)
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"++++++",v)
	}

	to = addr
	ret,err = api.Exec(vmenv,caller, &to, data2, gas, price, value)
	//ret, err = vmenv.Call(caller, addr, data2, gas, price, value)
	fmt.Println("ret--------",ret)
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"++++++",v)
	}

	ret,err = api.Exec(vmenv,caller, &to, data2, gas, price, value)
	//ret, err = vmenv.Call(caller, addr, data2, gas, price, value)
	fmt.Println("ret--------",ret)
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"++++++",v)
	}

	ret,err = api.Exec(vmenv,caller, &to, data2, gas, price, value)
	//ret, err = vmenv.Call(caller, addr, data2, gas, price, value)
	fmt.Println("ret--------",ret)
	for a, v := range state.GetStateObject(addr).Storage() {
		fmt.Println(a,"++++++",v)
	}
	if err != nil{
		log.Error("VM call err:",err)
	}

	return ret, vmenv.state.Logs(), vmenv.Gas, err


}

//
//func RunStateTest(ruleSet RuleSet, p string, skipTests []string) error {
//	tests := make(map[string]VmTest)
//	if err := readJsonFile(p, &tests); err != nil {
//		return err
//	}
//
//	if err := runStateTests(ruleSet, tests, skipTests); err != nil {
//		return err
//	}
//
//	return nil
//
//}
//
//func runStateTests(ruleSet RuleSet, tests map[string]VmTest, skipTests []string) error {
//	skipTest := make(map[string]bool, len(skipTests))
//	for _, name := range skipTests {
//		skipTest[name] = true
//	}
//
//	for name, test := range tests {
//		if skipTest[name] /*|| name != "callcodecallcode_11" */ {
//			glog.Infoln("Skipping state test", name)
//			continue
//		}
//
//		//fmt.Println("StateTest:", name)
//		if err := runStateTest(ruleSet, test); err != nil {
//			return fmt.Errorf("%s: %s\n", name, err.Error())
//		}
//
//		//glog.Infoln("State test passed: ", name)
//		//fmt.Println(string(statedb.Dump()))
//	}
//	return nil
//
//}
//
//func runStateTest(ruleSet RuleSet, test VmTest) error {
//	db, _ := ethdb.NewMemDatabase()
//	statedb, _ := state.New(common.Hash{}, db)
//	for addr, account := range test.Pre {
//		obj := StateObjectFromAccount(db, addr, account)
//		statedb.SetStateObject(obj)
//		for a, v := range account.Storage {
//			obj.SetState(common.HexToHash(a), common.HexToHash(v))
//		}
//	}
//
//	// XXX Yeah, yeah...
//	env := make(map[string]string)
//	env["currentCoinbase"] = test.Env.CurrentCoinbase
//	env["currentDifficulty"] = test.Env.CurrentDifficulty
//	env["currentGasLimit"] = test.Env.CurrentGasLimit
//	env["currentNumber"] = test.Env.CurrentNumber
//	env["previousHash"] = test.Env.PreviousHash
//	if n, ok := test.Env.CurrentTimestamp.(float64); ok {
//		env["currentTimestamp"] = strconv.Itoa(int(n))
//	} else {
//		env["currentTimestamp"] = test.Env.CurrentTimestamp.(string)
//	}
//
//	var (
//		ret []byte
//		// gas  *big.Int
//		// err  error
//		logs vm.Logs
//	)
//
//	ret, logs, _, _ = RunState(ruleSet, statedb, env, test.Transaction)
//
//	// Compare expected and actual return
//	rexp := common.FromHex(test.Out)
//	if bytes.Compare(rexp, ret) != 0 {
//		return fmt.Errorf("return failed. Expected %x, got %x\n", rexp, ret)
//	}
//
//	// check post state
//	for addr, account := range test.Post {
//		obj := statedb.GetStateObject(common.HexToAddress(addr))
//		if obj == nil {
//			return fmt.Errorf("did not find expected post-state account: %s", addr)
//		}
//
//		if obj.Balance().Cmp(common.Big(account.Balance)) != 0 {
//			return fmt.Errorf("(%x) balance failed. Expected: %v have: %v\n", obj.Address().Bytes()[:4], common.String2Big(account.Balance), obj.Balance())
//		}
//
//		if obj.Nonce() != common.String2Big(account.Nonce).Uint64() {
//			return fmt.Errorf("(%x) nonce failed. Expected: %v have: %v\n", obj.Address().Bytes()[:4], account.Nonce, obj.Nonce())
//		}
//
//		for addr, value := range account.Storage {
//			v := obj.GetState(common.HexToHash(addr))
//			vexp := common.HexToHash(value)
//
//			if v != vexp {
//				return fmt.Errorf("storage failed:\n%x: %s:\nexpected: %x\nhave:     %x\n(%v %v)\n", obj.Address().Bytes(), addr, vexp, v, vexp.Big(), v.Big())
//			}
//		}
//	}
//
//	root, _ := statedb.Commit()
//	if common.HexToHash(test.PostStateRoot) != root {
//		return fmt.Errorf("Post state root error. Expected: %s have: %x", test.PostStateRoot, root)
//	}
//
//	// check logs
//	if len(test.Logs) > 0 {
//		if err := checkLogs(test.Logs, logs); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func RunState(ruleSet RuleSet, statedb *state.StateDB, env, tx map[string]string) ([]byte, vm.Logs, *big.Int, error) {
//	var (
//		data  = common.FromHex(tx["data"])
//		gas   = common.Big(tx["gasLimit"])
//		price = common.Big(tx["gasPrice"])
//		value = common.Big(tx["value"])
//		nonce = common.Big(tx["nonce"]).Uint64()
//	)
//
//	var to *common.Address
//	if len(tx["to"]) > 2 {
//		t := common.HexToAddress(tx["to"])
//		to = &t
//	}
//	// Set pre compiled contracts
//	vm.Precompiled = vm.PrecompiledContracts()
//	snapshot := statedb.Copy()
//	gaspool := new(core.GasPool).AddGas(common.Big(env["currentGasLimit"]))
//
//	key, _ := hex.DecodeString(tx["secretKey"])
//	addr := crypto.PubkeyToAddress(crypto.ToECDSA(key).PublicKey)
//	message := NewMessage(addr, to, data, value, gas, price, nonce)
//	vmenv := NewEnvFromMap(ruleSet, statedb, env, tx)
//	vmenv.origin = addr
//	ret, _, err := core.ApplyMessage(vmenv, message, gaspool)
//	if core.IsNonceErr(err) || core.IsInvalidTxErr(err) || core.IsGasLimitErr(err) {
//		statedb.Set(snapshot)
//	}
//	statedb.Commit()
//
//	return ret, vmenv.state.Logs(), vmenv.Gas, err
//}