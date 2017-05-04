/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee;

import cn.hyperchain.jcee.executor.Caller;
import cn.hyperchain.jcee.executor.ContractExecutor;
import cn.hyperchain.jcee.util.Errors;
import cn.hyperchain.protos.ContractGrpc;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.RequestContext;
import cn.hyperchain.protos.Response;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.Logger;

public class ContractGrpcServerImpl extends ContractGrpc.ContractImplBase {
    private final String global = "global";
    private ContractExecutor contractExecutor;

    private static final Logger logger = Logger.getLogger(ContractGrpcServerImpl.class.getSimpleName());

    public ContractGrpcServerImpl(int ledgerPort) {
        contractExecutor = new ContractExecutor(ledgerPort);
    }
    /**
     * @param request
     * @param responseObserver
     */
    @Override
    public void execute(Request request, StreamObserver<Response> responseObserver) {

        RequestContext rc = request.getContext();
        if (rc == null) {
            String err = "No request context specified!";
            logger.error(err);
            Errors.ReturnErrMsg(err, responseObserver);
        }

        String namespace = rc.getNamespace();
        if (namespace == null || namespace.length() == 0) { //TODO: validate the namespace
            String err = "No valid namespace specified!";
            logger.error(err);
            Errors.ReturnErrMsg(err, responseObserver);
        }
        Caller caller = new Caller(request, responseObserver);
        try {
            contractExecutor.dispatch(caller);

        }catch (InterruptedException ie) {
            logger.error(ie.getMessage());
            Errors.ReturnErrMsg(ie.getMessage(), responseObserver);
        }
    }

    /**
     * @param request
     * @param responseObserver
     * heartBeat method is globally shared
     */
    @Override
    public void heartBeat(Request request, StreamObserver<Response> responseObserver) {
        Response r = Response.newBuilder().setOk(true).build();
        responseObserver.onNext(r);
        responseObserver.onCompleted();
    }
}
