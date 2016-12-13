package core

type Receipt interface {
	Version() string
}

func GetReceiptUnmarshalPolicy(version string, data []byte) {
	switch version {
	case "1.0":

	}
}
