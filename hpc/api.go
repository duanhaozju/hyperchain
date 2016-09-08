package hpc

import (
	"hyperchain/common"
	"github.com/op/go-logging"
	"github.com/syndtr/goleveldb/leveldb/errors"
)


var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("jsonrpc/api")
}

type PublicTransactionAPI struct {

}


// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	//Gas      *jsonrpc.HexNumber  `json:"gas"`
	//GasPrice *jsonrpc.HexNumber  `json:"gasPrice"`
	//Value    *jsonrpc.HexNumber  `json:"value"`
	Value    []byte  `json:"value"`
	//Payload     []byte          `json:"payload"`
	//Data     string          `json:"data"`
	//Nonce    *jsonrpc.HexNumber  `json:"nonce"`
}

func NewPublicTransactionAPI() *PublicTransactionAPI {
	return &PublicTransactionAPI{}
}

func (hpc *PublicTransactionAPI) SendTransaction(args SendTxArgs) error{

	log.Info("==========SendTransaction=====,args = ",args)

	//var tx *types.Transaction
	//
	//tx = types.NewTransaction([]byte(args.From), []byte(args.To), []byte(args.Value))
	//
	//if (core.VerifyBalance(tx)) {
	//
	//	// Balance is enough
	//	/*txBytes, err := proto.Marshal(tx)
	//	if err != nil {
	//		log.Fatalf("proto.Marshal(tx) error: %v",err)
	//	}*/
	//
	//	//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
	//
	//	log.Infof("############# %d: start send request#############", time.Now().Unix())
	//	start := time.Now().Unix()
	//	end:=start+1
	//
	//	for start := start ; start < end; start = time.Now().Unix() {
	//		for i := 0; i < 5000; i++ {
	//			tx.TimeStamp=time.Now().UnixNano()
	//			txBytes, err := proto.Marshal(tx)
	//			if err != nil {
	//				log.Fatalf("proto.Marshal(tx) error: %v",err)
	//			}
	//			go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
	//			time.Sleep(200 * time.Microsecond)
	//		}
	//	}
	//
	//	log.Infof("############# %d: end send request#############", time.Now().Unix())
	//
	//	//tx.TimeStamp=time.Now().UnixNano()
	//	//txBytes, err := proto.Marshal(tx)
	//	//if err != nil {
	//	//	log.Fatalf("proto.Marshal(tx) error: %v",err)
	//	//}
	//	//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
	//
	//	return true
	//
	//} else {
	//	// Balance isn't enough
	//	return false
	//}

	return errors.New("nothing")
}

