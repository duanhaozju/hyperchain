package version1_3

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
	Bloom        string
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
	Bloom        string
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
	Bloom             string
	CumulativeGasUsed int64
	TxHash            string
	ContractAddress   string
	GasUsed           int64
	Ret               string
	Logs              string
	Status            string
	Message           string
	VmType            string
}

type ChainView struct {
	Version         string
	LatestBlockHash string
	ParentBlockHash string
	Height          uint64
	Genesis         uint64
	CurrentTxSum    uint64
	Extra           string
}
