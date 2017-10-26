package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto/secp256k1"
	"testing"
)

func TestTOECDSAPub(t *testing.T) {
	pubs := "047c1d9ed2034f142a38e7c6887f7b027adb0490d9f27ede50b727956708c20040134e4b3db89d7876644b660c8d1b24570db5658b6bc7af73722c3b618a45ae9c"
	pubb := common.Hex2Bytes(pubs)
	P0 := pubb[0]
	fmt.Println("P0[0]", P0)
	X := pubb[1:33]
	fmt.Println("X:", X)
	Y := pubb[33:65]
	fmt.Println("Y", Y)
	x := common.Bytes2Big(X)
	y := common.Bytes2Big(Y)
	fmt.Println("x", x)
	fmt.Println("y", y)
	fmt.Println(common.Hex2Bytes(pubs))

	//pubkey := TOECDSAPub(common.Hex2Bytes(pubs))
	//var pubkey *ecdsa.PublicKey
	pubkey := new(ecdsa.PublicKey)
	pubkey.Curve = secp256k1.S256()
	pubkey.X = x
	pubkey.Y = y
	address := PubkeyToAddress(*pubkey)
	fmt.Println(address.Hex())

	msg := "hyperchain"
	msgb := []byte(msg)
	fmt.Println("msgb: ", msgb)
	fmt.Println("msghex: ", common.ToHex(msgb))

	hashb := Keccak256(msgb)
	fmt.Println("hashb: ", hashb)
	fmt.Println("hashhex: ", common.ToHex(hashb))

	signatures := "2d3b1e3648aef9ed071c078fd4a859478a1e9773de5f96877e7073f41d021b37022ccf9966eda9fff38483edbe35beadd873f89e71b14b7e04f738afb43ead8201"
	signatureb := common.Hex2BytesFixed(signatures, 65)

	recoveredpub, _ := secp256k1.RecoverPubkey(hashb, signatureb)

	fmt.Println("recoverd pubkey")
	fmt.Println(recoveredpub)

	fmt.Println("recoverd address")

	recoveredpubkey := new(ecdsa.PublicKey)
	recoveredpubkey.Curve = secp256k1.S256()
	recoveredpubkey.X = common.Bytes2Big(recoveredpub[1:33])
	recoveredpubkey.Y = common.Bytes2Big(recoveredpub[33:65])

	fmt.Println(PubkeyToAddress(*recoveredpubkey).Hex())

	//
	//pubb := FromECDSAPub(pubkey)
	//
	//assert.Equal(t,common.ToHex(pubb),pubs)
}

func TestKeccak256Hash_Hash(t *testing.T) {
	msg := "hyperchain"
	msgb := []byte(msg)
	fmt.Println("msgb: ", msgb)
	fmt.Println("msghex: ", common.ToHex(msgb))

	hashb := Keccak256(msgb)
	fmt.Println("hashb: ", hashb)
	fmt.Println("hashhex: ", common.ToHex(hashb))
}
