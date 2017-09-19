package cn.hyperchain.jcee.server.mocknet;

import cn.hyperchain.jcee.server.common.Errors;
import cn.hyperchain.protos.ContractProto;
import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.stub.StreamObserver;

public class RawNetServer {

    /**
     * @param request
     * @param responseObserver
     * heartBeat method is globally shared
     */
    public void heartBeat(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver) {
        ContractProto.Response r = ContractProto.Response.newBuilder().setOk(true).build();
        responseObserver.onNext(r);
        responseObserver.onCompleted();
    }

    public StreamObserver<ContractProto.Message> register(StreamObserver<ContractProto.Message> responseObserver) {
        return new StreamObserver<ContractProto.Message>() {
            @Override
            public void onNext(ContractProto.Message message) {

                switch (message.getType()){
                    case TRANSACTION:{
                        try {
                            ContractProto.Request request = ContractProto.Request.parseFrom(message.getPayload().toByteArray());
                            Errors.ReturnErrMsg("xxx", responseObserver);
                        }catch (InvalidProtocolBufferException ipbe){

                        }
                    }
                    case UNRECOGNIZED:{
//                        logger.error("Receive undefined transaction type");
                    }
                }
            }

            @Override
            public void onError(Throwable throwable) {

            }

            @Override
            public void onCompleted() {

            }
        };
    }

    public static void main(String []args) {

    }


}
