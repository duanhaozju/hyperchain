package version

import (
	"github.com/hyperchain/hyperchain/cmd/dbcli/constant"
	"github.com/hyperchain/hyperchain/cmd/dbcli/version/version1.1"
	"github.com/hyperchain/hyperchain/cmd/dbcli/version/version1.2"
	"github.com/hyperchain/hyperchain/cmd/dbcli/version/version1.3"
)

//NewResultFactory -- get json result by struct type and struct version.
func NewResultFactory(typi string, version string, data []byte, parameter *constant.Parameter) (string, error) {
	switch typi {
	case constant.BLOCK:
		switch version {
		case constant.VERSION1_1:
			return version1_1.GetBlockData(data, parameter)
		case constant.VERSION1_2:
			return version1_2.GetBlockData(data, parameter)
		case constant.VERSION1_3:
			return version1_3.GetBlockData(data, parameter)
		default:
			return "", constant.ErrDataVersion
		}
	case constant.TRANSACTION:
		switch version {
		case constant.VERSION1_1:
			return version1_1.GetTransactionData(data, parameter)
		case constant.VERSION1_2:
			return version1_2.GetTransactionData(data, parameter)
		case constant.VERSION1_3:
			return version1_3.GetTransactionData(data, parameter)
		default:
			return "", constant.ErrDataVersion
		}
	case constant.INVAILDTRANSACTION:
		switch version {
		case constant.VERSION1_1:
			return version1_1.GetInvaildTransactionData(data)
		case constant.VERSION1_2:
			return version1_2.GetInvaildTransactionData(data)
		case constant.VERSION1_3:
			// version1.2 and version1.3 are same.
			return version1_2.GetInvaildTransactionData(data)
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
		case constant.VERSION1_1:
			return version1_1.GetReceiptData(data)
		case constant.VERSION1_2:
			return version1_2.GetReceiptData(data)
		case constant.VERSION1_3:
			return version1_3.GetReceiptData(data)
		default:
			return "", constant.ErrDataVersion
		}
	case constant.CHAIN:
		switch version {
		case constant.VERSION1_1:
			return version1_1.GetChainData(data)
		case constant.VERSION1_2:
			return version1_2.GetChainData(data)
		case constant.VERSION1_3:
			return version1_3.GetChainData(data)
		default:
			return "", constant.ErrDataVersion
		}
	case constant.CHAINHEIGHT:
		switch version {
		case constant.VERSION1_1:
			return version1_1.GetChainHeight(data)
		case constant.VERSION1_2:
			return version1_2.GetChainHeight(data)
		case constant.VERSION1_3:
			return version1_3.GetChainHeight(data)
		default:
			return "", constant.ErrDataVersion
		}
	default:
		return "", constant.ErrDataType
	}
}
