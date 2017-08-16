package message

const p2pDevVer = 14

func NewPkg(payload []byte, tipe ControlType) *Package {
	return &Package{
		Version:p2pDevVer,
		Payload:payload,
		Type:tipe,
	}
}
