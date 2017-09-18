/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.common;

import cn.hyperchain.protos.ContractProto;
import com.google.protobuf.ByteString;
import io.grpc.stub.StreamObserver;

import java.nio.charset.Charset;

/**
 * Created by wangxiaoyi on 2017/5/3.
 */
public class Errors {

    /**
     * construct error msg and return to the invokeside
     * @param msg
     */
    public static void ReturnErrMsg(String msg, StreamObserver<ContractProto.Message> responseObserver) {
        ContractProto.Response response = ContractProto.Response.newBuilder()
                .setOk(false)
                .setResult(ByteString.copyFrom(msg, Charset.defaultCharset()))
                .build();
        ContractProto.Message rspMsg = ContractProto.Message.newBuilder()
                .setType(ContractProto.Message.Type.RESPONSE)
                .setPayload(response.toByteString())
                .build();
        responseObserver.onNext(rspMsg);
    }
}
