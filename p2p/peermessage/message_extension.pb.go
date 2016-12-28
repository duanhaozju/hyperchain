package peermessage

import (
	"context"
	"google.golang.org/grpc"
	//"hyperchain/crypto"
	"hyperchain/core/crypto/primitives"
)

func (c *chatClient) WrapperChat(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Message, error) {
	var pri interface{}
	var parErr error
	priStr,getErr := primitives.GetConfig("./config/cert/ecert.priv")
	if getErr == nil{
		//var parErr error
		pri,parErr = primitives.ParseKey(priStr)
	}

	ecdsaEncrypto := primitives.NewEcdsaEncrypto("ecdsa")

	if parErr == nil{
		sign,err := ecdsaEncrypto.Sign(in.Payload,pri)

		if err == nil{
			in.Signature.Signature = sign
		}
	}

	message,err := c.Chat(ctx,in,opts)
	return  message,err
}