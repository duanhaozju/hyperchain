package constant

const (
	PATH     = "path"
	DATABASE = "database"
	NUMBER   = "number"
	HASH     = "hash"
	OUTPUT   = "output"
	VERBOSE  = "verbose"

	VERSIONFINAL = "final"
	VERSION1_2   = "1.2"
	VERSION1_3   = "1.3"

	BLOCK              = "block"
	TRANSACTION        = "transaction"
	INVAILDTRANSACTION = "invaildTransaction"
	TRANSACTIONMETA    = "transactionMeta"
	RECEIPT            = "receipt"
	CHAIN              = "chain"

	PROTOERR           = "proto: bad wiretype for field"
)

var (
	VERSIONS     = []string{VERSION1_2, VERSION1_3}
)