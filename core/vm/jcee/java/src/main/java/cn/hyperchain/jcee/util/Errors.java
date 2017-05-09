package cn.hyperchain.jcee.util;

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
    public static void ReturnErrMsg(String msg, StreamObserver<ContractProto.Response> responseObserver) {
        ContractProto.Response response = ContractProto.Response.newBuilder()
                .setOk(false)
                .setResult(ByteString.copyFrom(msg, Charset.defaultCharset()))
                .build();
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }
}
