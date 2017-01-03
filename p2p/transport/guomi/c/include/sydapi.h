/*==================================================================
 *Copyright(C),2016,Sunyard Tech Co.,Ltd,All rights reversed
 *File name:		sydapi.h
 *Author:			Ren Qiu'an 
 *Version:			1.0
 *Date:				2016.12.20
 *Description:		cmd api
 *Function List:	
 *History:			null
 *==================================================================
 */
 
#ifndef __SYDAPI_FILE__
#define __SYDAPI_FILE__

    
int SYD_Connect(char* sIp,int nPort,char* sPwdStr,int nTimeOut,int* pSocketFd);
int SYD_Disconnect(int nSocketFd);
int	SYD_SM2_GenKeyPair(
    int nSocketFd,
    unsigned char *pPriKey,
    int *pPriKeyLen,
    unsigned char *pPubKey,
    int *pPubKeyLen
);
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

int SYD_SM2_Verify(
		int nSocketFd, 
		unsigned char *pPubKey,
		int nPubKeyLen,
		int nOrgDataType,
		unsigned char* pOrgData, 
		int nOrgDataSize, 
		unsigned char* pSignData, 
		int nSignDataSize 
);

int SYD_SM3_Hash(
		int nSocketFd, 
		unsigned char *pPubKey,
		int nPubKeyLen,
		unsigned char* pOrgData, 
		int nOrgDataSize, 
		unsigned char* pHash 
);

int	SYD_SM2_GenSessionKey(
    int nSocketFd,
    unsigned char *pPubKey,
    int nPubKeyLen,
    unsigned char  *pCipherKey,
    int *pCipherKeyLen,
    unsigned char  *pSessionKey,
    int *pSessionKeyLen,
    unsigned char *pKCV
);

int	SYD_SM2_ConfirmSessionKey(
    int nSocketFd,
    unsigned char *pPriKey,
    int nPriKeyLen,
    unsigned  char *pCipherKey,
    int nCipherKeyLen,
    unsigned char  *pSessionKey,
    int *pSessionKeyLen,
    unsigned char *pKCV
);

int	SYD_SM4_Encrypt_Data(
    int nSocketFd,
    unsigned char *pSessionKey,
    unsigned char* pInData, 
    int nInDataSize, 
    unsigned char* pOutData, 
    int *pOutDataSize
);

int	SYD_SM4_Decrypt_Data(
    int nSocketFd,
    unsigned char *pSessionKey,
    unsigned char* pInData, 
    int nInDataSize, 
    unsigned char* pOutData, 
    int *pOutDataSize
);

void PrintStr(char *Title,const char *fmt,...);
int  PrintHex(char *Title,unsigned char *Data,int DataLen);
#endif