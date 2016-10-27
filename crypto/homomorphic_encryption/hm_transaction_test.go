package homomorphic_encryption

import (
	"fmt"
	"math/big"
	"testing"
)

func Generate_wholenetpublickey(g []byte, n []byte, nsquare []byte) PaillierPublickey {
	var publickey PaillierPublickey
	G1 := new(big.Int)
	n1 := new(big.Int)
	nsquare1 := new(big.Int)
	publickey.G = G1.SetBytes(g)
	publickey.N = n1.SetBytes(n)
	publickey.nsquare = nsquare1.SetBytes(nsquare)
	return publickey
}

func Test_Hm_transaction(t *testing.T) {
	old := big.NewInt(500)
	old_byte := old.Bytes()
	trans := big.NewInt(200)
	trans_byte := trans.Bytes()

	oldbalance := make([]byte, 16)
	oldbalance = append(oldbalance[:(16-len(old_byte))], old_byte...)

	transferamount := make([]byte, 16)
	transferamount = append(transferamount[:(16-len(trans_byte))], trans_byte...)

	G1 := []byte{195, 153, 245, 222, 195, 152, 112, 47, 129, 79, 150, 122, 83, 4, 237, 142, 251, 215, 241, 160, 24, 216, 246, 105, 237, 83, 182, 53, 214, 53, 139, 73}
	N1 := []byte{238, 223, 119, 206, 43, 23, 14, 2, 159, 122, 233, 6, 43, 53, 223}
	nsquare1 := []byte{222, 228, 69, 213, 73, 95, 66, 80, 24, 124, 30, 174, 198, 90, 112, 198, 28, 116, 247, 151, 13, 92, 246, 31, 56, 225, 40, 131, 246, 8, 234, 161}
	publickey := Generate_wholenetpublickey(G1, N1, nsquare1)

	oldBalance_hm := []byte{139, 222, 135, 44, 112, 98, 183, 241, 13, 118, 121, 159, 232, 199, 15, 122, 142, 199, 189, 205, 14, 101, 203, 237, 113, 142, 85, 61, 15, 114, 205, 208}
	transferAmount_hm := []byte{33, 91, 137, 191, 96, 106, 118, 67, 185, 132, 43, 1, 104, 139, 5, 168, 242, 111, 122, 196, 29, 56, 118, 141, 111, 103, 112, 191, 17, 82, 21, 98}
	newBalance_hm := []byte{60, 17, 254, 228, 200, 71, 7, 13, 242, 62, 3, 241, 78, 200, 135, 145, 94, 162, 142, 172, 63, 58, 141, 153, 211, 230, 51, 98, 86, 97, 157, 94}
	flag := Node_Verify(publickey, oldBalance_hm, transferAmount_hm, newBalance_hm)
	fmt.Println(flag)

}
