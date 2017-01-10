package SMCrypto
/*
#cgo CFLAGS : -I./include
#cgo LDFLAGS: -L./lib -lsydapi

#include "sydapi.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
*/
import "C"
import (
	"unsafe"
	"fmt"
	"encoding/hex"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)
const (
	//缓冲区长度
	bufferLength = 136
)
// Init the log setting
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p")
}


type SMCrypto struct{
	//连接句柄
	SocketFd int
	// 连接ip
	IP string
	// 连接端口
	Port int
	//passwd
	Passwd string
	//timeout
	Timeout int

}

// NewSMCrypto return a new SMCCrypto instance
func NewSMCrypto() *SMCrypto{
	var smc SMCrypto
	smc.IP = "122.224.86.107"
	smc.Port = 8889
	smc.Passwd = "passwd"
	smc.Timeout = 30
	smc.SocketFd = 0
	return &smc
}
// Connect to the SM Server
func (smc *SMCrypto)Connect()error{
	cSsIP := C.CString(smc.IP)
	defer C.free(unsafe.Pointer(cSsIP))
	cInPort := C.int(smc.Port)
	cSsPwdStr := C.CString(smc.Passwd)
	defer C.free(unsafe.Pointer(cSsPwdStr))
	cInTimeOut := C.int(smc.Timeout)
	cIpSocketFd := C.int(smc.SocketFd)
	socket := C.SYD_Connect(cSsIP,cInPort,cSsPwdStr,cInTimeOut,&cIpSocketFd)
	if socket != 0 {
		log.Error("Connect failed,error code:", socket)
		return errors.New("Connect the SM server failed!")
	}
	log.Info("SM Connect Successful")
	log.Debug("Connect socket handler fd is ", cIpSocketFd);
	smc.SocketFd = int(cIpSocketFd)
	return nil
}

// GenKeyPair generate the  priv and pub Key Pair
func (smc *SMCrypto)GenKeyPair()(prikey string,pubkey string,err error){
	cIpSocketFd := C.int(smc.SocketFd)
	/*
	生成公私钥对
	int	SYD_SM2_GenKeyPair(
		int nSocketFd,
		unsigned char *pPriKey,
		int *pPriKeyLen,
		unsigned char *pPubKey,
		int *pPubKeyLen
	);
	*/

	if smc.SocketFd == 0{
		return "","",errors.New("smc instance hasn't connect the SM service")
	}

	cSPriKey := (*C.uchar)(C.malloc(bufferLength))
	//defer C.free(unsafe.Pointer(&cSPriKey))

	cIPriKeyLen := C.int(bufferLength)

	//var cSPubKey C.uchar
	//defer C.free(unsafe.Pointer(&cSPubKey))
	cSPubKey := (*C.uchar)(C.malloc(bufferLength))
	cIPubKeyLen := C.int(bufferLength)


	GenKeyPairResult := C.SYD_SM2_GenKeyPair(cIpSocketFd,cSPriKey,&cIPriKeyLen,cSPubKey,&cIPubKeyLen)

	if GenKeyPairResult != 0{
		log.Error("gen key pair failed ,code: ",GenKeyPairResult)
		return "","",errors.New("gen key pair failed")
	}
	//IMPORTANT convert unsigned char* to  char*
	prikey = C.GoString((*C.char)(unsafe.Pointer(cSPriKey)))
	log.Debugf("私钥：%v,长度%v\n",prikey,cIPriKeyLen)

	pubkey = C.GoString((*C.char)(unsafe.Pointer(cSPubKey)))
	log.Debug("公钥：%v,长度%v\n",pubkey,cIPubKeyLen)

	return prikey,pubkey,nil

}

// Sign sign the data by user private key
func (smc *SMCrypto)Sign(privateKey string, publicKey string, originData string, originDataType int)(string, error){
	cIpSocketFd := C.int(smc.SocketFd)
	//私钥
	//000100004935DA27518B6DF45ABC60C36D06C0021E8B3227538E419DCEB5A2995FC4E9CB0000000000000000000000000000000000000000000000000000000000000000
	//my_privkey := "000100004935DA27518B6DF45ABC60C36D06C0021E8B3227538E419DCEB5A2995FC4E9CB0000000000000000000000000000000000000000000000000000000000000000"
	//my_orgDataType := 1 // 输入的数据类型,0:HASH 值,1:原始数据
	//my_pubkey := "03420004523DB72CD84691890F709AF5532DDAC9BC904E64561DD63A71C4DE19ABDE2B7A0A594BE43A4C43078BA7A7170AC71C2B0944A6766AB13E31F2C392CA425CAE11"
	//my_orgData := "origindata"
	//my_orgDataSize := len(my_orgData)

	pMyPrivKey := (*C.uchar)(unsafe.Pointer(C.CString(privateKey)))
	nMyPrivkeyLen := C.int(len(privateKey))

	nMyorgDataType := C.int(originDataType)

	pMyPubKey := (*C.uchar)(unsafe.Pointer(C.CString(publicKey)))
	nMyPubkeyLen :=C.int(len(publicKey))

	pMyOrgData := (*C.uchar)(unsafe.Pointer(C.CString(originData)))
	nMyOrgDataSize := C.int(len(originData))

	pSignData := (*C.uchar)(C.malloc(512))
	pSignDataSize := C.int(512)

	SignRet := C.SYD_SM2_Sign(cIpSocketFd,pMyPrivKey,nMyPrivkeyLen,nMyorgDataType,pMyPubKey,nMyPubkeyLen,pMyOrgData,nMyOrgDataSize,pSignData,&pSignDataSize)

	if SignRet != 0{
		log.Error("Sign failed ,code: ", SignRet)
	}

	//IMPORTANT convert unsigned char* to  char*
	signdata := C.GoString((*C.char)(unsafe.Pointer(pSignData)))
	log.Debugf("签名：%v,长度%v\n",signdata,int(pSignDataSize))
	return signdata,nil

}

