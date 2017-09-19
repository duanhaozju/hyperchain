package guomi

/*
#cgo LDFLAGS: -L/usr/local/lib -L/usr/lib -L/usr/lib -lssl -lcrypto
#cgo CFLAGS: -DOPENSSL_NO_EC2M=1
#include <stdlib.h>
#include "./crypto/sm4/sm4.h"
#include "./crypto/sm4/sm4.c"

*/
import "C"

import (
	"unsafe"
)

//var encLock sync.Mutex
//var decLock sync.Mutex
func Sm4Enc(key, src []byte) ([]byte, error) {
	//encLock.Lock()
	var ctxSm4 C.sm4_context
	//确保没有0出现
	//hex := common.ToHex(src)
	//src = []byte(hex)

	C.sm4_setkey_enc(&ctxSm4, (*C.uchar)(unsafe.Pointer(&key[0])))
	var length int
	length = len(src)
	output := make([]byte, length)
	//fmt.Println("发出的原文长度：",len(src))
	//fmt.Println("发出的原文：",src)
	C.sm4_crypt_ecb(&ctxSm4, 1, C.int(len(src)), (*C.uchar)(unsafe.Pointer(&src[0])), (*C.uchar)(unsafe.Pointer(&output[0])))
	//fmt.Println("发出的密文长度：",length)
	//fmt.Println("发出的密文：",output[:length])
	//encLock.Unlock()
	return output[:length], nil
}

func Sm4Dec(key, src []byte) ([]byte, error) {
	//decLock.Lock()
	var ctxSm4 C.sm4_context
	//fmt.Println("收到的密文长度：",len(src))
	//fmt.Println("收到的密文：",src)

	C.sm4_setkey_dec(&ctxSm4, (*C.uchar)(unsafe.Pointer(&key[0])))
	output := make([]byte, len(src))

	C.sm4_crypt_ecb(&ctxSm4, 0, C.int(len(src)), (*C.uchar)(unsafe.Pointer(&src[0])), (*C.uchar)(unsafe.Pointer(&output[0])))

	//fmt.Println("收到原文长度:",msgLenth)
	//fmt.Println("收到的原文：",output[:msgLenth])

	//还原原文
	//hex := string(srcByte)
	//dst := common.Hex2Bytes(hex)
	//decLock.Unlock()
	return output, nil
}
