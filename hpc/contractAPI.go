package hpc

import (
	"hyperchain/common"
	"hyperchain/core/vm/compiler"
	"time"
	"github.com/golang/protobuf/proto"
	"hyperchain/manager"
	"hyperchain/core/types"
	"hyperchain/event"
	"errors"
)

type PublicContractAPI struct {
	eventMux *event.TypeMux
	pm *manager.ProtocolManager
}

func NewPublicContractAPI(eventMux *event.TypeMux, pm *manager.ProtocolManager) *PublicContractAPI {
	return &PublicContractAPI{
		eventMux :eventMux,
		pm:pm,
	}
}

func deployOrInvoke(contract *PublicContractAPI, args SendTxArgs) (common.Hash, error) {
	var tx *types.Transaction
	var found bool

	realArgs := prepareExcute(args)

	payload := common.FromHex(realArgs.Payload)

	txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(),realArgs.Gas.ToInt64(),realArgs.Value.ToInt64(),payload)

	value, err := proto.Marshal(txValue)

	if err != nil {
		return common.Hash{}, err
	}

	if args.To == nil {

		// 部署合约
		//tx = types.NewTransaction(realArgs.From[:], nil, value, []byte(args.Signature))
		tx = types.NewTransaction(realArgs.From[:], nil, value)

	} else {

		// 调用合约或者普通交易(普通交易还需要加检查余额)
		//tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, []byte(args.Signature))
		tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value)
	}

	if contract.pm == nil {

		// Test environment
		found = true
	} else {

		// Development environment
		am := contract.pm.AccountManager
		_, found = am.Unlocked[args.From]

		//// TODO replace password with test value
		//signature, err := am.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), "123")
		//if err != nil {
		//	log.Errorf("Sign(tx) error :%v", err)
		//}
		//tx.Signature = signature

	}
	//am := tran.pm.AccountManager

	if found == true {
		log.Infof("############# %d: start send request#############", time.Now().Unix())
		tx.Timestamp = time.Now().UnixNano()

		txBytes, err := proto.Marshal(tx)
		if err != nil {
			log.Errorf("proto.Marshal(tx) error: %v", err)
		}
		if manager.GetEventObject() != nil {
			go contract.eventMux.Post(event.NewTxEvent{Payload: txBytes})
		} else {
			log.Warning("manager is Nil")
		}

		log.Infof("############# %d: end send request#############", time.Now().Unix())
	} else {
		return common.Hash{}, errors.New("account don't unlock")
	}

	time.Sleep(2000 * time.Millisecond)
	/*
		receipt := core.GetReceipt(tx.BuildHash())
		fmt.Println("GasUsed", receipt.GasUsed)
		fmt.Println("PostState", receipt.PostState)
		fmt.Println("ContractAddress", receipt.ContractAddress)
		fmt.Println("CumulativeGasUsed", receipt.CumulativeGasUsed)
		fmt.Println("Ret", receipt.Ret)
		fmt.Println("TxHash", receipt.TxHash)
		fmt.Println("Status", receipt.Status)
		fmt.Println("Message", receipt.Message)
		fmt.Println("Log", receipt.Logs)
	*/
	return tx.BuildHash(), nil
}

type CompileCode struct{
	Abi []string
	Bin []string
}

// ComplieContract complies contract to ABI
func (contract *PublicContractAPI) CompileContract(ct string) (*CompileCode,error){
	abi, bin, err := compiler.CompileSourcefile(ct)

	if err != nil {
		return nil, err
	}

	return &CompileCode{
		Abi: abi,
		Bin: bin,
	}, nil
}

// DeployContract deploys contract.
func (contract *PublicContractAPI) DeployContract(args SendTxArgs) (common.Hash, error) {
	return deployOrInvoke(contract, args)
}

// InvokeContract invoke contract.
func (contract *PublicContractAPI) InvokeContract(args SendTxArgs) (common.Hash, error) {
	return deployOrInvoke(contract, args)
}