// Hash use the SM3 hash algo to hash the data
func (smc *SMCrypto)Hash(publicKey string, originData string)([]byte,error){
	cIpSocketFd := C.int(smc.SocketFd)
	/*
	生成hash
	int SYD_SM3_Hash(
		int nSocketFd,
		unsigned char *pPubKey,
		int nPubKeyLen,
		unsigned char* pOrgData,
		int nOrgDataSize,
		unsigned char* pHash
	);
	 */
	pMyPubKey := (*C.uchar)(unsafe.Pointer(C.CString(publicKey)))
	nMyPubkeyLen :=C.int(len(publicKey))

	pMyOrgData := (*C.uchar)(unsafe.Pointer(C.CString(originData)))
	nMyOrgDataSize := C.int(len(originData))

	pHash := (*C.uchar)(C.malloc(512))
	hashRet := C.SYD_SM3_Hash(cIpSocketFd,pMyPubKey,nMyPubkeyLen,pMyOrgData,nMyOrgDataSize,pHash)
	if hashRet != 0{
		log.Error("hash failed, code ",hashRet)
		return nil,errors.New("hash failed")
	}
	//IMPORTANT convert unsigned char* to  char*
	// IMPORTANT the hash length is 32 bytes
	hash := C.GoBytes(unsafe.Pointer(pHash),C.int(32))
	log.Debugf("hash：%v,%v\n",hash,hex.EncodeToString(hash))
	return hash,nil
}

// GenSessionKey generate the session key
func (smc *SMCrypto)GenSessionKey(publicKey string)(sessionKey []byte,KCV []byte,err error){
	cIpSocketFd := C.int(smc.SocketFd)
	/*
	生成会话秘钥
	int SYD_SM2_GenSessionKey(
		int nSocketFd,
		unsigned char *pPubKey,
		int nPubKeyLen,
		unsigned char *pCipherKey,
		int *pCipherKeyLen,
		unsigned char *pSessionKey,
		int *pSessionKeyLen,
		unsigned char *pKCV
		);
	*/

	pMyPubKey := (*C.uchar)(unsafe.Pointer(C.CString(publicKey)))
	nMyPubkeyLen :=C.int(len(publicKey))

	pCipherKey := (*C.uchar)(C.malloc(256))
	pCipherKeyLen := C.int(256)

	pSessionKey := (*C.uchar)(C.malloc(256))
	pSessionKeyLen := C.int(256)

	pKCV := (*C.uchar)(C.malloc(256))

	sessionRet := C.SYD_SM2_GenSessionKey(cIpSocketFd,pMyPubKey,nMyPubkeyLen,pCipherKey,&pCipherKeyLen,pSessionKey,&pSessionKeyLen,pKCV)

	if sessionRet != 0{
		log.Debugf("sessionKey gen failed,code ",sessionRet)
		return nil,nil,errors.New("sessionKey gen failed")
	}
	fmt.Println("session key：")

	session := C.GoBytes(unsafe.Pointer(pSessionKey),C.int(33))
	hexSession := hex.EncodeToString(session)
	if hexSession[0:1] =="S"{
		KCV = C.GoBytes(unsafe.Pointer(pKCV),C.int(32))
	}else{
		KCV = C.GoBytes(unsafe.Pointer(pKCV),C.int(16))
	}
	log.Debugf("session：%v,%v\n",session,hex.EncodeToString(session))
	log.Debugf("KCV：%v,%v\n",pKCV,hex.EncodeToString(KCV))
	return session,KCV,nil
}

