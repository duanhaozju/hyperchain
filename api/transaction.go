//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/juju/ratelimit"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/ledger/bloom"
	edb "hyperchain/core/ledger/chain"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/hyperdb/db"
	"hyperchain/manager"
	"hyperchain/manager/event"
	"time"
)

// This file implements the handler of Transaction service API which
// can be invoked by client in JSON-RPC request.

var (
	kec256Hash         = crypto.NewKeccak256Hash("keccak256")
	db_not_found_error = db.DB_NOT_FOUND.Error()
)

type Transaction struct {
	namespace   string
	eh          *manager.EventHub
	tokenBucket *ratelimit.Bucket
	config      *common.Config
	log         *logging.Logger
}

// SendTxArgs represents the arguments to submit a new transaction into the transaction pool.
type SendTxArgs struct {
	From      common.Address  `json:"from"`      // transaction sender address
	To        *common.Address `json:"to"`        // transaction receiver address
	Value     Number          `json:"value"`     // transaction amount
	Payload   string          `json:"payload"`   // contract payload
	Signature string          `json:"signature"` // signature of sender for the transaction
	Timestamp int64           `json:"timestamp"` // timestamp of the transaction happened
	Simulate  bool            `json:"simulate"`  // Simulate determines if the transaction requires consensus, if true, no consensus.
	Nonce     int64           `json:"nonce"`     // 16-bit random decimal number, for example 5956491387995926
	Extra     string          `json:"extra"`     // extra data stored in transaction
	VmType    string          `json:"type"`      // specify which engine executes contract

	// 1 value for Opcode means upgrading contract, 2 means freezing contract,
	// 3 means unfreezing contract, 4 means vm skipping, 100 means archiving data.
	Opcode int32 `json:"opcode"`

	// Snapshot saves the state of ledger at a moment.
	// SnapshotId specifies the based ledger when client sends transaction or invokes contract with Simulate=true.
	SnapshotId string `json:"snapshotId"`
}

// NewPublicTransactionAPI creates and returns a new Transaction instance for given namespace name.
func NewPublicTransactionAPI(namespace string, eh *manager.EventHub, config *common.Config) *Transaction {
	log := common.GetLogger(namespace, "api")
	fillrate, err := getFillRate(namespace, config, TRANSACTION)
	if err != nil {
		log.Errorf("invalid ratelimit fill rate parameters.")
		fillrate = 10 * time.Millisecond
	}
	peak := getRateLimitPeak(namespace, config, TRANSACTION)
	if peak == 0 {
		log.Errorf("got invalid ratelimit peak parameters as 0. use default peak parameters 500")
		peak = 500
	}
	return &Transaction{
		namespace:   namespace,
		eh:          eh,
		config:      config,
		tokenBucket: ratelimit.NewBucket(fillrate, peak),
		log:         log,
	}
}

// SendTransaction is to create a transaction object, and then post event NewTxEvent,
// if the sender's balance is not enough, account transfer will fail.
func (tran *Transaction) SendTransaction(args SendTxArgs) (common.Hash, error) {
	consentor := tran.eh.GetConsentor()
	normal, full := consentor.GetStatus()
	if !normal || full {
		return common.Hash{}, &common.SystemTooBusyError{}
	}

	if getRateLimitEnable(tran.config) && tran.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{}
	}

	// 1. create a new transaction instance
	tx, err := prepareTransaction(args, 0, tran.namespace, tran.eh)
	if err != nil {
		return common.Hash{}, err
	}

	// 2. post a event.NewTxEvent event
	err = postNewTxEvent(args, tx, tran.eh)
	if err != nil {
		return common.Hash{}, err
	}

	return tx.GetHash(), nil
}

func prepareTransaction(args SendTxArgs, txType int, namespace string, eh *manager.EventHub) (*types.Transaction, error) {

	var tx *types.Transaction
	var txValue *types.TransactionValue

	log := common.GetLogger(namespace, "api")

	// 1. verify if the parameters are valid
	realArgs, err := prepareExcute(args, txType)
	if err != nil {
		return nil, err
	}

	// 2. create a new transaction instance
	if txType == 0 {
		txValue = types.NewTransactionValue(DEFAULT_GAS_PRICE, DEFAULT_GAS,
			realArgs.Value.Int64(), nil, 0, []byte(args.Extra), types.TransactionValue_EVM)
	} else {
		payload := common.FromHex(realArgs.Payload)
		txValue = types.NewTransactionValue(DEFAULT_GAS_PRICE, DEFAULT_GAS,
			realArgs.Value.Int64(), payload, args.Opcode, []byte(args.Extra), parseVmType(realArgs.VmType))
	}

	value, err := proto.Marshal(txValue)
	if err != nil {
		return nil, &common.CallbackError{err.Error()}
	}

	if args.To == nil {
		tx = types.NewTransaction(realArgs.From[:], nil, value, realArgs.Timestamp, realArgs.Nonce)
	} else {
		tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp, realArgs.Nonce)
	}

	if eh.NodeIdentification() == manager.IdentificationVP {
		tx.Id = uint64(eh.GetPeerManager().GetNodeId())
	} else {
		hash := eh.GetPeerManager().GetLocalNodeHash()
		if err := tx.SetNVPHash(hash); err != nil {
			log.Errorf("set NVP hash failed! err Msg: %v.", err.Error())
			return nil, &common.CallbackError{Message: "Marshal nvp hash error"}
		}
	}
	tx.Signature = common.FromHex(realArgs.Signature)
	tx.TransactionHash = tx.Hash().Bytes()

	// 3. check if there is duplicated transaction
	var exist bool
	if exist, err = bloom.LookupTransaction(namespace, tx.GetHash()); err != nil || exist == true {
		if exist, _ = edb.IsTransactionExist(namespace, tx.TransactionHash); exist {
			return nil, &common.RepeadedTxError{TxHash: common.ToHex(tx.TransactionHash)}
		}
	}

	// 4. verify transaction signature
	if !tx.ValidateSign(eh.GetAccountManager().Encryption, kec256Hash) {
		log.Errorf("invalid signature, tx hash %v", common.ToHex(tx.TransactionHash))
		return nil, &common.SignatureInvalidError{Message: "Invalid signature, tx hash " + common.ToHex(tx.TransactionHash)}
	}

	return tx, nil
}

func postNewTxEvent(args SendTxArgs, tx *types.Transaction, eh *manager.EventHub) error {

	// post transaction event
	if eh.NodeIdentification() == manager.IdentificationNVP {
		ch := make(chan bool)
		go eh.GetEventObject().Post(event.NewTxEvent{
			Transaction: tx,
			Simulate:    args.Simulate,
			SnapshotId:  args.SnapshotId,
			Ch:          ch,
		})
		res := <-ch
		close(ch)
		if res == false {
			// nvp node fails to forward tx to vp node
			return &common.CallbackError{Message: "Send tx to nvp failed."}
		}
	} else {
		go eh.GetEventObject().Post(event.NewTxEvent{
			Transaction: tx,
			Simulate:    args.Simulate,
			SnapshotId:  args.SnapshotId,
		})
	}
	return nil
}
