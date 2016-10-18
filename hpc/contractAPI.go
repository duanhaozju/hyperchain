package hpc

import (
	"hyperchain/common"
	"hyperchain/core/vm/compiler"
	"time"
	"github.com/golang/protobuf/proto"
	"hyperchain/manager"
	"hyperchain/core/types"
	"hyperchain/event"
	"fmt"
	"errors"
	"encoding/hex"
	"hyperchain/crypto"
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
	//var found bool
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

	//if contract.pm == nil {
	//
	//	// Test environment
	//	found = true
	//} else {
	//
	//	// Development environment
	//	am := contract.pm.AccountManager
	//	_, found = am.Unlocked[args.From]
	//
	//}
	//am := tran.pm.AccountManager

	//if found == true {
		log.Infof("############# %d: start send request#############", time.Now().Unix())
		tx.Timestamp = time.Now().UnixNano()
		tx.Id = uint64(contract.pm.Peermanager.GetNodeId())

		if realArgs.PrivKey == "" {
			// For Hyperchain test

			// TODO replace password with test value
			signature, err := contract.pm.AccountManager.Sign(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes())
			if err != nil {
				log.Errorf("Sign(tx) error :%v", err)
			}
			tx.Signature = signature
		} else {
			// For Dashboard test

			key, err := hex.DecodeString(args.PrivKey)
			if err != nil {
				return common.Hash{}, err
			}
			pri := crypto.ToECDSA(key)

			hash := tx.SighHash(kec256Hash).Bytes()
			sig, err := encryption.Sign(hash, pri)
			if err != nil {
				return common.Hash{}, err
			}

			tx.Signature = sig
		}

		tx.TransactionHash = tx.GetTransactionHash().Bytes()

		// Unsign
		if !tx.ValidateSign(contract.pm.AccountManager.Encryption, kec256Hash) {
			return common.Hash{}, errors.New("invalid signature")
		}

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
	//} else {
	//	return common.Hash{}, errors.New("account don't unlock")
	//}

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
	return tx.GetTransactionHash(), nil
}

type CompileCode struct{
	Abi []string
	Bin []string
	Types []string
}

// ComplieContract complies contract to ABI
func (contract *PublicContractAPI) CompileContract(ct string) (*CompileCode,error){
	abi, bin, names, err := compiler.CompileSourcefile(ct)

	if err != nil {
		return nil, err
	}

	return &CompileCode{
		Abi: abi,
		Bin: bin,
		Types: names,
	}, nil
}

// DeployContract deploys contract.
func (contract *PublicContractAPI) DeployContract(args SendTxArgs) (common.Hash, error) {
	return deployOrInvoke(contract, args)
}

// InvokeContract invokes contract.
func (contract *PublicContractAPI) InvokeContract(args SendTxArgs) (common.Hash, error) {
	return deployOrInvoke(contract, args)
}

// GetCode returns the code from the given contract address and block number.
func (contract *PublicContractAPI) GetCode(addr common.Address, n Number) (string, error) {

	_, stateDb, err := getBlockAndStateDb(n)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
		return "", err
	}

	return fmt.Sprintf(`0x%x`, stateDb.GetCode(addr)), nil
}

// GetContractCountByAddr returns the number of contract that has been deployed by given account address,
// if addr is nil, returns the number of all the contract that has been deployed.
func (contract *PublicContractAPI) GetContractCountByAddr(addr common.Address) (uint64, error) {

	_, stateDb, err := getBlockAndStateDb(Number(latestBlockNumber))

	if err != nil {
		return 0, err
	}

	return stateDb.GetNonce(addr), nil

}



