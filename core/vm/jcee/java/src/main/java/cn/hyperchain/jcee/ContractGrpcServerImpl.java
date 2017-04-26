/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee;

import cn.hyperchain.protos.ContractGrpc;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.Response;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.Logger;

public class ContractGrpcServerImpl extends ContractGrpc.ContractImplBase {
    private final String query = "query";
    private final String invoke = "invoke";
    private final String deploy = "deploy";

    private static final Logger logger = Logger.getLogger(ContractGrpcServerImpl.class.getSimpleName());
    private Handler handler;

    public ContractGrpcServerImpl() {
        handler = new Handler();
    }
    /**
     * @param request
     * @param responseObserver
     */
    @Override
    public void execute(Request request, StreamObserver<Response> responseObserver) {
        switch (request.getMethod()) {
            case query:{
                handler.query(request, responseObserver);
                break;
            }
            case invoke: {
                handler.invoke(request, responseObserver);
                break;
            }
            case deploy: {
                handler.deploy(request, responseObserver);
                break;
            }
            default:
                logger.error("method " + request.getMethod() + " is not implemented yet!");
        }
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

    public Handler getHandler() {
        return handler;
    }

    public void setHandler(Handler handler) {
        this.handler = handler;
    }
}