// ConfirmSessionKey confirm the session key
func (smc *SMCrypto)ConfirmSessionKey(privateKey string)(comfirmSessionKey []byte,KCV []byte,err error){
	cIpSocketFd := C.int(smc.SocketFd)
	/*
	确认会话秘钥
	int SYD_SM2_ConfirmSessionKey(
		int nSocketFd,
		unsigned char *pPriKey,
		int nPriKeyLen,
		unsigned char *pCipherKey,
		int nCipherKeyLen,
		unsigned char *pSessionKey,
		int *pSessionKeyLen,
		unsigned char *pKCV
		);
	*/
	pMyPrivKey := (*C.uchar)(unsafe.Pointer(C.CString(privateKey)))
	nMyPrivkeyLen := C.int(len(privateKey))

	pCipherKey := (*C.uchar)(C.malloc(256))
	pCipherKeyLen := C.int(256)

	pSessionKey := (*C.uchar)(C.malloc(256))
	pSessionKeyLen := C.int(256)

	pKCV := (*C.uchar)(C.malloc(256))

	confirmSessionRet:= C.SYD_SM2_ConfirmSessionKey(cIpSocketFd,pMyPrivKey,nMyPrivkeyLen,pCipherKey,pCipherKeyLen,pSessionKey,&pSessionKeyLen,pKCV)


	if confirmSessionRet != 0{
		log.Error("confirmSessionKey gen failed,code: ",confirmSessionRet)
		return nil,nil,errors.New("confirmSessionKey gen failed")
	}

	confirmSession := C.GoBytes(unsafe.Pointer(pSessionKey),C.int(33))
	hexConfirmSession := hex.EncodeToString(confirmSession)
	if hexConfirmSession[0:1] =="S"{
		KCV = C.GoBytes(unsafe.Pointer(pKCV),C.int(32))
	}else{
		KCV = C.GoBytes(unsafe.Pointer(pKCV),C.int(16))
	}

	log.Errorf("confirmSession：%v,%v\n",confirmSession,hex.EncodeToString(confirmSession))
	log.Errorf("KCV：%v,%v\n",pKCV,hex.EncodeToString(KCV))
	return confirmSession,KCV,nil
}

// EncryptData encrypt the data
func (smc *SMCrypto)EncryptData(indata,sessionKey string)(string,error){
	cIpSocketFd := C.int(smc.SocketFd)
	/*
	int SYD_SM4_Encrypt_Data(
		    int nSocketFd,
		    unsigned char *pSessionKey,
		    unsigned char* pInData,
		    int nInDataSize,
		    unsigned char* pOutData,
		    int *pOutDataSize
		);
	*/

	inData := C.CString(indata)
	pInData := (*C.uchar)(unsafe.Pointer(inData))
	nInDataSize := C.int(len(indata))
	pOutData := (*C.uchar)(C.malloc(256))
	pOutDataSize :=C.int(256)


	pSessionKey := (*C.uchar)(unsafe.Pointer(C.CString(sessionKey)))
	//C.strncpy(pSessionKey, C.CString(indata), 256);


	//pSessionKey := (*C.uchar)(C.CString(sessionKey))

	encDataRet := C.SYD_SM4_Encrypt_Data(cIpSocketFd,pSessionKey,pInData,nInDataSize,pOutData,&pOutDataSize);

	if encDataRet != 0{
		log.Error("enc data failed,code: ",encDataRet)
		return "",errors.New("enc data failed")
	}

	encData := C.GoBytes(unsafe.Pointer(pOutData),pOutDataSize)
	hexencData := hex.EncodeToString(encData)
	log.Debugf("enc data：%v\n",hexencData)
	return hexencData,nil
}

// DecryptData　decrypt the data
func (smc *SMCrypto) DecryptData(encryptedData string,sessionKey string)(string,error){
	cIpSocketFd := C.int(smc.SocketFd)
	/*
	解密
	int SYD_SM4_Decrypt_Data(
		int nSocketFd,
		unsigned char *pSessionKey,
		unsigned char* pInData,
		int nInDataSize,
		unsigned char* pOutData,
		int *pOutDataSize,
		);

	*/
	inData := C.CString(encryptedData)
	pInData := (*C.uchar)(unsafe.Pointer(inData))
	nInDataSize := C.int(len("indata"))
	pOutData := (*C.uchar)(C.malloc(256))
	pOutDataSize :=C.int(256)

	pSessionKey := (*C.uchar)(unsafe.Pointer(C.CString(sessionKey)))
	//pSessionKey := (*C.uchar)(C.malloc(256))
	//C.strncpy(pSessionKey, C.CString(sessionKey), 256);

	//pSessionKey := (*C.uchar)(C.CString(sessionKey))

	pOutDataAgain := (*C.uchar)(C.malloc(256))
	pOutDataSizeAgain :=C.int(256)
	decDataRet := C.SYD_SM4_Decrypt_Data(cIpSocketFd,pSessionKey,pInData,nInDataSize,pOutData,&pOutDataSize);
	if decDataRet != 0{
		log.Error("dec data failed,code: ",decDataRet)
		return "",errors.New("dec data failed")
	}
	decData := C.GoBytes(unsafe.Pointer(pOutDataAgain),pOutDataSizeAgain)
	hexdecData := hex.EncodeToString(decData)

	log.Debugf("dec data：%v\n",hexdecData)
	return hexdecData,nil
}
