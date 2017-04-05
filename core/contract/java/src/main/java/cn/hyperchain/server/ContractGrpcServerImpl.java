/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.server;

import cn.hyperchain.protos.ContractGrpc;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.Response;
import io.grpc.stub.StreamObserver;


public class ContractGrpcServerImpl extends ContractGrpc.ContractImplBase {
    /**
     * @param responseObserver
     */
    @Override
    public StreamObserver<Request> execute(StreamObserver<Response> responseObserver) {

        Response r = Response.newBuilder().setOk(true).build();
        responseObserver.onNext(r);
        responseObserver.onCompleted();
        return super.execute(responseObserver);
    }

    /**
     * @param request
     * @param responseObserver
     */
    @Override
    public void heartBeat(Request request, StreamObserver<Response> responseObserver) {
        Response r = Response.newBuilder().setOk(true).build();
        responseObserver.onNext(r);
        responseObserver.onCompleted();
    }
}
