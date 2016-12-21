//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package vm

import (
	"math/big"
	"testing"
	"hyperchain/hyperdb"
	"hyperchain/common"
	"strconv"
	"fmt"
)

var (
	ForceJit  bool //=true
	EnableJit bool = true
	MainNetHomesteadBlock = big.NewInt(1150000) // Mainnet homestead block
	MainNetDAOForkBlock = big.NewInt(1920000)
	gas = big.NewInt(1000000)
	gasPrice  = big.NewInt(10000)
)
func NewEnv2(ruleSet RuleSet, state *StateDB) *Env {
	env := &Env{
		ruleSet: ruleSet,
		state:   state,
	}
	return env
}
type RuleSetUnitTest struct {
	HomesteadBlock *big.Int
	DAOForkBlock   *big.Int
	DAOForkSupport bool
}

func (r RuleSetUnitTest) IsHomestead(n *big.Int) bool {
	return true
	return n.Cmp(r.HomesteadBlock) >= 0
}
func NewEnvFromMap(ruleSet RuleSet, state *StateDB, envValues map[string]string) *Env {
	env := NewEnv2(ruleSet, state)

	env.time = common.Big(envValues["currentTimestamp"])
	env.gasLimit = common.Big(envValues["currentGasLimit"])
	env.number = common.Big(envValues["currentNumber"])
	env.Gas = new(big.Int)
	env.evm = New(env, Config{
		EnableJit: EnableJit,
		ForceJit:  ForceJit,
	})

	return env
}
func NewTestVmenv() *Env {
	var env = make(map[string]string)
	db, _ := hyperdb.NewMemDatabase()
	statedb, _ := NewStateDB(common.Hash{},db)

	env["currentNumber"] = strconv.FormatUint(uint64(1), 10)
	env["currentGasLimit"] = "10000000"
	vmenv := NewEnvFromMap(RuleSetUnitTest{MainNetHomesteadBlock, MainNetDAOForkBlock, true}, statedb, env)
	return vmenv
}
func TestVmRun(t *testing.T) {

	vmenv := NewTestVmenv()
	evm := vmenv.evm

	from := common.HexToAddress("000f1a7a08ccc48e5d30f80850cf1cf283aa3abd")
	sender := evm.env.Db().CreateAccount(from)
	contractAddr := common.HexToAddress("0xbbe2b6412ccf633222374de8958f2acc76cda9c9")
	to := evm.env.Db().CreateAccount(contractAddr)
	value := big.NewInt(0)
	//haizhen zheshang code
	code :="0x60606040526117e1806100126000396000f36060604052361561008d5760e060020a60003504631d3ac435811461009257806332c1680c1461013657806354f445f1146101825780635d47f91c146101a1578063b5cf8eab14610273578063bf7716ac14610280578063cb7f0472146102fb578063e98e84d414610464578063eb3c667f1461056a578063ed29244f146105e7578063ff86a17914610662575b610002565b34610002576107046004356024356040805160208101909152600080825290818080808080600160e060020a0319891660e460020a6305355303021415610a8b57898152600560205260408120600a0154600160a060020a03908116339091161461097c5760408051808201909152601481527f596f7520646f6e277420686173206163636573730000000000000000000000006020820152600198509650610cf7565b346100025733600160a060020a03166000908152600260208190526040909120602435815560443591810191909155600435600191820155604080519115158252519081900360200190f35b346100025761077c60043560243560443560643560843560a4356107a5565b346100025761081d6004355b60408051602081810183526000808352835180830185528181528451808401865282815285518085018752838152865194850190965282845291949092858080808080805b60008e815260208190526040902054821015610f2157604060009081208f825260209190915280548390811015610002579060005260206000209060050201600050905086805480600101828181548183558181151161112b57601f016020900481601f0160209004836000526020600020918201910161112b9190610ca4565b346100025761095e610980565b34610002576040805160208181018352600080835233600160a060020a031681526003825283902080548451818402810184019095528085526109ae94928301828280156102f057602002820191906000526020600020905b8160005054815260200190600101908083116102d9575b505050505090505b90565b34610002576109f8600435600081815260056020819052604082209081015482918291829182918291829182918291829190600160a060020a039081163391909116148061035a57506006810154600160a060020a0390811633909116145b8061037657506007810154600160a060020a0390811633909116145b8061039257506008810154600160a060020a0390811633909116145b1561045657806000016000505481600101600050548260020160009054906101000a900460e060020a02836003016000505484600401600050548560080160149054906101000a900460c060020a028660050160009054906101000a9004600160a060020a03168760060160009054906101000a9004600160a060020a03168860070160009054906101000a9004600160a060020a03168960080160009054906101000a9004600160a060020a03169a509a509a509a509a509a509a509a509a509a505b509193959799509193959799565b346100025761081d60043560408051602081810183526000808352835180830185528181528451808401865282815285518085018752838152865180860188528481528751808701895285815288518088018a5286815289518089018b528781528a51808a018c528881528b51808b018d528981528d8a5260059a8b90529b8920998a0154989b9799969895979496958c9590600160a060020a039081163391909116148061052457506006870154600160a060020a0390811633909116145b8061054057506007870154600160a060020a0390811633909116145b8061055c57506008870154600160a060020a0390811633909116145b1561111b576112dc8e6101ad565b346100025761070460043560243560443560643560843560a43560c43560e4356101043560408051602081019091526000808252908180888a126112ff5760408051808201909152601a81527f69737375654461746520626967207468656e206475656461746500000000000060208201526001945092506113e2565b34610002576040805160208181018352600080835233600160a060020a031681526001825283902080548451818402810184019095528085526109ae94928301828280156102f057602002820191906000526020600020908160005054815260200190600101908083116102d9575b505050505090506102f8565b3461000257600435600081815260056020526040812060080154610a6d9291602435918190819074010000000000000000000000000000000000000000900460c060020a02600160c060020a0319167f3030303030303200000000000000000000000000000000000000000000000000141561165557600192507f686173206265656e20656e64720000000000000000000000000000000000000091506116cf565b604051808360ff168152602001806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f16801561076d5780820380516001836020036101000a031916815260200191505b50935050505060405180910390f35b005b5050506000928352506020909120018a9055610cbc8a600160e460020a6305355303028989895b600086815260208190526040902080546001810180835590918291828015829011610d8457600502816005028360005260206000209182019101610d8491905b80821115610cb857600080825560018201805464ffffffffff19169055600282018190556003820181905560048201556005016107e5565b6040518087600019168152602001806020018060200180602001806020018060200186810386528b8181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f15090500186810385528a8181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f1509050018681038452898181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f1509050018681038352888181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f1509050018681038252878181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f1509050019b50505050505050505050505060405180910390f35b60408051938452602084019290925282820152519081900360600190f35b610af35b33600160a060020a031660009081526002602081905260409091208054600182015491909201549192909190565b60405180806020018281038252838181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f1509050019250505060405180910390f35b604080519a8b5260208b0199909952600160e060020a03199097168989015260608901959095526080880193909352600160c060020a03199190911660a0870152600160a060020a0390811660c087015290811660e08601529081166101008501521661012083015251908190036101400190f35b6040805160ff93909316835260208301919091528051918290030190f35b600160e060020a0319891660e060020a6353553031021415610cf75760008a8152600560205260409020600801805460a060020a67ffffffffffffffff0219167b30303030303032000000000000000000000000000000000000000000179055610d04610980565b600160a060020a03331660009081526003602052604090209298509096509450610bba908b5b6000805b8354811015610b4c578260001916848281548110156100025760009182526020909120015414156117ca579050805b835482146117db578354849060001981019081101561000257906000526020600020900160005054848381548110156100025760206000908120929052015583546000198101808655908590829080158290116117d6576000838152602090206117d6918101908301610ca4565b60008a815260056020908152604080832060080154600160a060020a0316835260019091529020610beb908b610b19565b60008a81526005602090815260408083206008810180546009909201805473ffffffffffffffffffffffffffffffffffffffff19908116339081179092557b3030303030303300000000000000000000000000000000000000000060a060020a67ffffffffffffffff02199094169390931790921682179055600160a060020a031683526001918290529091208054918201808255909190828183801582901161077e5781836000526020600020918201910161077e91905b80821115610cb85760008155600101610ca4565b5090565b60408051808201909152601481527f656e6472726573706f6e7365207375636365656400000000000000000000000060208201526000985096505b5050505050509250929050565b925092509250610d238a600160e060020a6353553031028686866107a5565b600160a060020a0333166000908152600360205260409020610d45908b610b19565b60408051808201909152601281527f656e64726573706f6e7365206661696c656400000000000000000000000000006020820152600198509650610cf7565b5050506000888152602081905260409020805489925083908110156100025790600052602060002090600502016000505560008781526020819052604090208054879190839081101561000257906000526020600020906005020160005060010160006101000a81548160ff0219169083021790555084600060005060008960001916815260200190815260200160002060005082815481101561000257906000526020600020906005020160005060010160016101000a81548163ffffffff021916908360e060020a9004021790555083600060005060008960001916815260200190815260200160002060005082815481101561000257906000526020600020906005020160005060020160005081905550826000600050600089600019168152602001908152602001600020600050828154811015610002579060005260206000209060050201600050600301600050819055508160006000506000896000191681526020019081526020016000206000508281548110156100025790600052602060002090600502016000506004015550505050505050565b8d878787878784805480602002602001604051908101604052809291908181526020018280548015610f9057602002820191906000526020600020906000905b825461010083900a900460ff16815260206001928301818104948501949093039092029101808411610f615790505b505050505094508380548060200260200160405190810160405280929190818152602001828054801561100857602002820191906000526020600020906000905b82829054906101000a900460e060020a0281526020019060040190602082600301049283019260010382029150808411610fd15790505b505050505093508280548060200260200160405190810160405280929190818152602001828054801561105d57602002820191906000526020600020905b816000505481526020019060010190808311611046575b50505050509250818054806020026020016040519081016040528092919081815260200182805480156110b257602002820191906000526020600020905b81600050548152602001906001019080831161109b575b505050505091508080548060200260200160405190810160405280929190818152602001828054801561110757602002820191906000526020600020905b8160005054815260200190600101908083116110f0575b505050505090509c509c509c509c509c509c505b5050505050505091939550919395565b5050506000928352506020918290206001848101548484049092018054949093066101000a60ff8181021990951692909416909302179055865490810180885587919082818380158290116111a15760070160089004816007016008900483600052602060002091820191016111a19190610ca4565b505050919090600052602060002090600891828204019190066004028360010160019054906101000a900460e060020a02909190916101000a81548163ffffffff021916908360e060020a900402179055505084805480600101828181548183558181151161122157600083815260209020611221918101908301610ca4565b505050919090600052602060002090016000506002830154905550835460018101808655859190828183801582901161126b5760008381526020902061126b918101908301610ca4565b50505091909060005260206000209001600050600383015490555082546001810180855584919082818380158290116112b5576000838152602090206112b5918101908301610ca4565b505050919090600052602060002090016000506004830154905550600191909101906101f2565b9550955095509550955095508585858585859c509c509c509c509c509c5061111b565b7f4143303100000000000000000000000000000000000000000000000000000000600160e060020a03198c161480159061136357507f4143303200000000000000000000000000000000000000000000000000000000600160e060020a03198c1614155b156113f25760408051808201909152601581527f62696c6c74797065206973206e6f74206d61746368000000000000000000000060208201526001945092506113e2565b60408051808201909152601381527f70756242696c6c496e666f20737563636565640000000000000000000000000060208201526000945092505b5050995099975050505050505050565b33600160a060020a031688600160a060020a031614151561144c5760408051808201909152601181527f64727772206973206e6f74206d6174636800000000000000000000000000000060208201526001945092506113e2565b60008d8152600560209081526040808320600160a060020a0389168452600192839052922080549182018082559294509182818380158290116114a2578183600052602060002091820191016114a29190610ca4565b5050509190906000526020600020900160008f909190915055508c82600001600050819055508b82600101600050819055508a8260020160006101000a81548163ffffffff021916908360e060020a90040217905550898260030160005081905550878260050160006101000a815481600160a060020a0302191690830217905550868260060160006101000a815481600160a060020a0302191690830217905550858260070160006101000a815481600160a060020a0302191690830217905550848260080160006101000a815481600160a060020a03021916908302179055507f30303030303031000000000000000000000000000000000000000000000000008260080160146101000a81548167ffffffffffffffff021916908360c060020a90040217905550848260090160006101000a815481600160a060020a0302191690830217905550848260090160006101000a815481600160a060020a03021916908302179055506002600050600033600160a060020a0316815260200190815260200160002060005090506113a78d600060008460000160005054856001016000505486600201600050546107a5565b600085815260056020526040902060080154600160a060020a0390811633909116146116d757600192507f596f7520646f6e2774206861732061636365737300000000000000000000000091506116cf565b600092507f737563636573730000000000000000000000000000000000000000000000000091505b509250929050565b600085815260056020908152604080832060088101805460a060020a67ffffffffffffffff0219167b30303030303032000000000000000000000000000000000000000000179055600a01805473ffffffffffffffffffffffffffffffffffffffff191688179055600160a060020a0387168352600390915290208054600181018083558281838015829011611780578183600052602060002091820191016117809190610ca4565b505050600092835250602080832090910187905533600160a060020a03168252600290819052604082208054600182810154938301549295506116a7948a949193909291906107a5565b60019182019101610b1d565b505050505b5050505056"

	contractdeploy := NewContract(sender, to, value, gas, gasPrice)
	contractdeploy.SetCallCode(&common.Address{}, common.FromHex(code))

	expectRet := "0x6060604052361561008d5760e060020a60003504631d3ac435811461009257806332c1680c1461013657806354f445f1146101825780635d47f91c146101a1578063b5cf8eab14610273578063bf7716ac14610280578063cb7f0472146102fb578063e98e84d414610464578063eb3c667f1461056a578063ed29244f146105e7578063ff86a17914610662575b610002565b34610002576107046004356024356040805160208101909152600080825290818080808080600160e060020a0319891660e460020a6305355303021415610a8b57898152600560205260408120600a0154600160a060020a03908116339091161461097c5760408051808201909152601481527f596f7520646f6e277420686173206163636573730000000000000000000000006020820152600198509650610cf7565b346100025733600160a060020a03166000908152600260208190526040909120602435815560443591810191909155600435600191820155604080519115158252519081900360200190f35b346100025761077c60043560243560443560643560843560a4356107a5565b346100025761081d6004355b60408051602081810183526000808352835180830185528181528451808401865282815285518085018752838152865194850190965282845291949092858080808080805b60008e815260208190526040902054821015610f2157604060009081208f825260209190915280548390811015610002579060005260206000209060050201600050905086805480600101828181548183558181151161112b57601f016020900481601f0160209004836000526020600020918201910161112b9190610ca4565b346100025761095e610980565b34610002576040805160208181018352600080835233600160a060020a031681526003825283902080548451818402810184019095528085526109ae94928301828280156102f057602002820191906000526020600020905b8160005054815260200190600101908083116102d9575b505050505090505b90565b34610002576109f8600435600081815260056020819052604082209081015482918291829182918291829182918291829190600160a060020a039081163391909116148061035a57506006810154600160a060020a0390811633909116145b8061037657506007810154600160a060020a0390811633909116145b8061039257506008810154600160a060020a0390811633909116145b1561045657806000016000505481600101600050548260020160009054906101000a900460e060020a02836003016000505484600401600050548560080160149054906101000a900460c060020a028660050160009054906101000a9004600160a060020a03168760060160009054906101000a9004600160a060020a03168860070160009054906101000a9004600160a060020a03168960080160009054906101000a9004600160a060020a03169a509a509a509a509a509a509a509a509a509a505b509193959799509193959799565b346100025761081d60043560408051602081810183526000808352835180830185528181528451808401865282815285518085018752838152865180860188528481528751808701895285815288518088018a5286815289518089018b528781528a51808a018c528881528b51808b018d528981528d8a5260059a8b90529b8920998a0154989b9799969895979496958c9590600160a060020a039081163391909116148061052457506006870154600160a060020a0390811633909116145b8061054057506007870154600160a060020a0390811633909116145b8061055c57506008870154600160a060020a0390811633909116145b1561111b576112dc8e6101ad565b346100025761070460043560243560443560643560843560a43560c43560e4356101043560408051602081019091526000808252908180888a126112ff5760408051808201909152601a81527f69737375654461746520626967207468656e206475656461746500000000000060208201526001945092506113e2565b34610002576040805160208181018352600080835233600160a060020a031681526001825283902080548451818402810184019095528085526109ae94928301828280156102f057602002820191906000526020600020908160005054815260200190600101908083116102d9575b505050505090506102f8565b3461000257600435600081815260056020526040812060080154610a6d9291602435918190819074010000000000000000000000000000000000000000900460c060020a02600160c060020a0319167f3030303030303200000000000000000000000000000000000000000000000000141561165557600192507f686173206265656e20656e64720000000000000000000000000000000000000091506116cf565b604051808360ff168152602001806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f16801561076d5780820380516001836020036101000a031916815260200191505b50935050505060405180910390f35b005b5050506000928352506020909120018a9055610cbc8a600160e460020a6305355303028989895b600086815260208190526040902080546001810180835590918291828015829011610d8457600502816005028360005260206000209182019101610d8491905b80821115610cb857600080825560018201805464ffffffffff19169055600282018190556003820181905560048201556005016107e5565b6040518087600019168152602001806020018060200180602001806020018060200186810386528b8181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f15090500186810385528a8181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f1509050018681038452898181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f1509050018681038352888181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f1509050018681038252878181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f1509050019b50505050505050505050505060405180910390f35b60408051938452602084019290925282820152519081900360600190f35b610af35b33600160a060020a031660009081526002602081905260409091208054600182015491909201549192909190565b60405180806020018281038252838181518152602001915080519060200190602002808383829060006004602084601f0104600302600f01f1509050019250505060405180910390f35b604080519a8b5260208b0199909952600160e060020a03199097168989015260608901959095526080880193909352600160c060020a03199190911660a0870152600160a060020a0390811660c087015290811660e08601529081166101008501521661012083015251908190036101400190f35b6040805160ff93909316835260208301919091528051918290030190f35b600160e060020a0319891660e060020a6353553031021415610cf75760008a8152600560205260409020600801805460a060020a67ffffffffffffffff0219167b30303030303032000000000000000000000000000000000000000000179055610d04610980565b600160a060020a03331660009081526003602052604090209298509096509450610bba908b5b6000805b8354811015610b4c578260001916848281548110156100025760009182526020909120015414156117ca579050805b835482146117db578354849060001981019081101561000257906000526020600020900160005054848381548110156100025760206000908120929052015583546000198101808655908590829080158290116117d6576000838152602090206117d6918101908301610ca4565b60008a815260056020908152604080832060080154600160a060020a0316835260019091529020610beb908b610b19565b60008a81526005602090815260408083206008810180546009909201805473ffffffffffffffffffffffffffffffffffffffff19908116339081179092557b3030303030303300000000000000000000000000000000000000000060a060020a67ffffffffffffffff02199094169390931790921682179055600160a060020a031683526001918290529091208054918201808255909190828183801582901161077e5781836000526020600020918201910161077e91905b80821115610cb85760008155600101610ca4565b5090565b60408051808201909152601481527f656e6472726573706f6e7365207375636365656400000000000000000000000060208201526000985096505b5050505050509250929050565b925092509250610d238a600160e060020a6353553031028686866107a5565b600160a060020a0333166000908152600360205260409020610d45908b610b19565b60408051808201909152601281527f656e64726573706f6e7365206661696c656400000000000000000000000000006020820152600198509650610cf7565b5050506000888152602081905260409020805489925083908110156100025790600052602060002090600502016000505560008781526020819052604090208054879190839081101561000257906000526020600020906005020160005060010160006101000a81548160ff0219169083021790555084600060005060008960001916815260200190815260200160002060005082815481101561000257906000526020600020906005020160005060010160016101000a81548163ffffffff021916908360e060020a9004021790555083600060005060008960001916815260200190815260200160002060005082815481101561000257906000526020600020906005020160005060020160005081905550826000600050600089600019168152602001908152602001600020600050828154811015610002579060005260206000209060050201600050600301600050819055508160006000506000896000191681526020019081526020016000206000508281548110156100025790600052602060002090600502016000506004015550505050505050565b8d878787878784805480602002602001604051908101604052809291908181526020018280548015610f9057602002820191906000526020600020906000905b825461010083900a900460ff16815260206001928301818104948501949093039092029101808411610f615790505b505050505094508380548060200260200160405190810160405280929190818152602001828054801561100857602002820191906000526020600020906000905b82829054906101000a900460e060020a0281526020019060040190602082600301049283019260010382029150808411610fd15790505b505050505093508280548060200260200160405190810160405280929190818152602001828054801561105d57602002820191906000526020600020905b816000505481526020019060010190808311611046575b50505050509250818054806020026020016040519081016040528092919081815260200182805480156110b257602002820191906000526020600020905b81600050548152602001906001019080831161109b575b505050505091508080548060200260200160405190810160405280929190818152602001828054801561110757602002820191906000526020600020905b8160005054815260200190600101908083116110f0575b505050505090509c509c509c509c509c509c505b5050505050505091939550919395565b5050506000928352506020918290206001848101548484049092018054949093066101000a60ff8181021990951692909416909302179055865490810180885587919082818380158290116111a15760070160089004816007016008900483600052602060002091820191016111a19190610ca4565b505050919090600052602060002090600891828204019190066004028360010160019054906101000a900460e060020a02909190916101000a81548163ffffffff021916908360e060020a900402179055505084805480600101828181548183558181151161122157600083815260209020611221918101908301610ca4565b505050919090600052602060002090016000506002830154905550835460018101808655859190828183801582901161126b5760008381526020902061126b918101908301610ca4565b50505091909060005260206000209001600050600383015490555082546001810180855584919082818380158290116112b5576000838152602090206112b5918101908301610ca4565b505050919090600052602060002090016000506004830154905550600191909101906101f2565b9550955095509550955095508585858585859c509c509c509c509c509c5061111b565b7f4143303100000000000000000000000000000000000000000000000000000000600160e060020a03198c161480159061136357507f4143303200000000000000000000000000000000000000000000000000000000600160e060020a03198c1614155b156113f25760408051808201909152601581527f62696c6c74797065206973206e6f74206d61746368000000000000000000000060208201526001945092506113e2565b60408051808201909152601381527f70756242696c6c496e666f20737563636565640000000000000000000000000060208201526000945092505b5050995099975050505050505050565b33600160a060020a031688600160a060020a031614151561144c5760408051808201909152601181527f64727772206973206e6f74206d6174636800000000000000000000000000000060208201526001945092506113e2565b60008d8152600560209081526040808320600160a060020a0389168452600192839052922080549182018082559294509182818380158290116114a2578183600052602060002091820191016114a29190610ca4565b5050509190906000526020600020900160008f909190915055508c82600001600050819055508b82600101600050819055508a8260020160006101000a81548163ffffffff021916908360e060020a90040217905550898260030160005081905550878260050160006101000a815481600160a060020a0302191690830217905550868260060160006101000a815481600160a060020a0302191690830217905550858260070160006101000a815481600160a060020a0302191690830217905550848260080160006101000a815481600160a060020a03021916908302179055507f30303030303031000000000000000000000000000000000000000000000000008260080160146101000a81548167ffffffffffffffff021916908360c060020a90040217905550848260090160006101000a815481600160a060020a0302191690830217905550848260090160006101000a815481600160a060020a03021916908302179055506002600050600033600160a060020a0316815260200190815260200160002060005090506113a78d600060008460000160005054856001016000505486600201600050546107a5565b600085815260056020526040902060080154600160a060020a0390811633909116146116d757600192507f596f7520646f6e2774206861732061636365737300000000000000000000000091506116cf565b600092507f737563636573730000000000000000000000000000000000000000000000000091505b509250929050565b600085815260056020908152604080832060088101805460a060020a67ffffffffffffffff0219167b30303030303032000000000000000000000000000000000000000000179055600a01805473ffffffffffffffffffffffffffffffffffffffff191688179055600160a060020a0387168352600390915290208054600181018083558281838015829011611780578183600052602060002091820191016117809190610ca4565b505050600092835250602080832090910187905533600160a060020a03168252600290819052604082208054600182810154938301549295506116a7948a949193909291906107a5565b60019182019101610b1d565b505050505b5050505056"

	ret1, err := evm.Run(contractdeploy, nil)
	if err!=nil{
		t.Error("contract deploy error")
		t.Fatal(err)
	}
	if expectRet!=common.ToHex(ret1){
		t.Error("contract deploy error,ret doesn't match")
	}
	contract := NewContract(sender,to,value,gas,gasPrice)
	contract.SetCallCode(&contractAddr,ret1)

	input := common.FromHex("0x32c1680c100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000000")
	ret,err := evm.Run(contract,input)
	if err!=nil{
		t.Error("contract invoke error")
	}
	if common.ToHex(ret)!="0x0000000000000000000000000000000000000000000000000000000000000001"{
		t.Error("contract invoke error")
	}
}

