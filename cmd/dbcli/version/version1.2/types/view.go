package version1_2

type BlockVerboseView struct {
	Version      string
	ParentHash   string
	BlockHash    string
	Transactions []*TransactionView
	Timestamp    int64
	MerkleRoot   string
	TxRoot       string
	ReceiptRoot  string
	Number       uint64
	WriteTime    int64
	CommitTime   int64
	EvmTime      int64
}

type BlockView struct {
	Version      string
	ParentHash   string
	BlockHash    string
	Transactions []*TransactionViewHash
	Timestamp    int64
	MerkleRoot   string
	TxRoot       string
	ReceiptRoot  string
	Number       uint64
	WriteTime    int64
	CommitTime   int64
	EvmTime      int64
}

type TransactionViewHash struct {
	TransactionHash string
}

type TransactionView struct {
	Version         string
	From            string
	To              string
	Value           string
	Timestamp       int64
	Signature       string
	Id              uint64
	TransactionHash string
	Nonce           int64
}

type InvalidTransactionView struct {
	Tx      *TransactionView
	ErrType string
	ErrMsg  string
}

type ReceiptView struct {
	Version           string
	PostState         string
	CumulativeGasUsed int64
	TxHash            string
	ContractAddress   string
	GasUsed           int64
	Ret               string
	Logs              string
	Status            string
	Message           string
}

type ChainView struct {
	LatestBlockHash  string
	ParentBlockHash  string
	Height           uint64
	RequiredBlockNum uint64
	RequireBlockHash string
	RecoveryNum      uint64
	CurrentTxSum     uint64
}
