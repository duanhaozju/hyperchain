// Package hts implements the Hyper Transport security
// include double side key agreement, and key session management
// this is the feature of hyperchian release 1.3 draft
// more details:
/*
 Hyperchain Hyper Transport Security (HTS) Draft
 Version: 1.0
 Date: 2017-07-06
 Author: Chen Quan <quan.chen@hyperchain.cn>

 1. 简介
     HTS 旨在能够提供一个安全的高层对称加密协议，能够在不同的Client-Server(Pair)之间实现信息的加密通信。
 HTS 不是为了实现TLS的相应功能，在实现上尽量简化了TLS的密钥协商步骤，总体上同TLS的思想类似，HTS是
 Hyperchain 在实现权限控制和其衍生的安全体系的需求下诞生的，通过Hyperchain自带的HTS能够实现P2P通信
 的安全性保证以及准入控制。

 2. 术语
  - Client 连接发起端
  - Server 连接监听端(被动连接端)
  - ECert Enrollment Certificate 准入证书
  - ECA Enrollment Certificate Authority 准入证书认证机构
  - KeyExchange 密钥交换
  - CipherSpec 加密细节
  - DH Diffie-Hellman key agree algorithm

 3. 细节
      一般来说，认证通常由Client发起，下面讨论的密钥协商均是建立在Hyperchain 底层TLS安全通信的基础上进行
  的。简化的HTS权限认证以及密钥协商主要包括如下步骤：
  1) ClientHello
  	Client发起连接请求，携带Certificate,ClientCipherSpec,以及ClientKeyExchangeParams,与TLS不同
  	之处在于，HTS在发起连接的时候就会携带相应的证书以及协商参数，这是因为所有的通信信道都是建立在基础可信
  	的TLS基础上的，而且在P2P网络设计中，为了避免恶意的证书请求以及的网络流量攻击要求Client先提供相应的身
  	份信息。
  2) ServerHello
  	Server收到Client发起的ClientHello请求之后，将会对Client的身份信息进行验证，这里包括对Client-
  	Certificate的合法性验证，以及Client发起的信息的来源验证，验证ClientSignature,主要是为了确认信息来
  	源的合法性以及可靠性，避免中间人攻击，这里分为两种情况：
  	1> 如果验证不通过，返回ServerReject,断开连接
  	2> 如果验证通过，则将该客户端发送过来的ClientKeyExchange,ClientCipherSpec,以及ClientSessionID
  	   缓存起来，然后向该客户端提供ServerCertificate,ServerKeyExchange,ServerCipherSpec参数，发还
  	   ServerHello响应。
  	注：这两步可以在一次request,response中完成
  3) ClientResponse
  	当Client收到Server响应的时,如果是ServerReject,断开连接;如果收到的是ServerHello消息，则进行如下处理:
  	1> 如果认证服务器信息，包括ServerCertificate,ServerSignature通过，则接受ServerKeyExchange,
  	ServerCipherSpec,并生成ClientSessionKey,用于在通信过程中进行对称加密。然后向服务器返回ClientAccept
  	2> 如果认证服务器信息失败，则发送ClientReject,断开连接.
  4) ServerDone
  	当Server收到ClientAccept的时候，就将前面缓存的信息计算SessionKey，用于在通信过程中进行对称加密。
  5) 至此，HTS 完成。

 4. 图解
	Client                                         Server
	^ClientHello
	  *ClientCertificate
	  *ClientSignature      ------------>        Listening
	  *ClientCipher
	  *ClientKeyExchange

	                                            ^ServerReject
	                                                or
	                                            ^ServerHello
	                                             *ServerCertificate
	                       <------------         *ServerSignature
	                                             *ServerCipherSpec
	                                             *ServerKeyExchange

	  ^ClientAccept
	      or               ------------->
	  ^ClientReject

	                                            ^ServerDone
                               <------------
 5. 加密细节
     在整个HTS体系中要求实现如下三种密码学算法：1)密钥交换算法 2)数字签名算法 3) 对称加密算法，按照可插拔设计，
 在HTS体系中三种算法都是可变的，目前主要可用的算法有：
     1> 密钥交换算法
       - ECDH 基于椭圆曲线的DH算法
       - SM2DH 基于国密SM2算法的DH算法
       - DH 基于RSA加密体系的DH算法
     2> 数字签名算法
       - ECDSA 基于椭圆曲线的数字签名算法
       - DSA 基于RSA的数字签名算法
       - SM2 国密SM2
     3> 对称加密算法
       - 3DES Triple Digital Encryption Standard
       - AES Advance Encryption Standard
       - SM4 国密SM4对称加密算法
 6. 实现细节
     在hyperchain p2p包中定义了一个`hts`包,用于实现HTS相关内容。
     （待补充）
*/
package hts

import (
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/p2p/hts/secimpl"
	"github.com/pkg/errors"
	"github.com/terasum/viper"
)

type HTS struct {
	sec secimpl.Security
	cg  *CertGroup
}

// NewHTS creates and returns a new HTS(Hyper Transport Security) instance.
func NewHTS(namespace string, sec secimpl.Security, caConfigPath string) (*HTS, error) {
	hts := &HTS{
		sec: sec,
		cg:  new(CertGroup),
	}

	// check ca config path
	if !common.FileExist(caConfigPath) {
		return nil, errors.New(fmt.Sprintf("CA config file %s doesn't exist, please check it.", caConfigPath))
	}

	// read in config, and get all certs
	vip := viper.New()
	vip.SetConfigFile(caConfigPath)
	err := vip.ReadInConfig()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read in the caconfig, reason: %s ", err.Error()))
	}

	if hts.cg, err = NewCertGroup(namespace, vip); err != nil {
		return nil, err
	}

	return hts, nil
}

// GetAClientHTS creates and returns a new client HTS instance.
func (hts *HTS) GetAClientHTS() (*ClientHTS, error) {
	chts, err := NewClientHTS(hts.sec, hts.cg)
	if err != nil {
		return nil, err
	}
	return chts, nil
}

// GetServerHTS creates and returns a new server HTS instance.
// Generally this function will be invoked only once in a namespace.
func (hts *HTS) GetServerHTS(peermgrEv *event.TypeMux) (*ServerHTS, error) {
	shts, err := NewServerHTS(hts.sec, hts.cg, peermgrEv)
	if err != nil {
		return nil, err
	}
	return shts, nil
}
