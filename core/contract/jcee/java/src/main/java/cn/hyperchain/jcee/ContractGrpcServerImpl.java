/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee;

import cn.hyperchain.jcee.contract.ContractManager;
import cn.hyperchain.jcee.executor.ContractExecutor;
import cn.hyperchain.protos.ContractGrpc;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.Response;
import io.grpc.stub.StreamObserver;

public class ContractGrpcServerImpl extends ContractGrpc.ContractImplBase {
    private ContractManager cm;
    private ContractExecutor executor;

    public ContractGrpcServerImpl() {

    }
    /**
     * @param request
     * @param responseObserver
     */
    @Override
    public void execute(Request request, StreamObserver<Response> responseObserver) {
        Response r = Response.newBuilder().setOk(true).build();
        responseObserver.onNext(r);
        responseObserver.onCompleted();
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
