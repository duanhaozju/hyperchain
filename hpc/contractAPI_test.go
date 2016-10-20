package hpc

import (
	"testing"
	"hyperchain/common"
)

var contractAPI = NewPublicContractAPI(nil, nil)

func TestPublicContractAPI_DeployContract(t *testing.T) {
	// deploy contract test
	args.To = nil
	args.Payload = "0x6000805463ffffffff1916815560a0604052600b6060527f68656c6c6f20776f726c6400000000000000000000000000000000000000000060805260018054918190527f68656c6c6f20776f726c6400000000000000000000000000000000000000001681559060be907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf66020600261010084871615026000190190931692909204601f01919091048101905b8082111560ce576000815560010160ac565b50506101cd806100d26000396000f35b509056606060405260e060020a60003504633ad14af3811461003f578063569c5f6d146100675780638da9b7721461007c578063d09de08a146100ea575b610002565b34610002576000805460043563ffffffff8216016024350163ffffffff19919091161790555b005b346100025761010e60005463ffffffff165b90565b3461000257604080516020818101835260008252600180548451600261010083851615026000190190921691909104601f81018490048402820184019095528481526101289490928301828280156101c15780601f10610196576101008083540402835291602001916101c1565b34610002576100656000805463ffffffff19811663ffffffff909116600101179055565b6040805163ffffffff929092168252519081900360200190f35b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f1680156101885780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b820191906000526020600020905b8154815290600101906020018083116101a457829003601f168201915b5050505050905061007956"

	if hash, err := contractAPI.DeployContract(args); err == nil {
		t.Logf("the new contract tx hash is %v", common.ToHex(hash[:]))
	} else {
		t.Errorf("%v", err)
	}
}

func TestPublicContractAPI_InvokeContract(t *testing.T) {

}