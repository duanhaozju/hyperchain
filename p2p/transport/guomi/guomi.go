package main

/*
#cgo CFLAGS : -I./include
#cgo LDFLAGS: -L./lib -lsydapi

#include "sydapi.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
*/
import "C"
import ("fmt"
	"unsafe"
	"encoding/hex"
)

func main() {
	//int SYD_Connect(char* sIp,int nPort,char* sPwdStr,int nTimeOut,int* pSocketFd);
	cSsIP := C.CString("122.224.86.107")
	defer C.free(unsafe.Pointer(cSsIP))
	cInPort := C.int(8889)
	cSsPwdStr := C.CString("passwd")
	defer C.free(unsafe.Pointer(cSsPwdStr))
	cInTimeOut := C.int(30)
	cIpSocketFd := C.int(0)

	socket := C.SYD_Connect(cSsIP,cInPort,cSsPwdStr,cInTimeOut,&cIpSocketFd)

	if socket != 0 {
		fmt.Println("连接失败")
		return
	}
	fmt.Println("连接成功")
	fmt.Println("链接句柄为",cIpSocketFd);

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

	//var cSPriKey C.uchar
	//unsigned uchar* csPrikey = (unsigned char*) malloc(256)
	// for(int i ;i<256;i++){
	//     printf("%c",*(csPrikey+i))
	// }

	cSPriKey := (*C.uchar)(C.malloc(256))
	//defer C.free(unsafe.Pointer(&cSPriKey))

	cIPriKeyLen := C.int(256)

	//var cSPubKey C.uchar
	//defer C.free(unsafe.Pointer(&cSPubKey))
	cSPubKey := (*C.uchar)(C.malloc(256))
	cIPubKeyLen := C.int(256)


	GenKeyPairResult := C.SYD_SM2_GenKeyPair(cIpSocketFd,cSPriKey,&cIPriKeyLen,cSPubKey,&cIPubKeyLen)

	if GenKeyPairResult != 0{
		fmt.Println("生成key pair 失败")
	}


	fmt.Println("生成秘钥对：")
	fmt.Println("私钥------------")
	//IMPORTANT convert unsigned char* to  char*
	prikey := C.GoString((*C.char)(unsafe.Pointer(cSPriKey)))
	fmt.Printf("私钥：%v,长度%v\n",prikey,cIPriKeyLen)

	fmt.Println("公钥------------")
	pubkey := C.GoString((*C.char)(unsafe.Pointer(cSPubKey)))
	fmt.Printf("公钥：%v,长度%v\n",pubkey,cIPubKeyLen)



	// C的指针操作
	//for i := 0; i < (int)(cIPriKeyLen); i ++ {
	//	tmp := (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cSPriKey)) + uintptr(i)))
	//	*tmp = C.char(C.int(*tmp))
	//	fmt.Printf("%c",*tmp)
	//}


	/*
	生成签名
	int SYD_SM2_Sign(
		    int nSocketFd,
		    unsigned char *pPriKey,
		    unsigned char nPriKeyLen,
		    int nOrgDataType,
		    unsigned char *pPubKey,
		    int nPubKeyLen,
		unsigned char* pOrgData,
		int nOrgDataSize,
		unsigned char* pSignData,
		int* pSignDataSize
);
	*/
	//句柄
	//cIpSocketFd

	//私钥
	//000100004935DA27518B6DF45ABC60C36D06C0021E8B3227538E419DCEB5A2995FC4E9CB0000000000000000000000000000000000000000000000000000000000000000
	my_privkey := "000100004935DA27518B6DF45ABC60C36D06C0021E8B3227538E419DCEB5A2995FC4E9CB0000000000000000000000000000000000000000000000000000000000000000"
	my_privkeyLen := 136
	my_orgDataType := 1 // 输入的数据类型,0:HASH 值,1:原始数据
	my_pubkey := "03420004523DB72CD84691890F709AF5532DDAC9BC904E64561DD63A71C4DE19ABDE2B7A0A594BE43A4C43078BA7A7170AC71C2B0944A6766AB13E31F2C392CA425CAE11"
	my_pubkeyLen := 136
	my_orgData := "origindata"
	my_orgDataSize := len(my_orgData)

	pMyPrivKey := (*C.uchar)(unsafe.Pointer(C.CString(my_privkey)))
	nMyPrivkeyLen := C.int(my_privkeyLen)
	nMyPrivkeyLen2 := (*C.uchar)(unsafe.Pointer(C.CString("136")))

	nMyorgDataType := C.int(my_orgDataType)

	pMyPubKey := (*C.uchar)(unsafe.Pointer(C.CString(my_pubkey)))
	nMyPubkeyLen :=C.int(my_pubkeyLen)

	pMyOrgData := (*C.uchar)(unsafe.Pointer(C.CString(my_orgData)))
	nMyOrgDataSize := C.int(my_orgDataSize)

	pSignData := (*C.uchar)(C.malloc(512))
	pSignDataSize := C.int(512)

	SignRet := C.SYD_SM2_Sign(cIpSocketFd,pMyPrivKey,*nMyPrivkeyLen2,nMyorgDataType,pMyPubKey,nMyPubkeyLen,pMyOrgData,nMyOrgDataSize,pSignData,&pSignDataSize)

	if SignRet != 0{
		fmt.Println("签名失败")
	}

	fmt.Println("签名：")
	//IMPORTANT convert unsigned char* to  char*
	signdata := C.GoString((*C.char)(unsafe.Pointer(pSignData)))
	fmt.Printf("签名：%v,长度%v\n",signdata,cIPriKeyLen)

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

	pHash := (*C.uchar)(C.malloc(512))
	hashRet := C.SYD_SM3_Hash(cIpSocketFd,pMyPubKey,nMyPubkeyLen,pMyOrgData,nMyOrgDataSize,pHash)
	if hashRet != 0{
		fmt.Println("hash失败")
	}
	fmt.Println("hash：")
	//IMPORTANT convert unsigned char* to  char*
	// IMPORTANT the hash length is 32 bytes
	hash := C.GoBytes(unsafe.Pointer(pHash),C.int(32))
	fmt.Printf("hash：%v,%v\n",hash,hex.EncodeToString(hash))

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

	pCipherKey := (*C.uchar)(C.malloc(256))
	pCipherKeyLen := C.int(256)

	pSessionKey := (*C.uchar)(C.malloc(256))
	pSessionKeyLen := C.int(256)

	pKCV := (*C.uchar)(C.malloc(256))

	sessionRet := C.SYD_SM2_GenSessionKey(cIpSocketFd,pMyPubKey,nMyPubkeyLen,pCipherKey,&pCipherKeyLen,pSessionKey,&pSessionKeyLen,pKCV)

	if sessionRet != 0{
		fmt.Println("sessionKey gen失败")
	}
	fmt.Println("session key：")

	session := C.GoBytes(unsafe.Pointer(pSessionKey),C.int(33))
	hexSession := hex.EncodeToString(session)
	var KCV []byte
	if hexSession[0:1] =="S"{
		KCV = C.GoBytes(unsafe.Pointer(pKCV),C.int(32))
	}else{
		KCV = C.GoBytes(unsafe.Pointer(pKCV),C.int(16))
	}


	fmt.Printf("session：%v,%v\n",session,hex.EncodeToString(session))
	fmt.Printf("KCV：%v,%v\n",pKCV,hex.EncodeToString(KCV))


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

	confirmSessionRet:= C.SYD_SM2_ConfirmSessionKey(cIpSocketFd,pMyPrivKey,nMyPrivkeyLen,pCipherKey,pCipherKeyLen,pSessionKey,&pSessionKeyLen,pKCV)


	if confirmSessionRet != 0{
		fmt.Println("confirmSessionKey gen失败")
	}
	fmt.Println("confirmSession key：")

	confirmSession := C.GoBytes(unsafe.Pointer(pSessionKey),C.int(33))
	hexConfirmSession := hex.EncodeToString(confirmSession)
	if hexConfirmSession[0:1] =="S"{
		KCV = C.GoBytes(unsafe.Pointer(pKCV),C.int(32))
	}else{
		KCV = C.GoBytes(unsafe.Pointer(pKCV),C.int(16))
	}


	fmt.Printf("confirmSession：%v,%v\n",session,hex.EncodeToString(session))
	fmt.Printf("KCV：%v,%v\n",pKCV,hex.EncodeToString(KCV))

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

	inData := C.CString("indata")
	pInData := (*C.uchar)(unsafe.Pointer(inData))
	nInDataSize := C.int(len("indata"))
	pOutData := (*C.uchar)(C.malloc(256))
	pOutDataSize :=C.int(256)

	encDataRet := C.SYD_SM4_Encrypt_Data(cIpSocketFd,pSessionKey,pInData,nInDataSize,pOutData,&pOutDataSize);

	if encDataRet != 0{
		fmt.Println("enc data 失败")
	}

	encData := C.GoBytes(unsafe.Pointer(pOutData),pOutDataSize)
	hexencData := hex.EncodeToString(encData)

	fmt.Printf("enc data：%v\n",hexencData)



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

	pOutDataAgain := (*C.uchar)(C.malloc(256))
	pOutDataSizeAgain :=C.int(256)
	decDataRet := C.SYD_SM4_Decrypt_Data(cIpSocketFd,pSessionKey,pOutData,nInDataSize,pOutDataAgain,&pOutDataSizeAgain);

	if decDataRet != 0{
		fmt.Println("dec data 失败")
	}

	decData := C.GoBytes(unsafe.Pointer(pOutDataAgain),pOutDataSizeAgain)
	hexdecData := hex.EncodeToString(decData)

	fmt.Printf("enc data：%v\n",hexdecData)
}
