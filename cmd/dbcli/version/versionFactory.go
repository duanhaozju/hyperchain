package version

import (
	"hyperchain/cmd/dbcli/constant"
	"hyperchain/cmd/dbcli/version/version1.2"
)

//NewResultFactory -- get json result by struct type and struct version.
func NewResultFactory(typi string, version string, data []byte) (string, error) {
	switch typi {
	case constant.BLOCK:
		switch version {
		case constant.VERSION1_2:
			return version1_2.GetBlockData(data)
		default:
			return "", constant.ErrDataVersion
		}
	case constant.TRANSACTION:
		switch version {
		case constant.VERSION1_2:
			return version1_2.GetTransactionData(data)
		default:
			return "", constant.ErrDataVersion
		}
	case constant.INVAILDTRANSACTION:
		switch version {
		case constant.VERSIONFINAL:
			return version1_2.GetInvaildTransactionData(data)
		default:
			return "", constant.ErrDataVersion
		}
	case constant.TRANSACTIONMETA:
		switch version {
		case constant.VERSIONFINAL:
			return version1_2.GetTransactionMetaData(data)
		default:
			return "", constant.ErrDataVersion
		}
	case constant.RECEIPT:
		switch version {
		case constant.VERSION1_2:
			return version1_2.GetReceiptData(data)
		default:
			return "", constant.ErrDataVersion
		}
	case constant.CHAIN:
		switch version {
		case constant.VERSIONFINAL:
			return version1_2.GetChainData(data)
		default:
			return "", constant.ErrDataVersion
		}
	default:
		return "", constant.ErrDataType
	}
}