func TestEVM_Run2(t *testing.T) {

	vmenv := NewTestVmenv()
	evm := vmenv.evm

	from := common.HexToAddress("000f1a7a08ccc48e5d30f80850cf1cf283aa3abd")

	sender := evm.env.Db().CreateAccount(from)
	contractAddr := common.HexToAddress("0xbbe2b6412ccf633222374de8958f2acc76cda9c9")
	to := evm.env.Db().CreateAccount(contractAddr)
	value := big.NewInt(0)
	code :="0x606060405260918060106000396000f3606060405260e060020a6000350463eb8ac9218114601c575b6002565b34600257607f60243560043581028101819003906000818381156002570492508183811560025760029190060a82018290038202925081838115600257049250818381156002576004600081905591900683178316831890921890910192915050565b60408051918252519081900360200190f3"

	contractdeploy := NewContract(sender, to, value, gas, gasPrice)
	contractdeploy.SetCallCode(&common.Address{}, common.FromHex(code))
	expectRet := "0x606060405260e060020a6000350463eb8ac9218114601c575b6002565b34600257607f60243560043581028101819003906000818381156002570492508183811560025760029190060a82018290038202925081838115600257049250818381156002576004600081905591900683178316831890921890910192915050565b60408051918252519081900360200190f3"

	ret1, err:= evm.Run(contractdeploy, nil)
	if err!=nil{
		t.Error("contract deploy error")
	}
	if expectRet!=common.ToHex(ret1){
		t.Error("contract deploy error,ret doesn't match")
	}

	contract := NewContract(sender,to,value,gas,gasPrice)
	contract.SetCallCode(&contractAddr,ret1)

	input := common.FromHex("0xeb8ac92100000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000006")
	ret,err := evm.Run(contract,input)
	if err!=nil{
		t.Error("contract deploy error")
	}
	if common.ToHex(ret)!="0x000000000000000000000000000000000000000000000000000000000000000a"{
		t.Error("contract invoke error")
	}
}
func TestEVM_RunPrecompiled(t *testing.T) {
	pcontracts := PrecompiledContracts()
	sha256key :=string(common.LeftPadBytes([]byte{2}, 20))
	p := pcontracts[sha256key]
	vmenv := NewTestVmenv()
	evm:=vmenv.evm
	from := common.HexToAddress("000f1a7a08ccc48e5d30f80850cf1cf283aa3abd")
	sender := evm.env.Db().CreateAccount(from)
	contractAddr := common.HexToAddress("0xbbe2b6412ccf633222374de8958f2acc76cda9c9")
	to := evm.env.Db().CreateAccount(contractAddr)
	value := big.NewInt(0)
	contract := NewContract(sender,to,value,gas,gasPrice)
	code :="0x606060405260918060106000396000f3606060405260e060020a6000350463eb8ac9218114601c575b6002565b34600257607f60243560043581028101819003906000818381156002570492508183811560025760029190060a82018290038202925081838115600257049250818381156002576004600081905591900683178316831890921890910192915050565b60408051918252519081900360200190f3"

	input := common.FromHex(code)

	ret,err:=evm.RunPrecompiled(p,input,contract)
	fmt.Println(common.ToHex(ret))
	if err!=nil{
		t.Error("run precompiled error")
	}

}