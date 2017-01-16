package transport

import (
	"hyperchain/p2p/transport/ecdh"
	"hyperchain/p2p/transport/SMCrypto"
	"hyperchain/membersrvc"
	"encoding/hex"
	"errors"
	"crypto/elliptic"
)

type GuomiHandShakeManager struct {
	e                 ecdh.ECDH
	sessionPrivateKey string
	sessionPublicKey  string
	secrets           map[string][]byte
	signPublickey     map[string][]byte
	isVerified        map[string]bool
	cm                *membersrvc.CAManager
	smc               *SMCrypto.SMCrypto
}

func NewGuomiHandShakeManager(cm *membersrvc.CAManager) *GuomiHandShakeManager{
	var ghSMN GuomiHandShakeManager
	ghSMN.secrets = make(map[string][]byte)
	ghSMN.signPublickey = make(map[string][]byte)
	ghSMN.isVerified = make(map[string]bool)
	ghSMN.e = ecdh.NewEllipticECDH(elliptic.P384())
	//var err error
	//若无私钥，相当于无ecert,但为确保节点启动，自动生产公私钥对
	//if(cm.GetIsUsed()==false){
	//	ghSMN.sessionPrivateKey, ghSMN.sessionPublicKey, err = ghSMN.e.GenerateKey(rand.Reader)
	//	if err!=nil {
	//		panic("GenerateKey failed,please restart the node.")
	//	}
	//}else {
	//	//var pri *ecdsa.PrivateKey
	//	contentPri := cm.GetECertPrivateKeyByte()
	//	pri,err1 := primitives.ParseKey(string(contentPri))
	//	privateKey := pri.(*ecdsa.PrivateKey)
	//	//cert := primitives.ParseCertificate(contenrPub)
	//	if err1!=nil {
	//		panic("Parse PrivateKey or Ecert failed,please check the privateKey or Ecert and restart the node!")
	//	}else {
	//		ghSMN.sessionPrivateKey = privateKey
	//		ghSMN.sessionPublicKey = (*privateKey).PublicKey
	//	}
	//}

	//生成国密公私要

	return &ghSMN
}


func (ghsm *GuomiHandShakeManager)GenerateSecret(remotePublicKey []byte, peerHash string) error{
	//comfirm the key
	return nil
}

func (ghsm *GuomiHandShakeManager)EncWithSecret(message []byte, peerHash string) ([]byte,error){
	if _,ok := ghsm.secrets[peerHash];!ok{
		return []byte(""),errors.New("the peer hasn't negotiate the share secret, and please restart this node")
	}
	key := ghsm.secrets[peerHash]
	encData,err  := ghsm.smc.EncryptData(string(message),string(key))
	return []byte(encData),err
}
func (ghsm *GuomiHandShakeManager)DecWithSecret(message []byte, peerHash string) ([]byte,error){
	if _,ok := ghsm.secrets[peerHash];!ok{
		return []byte(""),errors.New("the peer hasn't negotiate the share secret, and please restart this node")
	}
	key := ghsm.secrets[peerHash]
	decData,err := ghsm.smc.DecryptData(string(message),string(key))
	return []byte(decData),err
}

func (ghsm *GuomiHandShakeManager)GetSecret(peerHash string) string{
	if sc, ok := ghsm.secrets[peerHash]; ok {
		return hex.EncodeToString(sc)
	} else {
		log.Error("无法取得相应秘钥", peerHash)
		return ""
	}
}

func (ghsm *GuomiHandShakeManager) GetSceretPoolSize() int {
	return len(ghsm.secrets)
}

func (ghsm *GuomiHandShakeManager) PrintAllSecHash() {
	for hash, _ := range ghsm.secrets {
		log.Notice(hash)
	}

}

func (ghsm *GuomiHandShakeManager)GetLocalPublicKey() []byte{
	return []byte(ghsm.sessionPublicKey)
}



func (ghsm *GuomiHandShakeManager) SetSignPublicKey(pub []byte,peerHash string){
	ghsm.signPublickey[peerHash] = pub
}

func (ghsm *GuomiHandShakeManager) GetSignPublicKey(peerHash string) []byte{
	if pub, ok := ghsm.signPublickey[peerHash]; ok {
		return pub
	}else {
		log.Error("无法取得相应公钥", peerHash)
		return nil
	}
}

func (ghsm *GuomiHandShakeManager)SetIsVerified(is_verified bool,peerHash string){
	ghsm.isVerified[peerHash] = is_verified
}

func (ghsm *GuomiHandShakeManager)GetIsVerified(peerHash string) bool{
	if bol, ok := ghsm.isVerified[peerHash]; ok {
		return bol
	}else {
		log.Error("无法取得相应公钥", peerHash)
		return false
	}
}