/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee;

import cn.hyperchain.jcee.contract.ContractManager;
import cn.hyperchain.jcee.executor.*;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.RequestContext;
import cn.hyperchain.protos.Response;
import com.google.protobuf.ByteString;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.Logger;

import java.util.concurrent.Future;

/**
 * Handler used to handle the real request
 */
public class Handler {

    private Logger logger = Logger.getLogger(Handler.class.getSimpleName());
    private ContractManager cm;
    private ContractExecutor executor;

    public Handler(){
        cm = new ContractManager();
        executor = new ContractExecutor();
    }

    public void query(Request request, StreamObserver<Response> responseObserver){
        Task task = new QueryTask(cm.getContract(request.getContext().getCid()), request, constructContext(request.getContext()));
        Future<Response> future = executor.execute(task);
        Response response = null;
        try{
            response = future.get();
        }catch (Exception e) {
            logger.error(e);
            response = Response.newBuilder().setOk(false)
                    //.setId(request.getTxid())
                    .setResult(ByteString.copyFromUtf8(e.getMessage()))
                    .build();
        }finally {
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
    }

    public void invoke(Request request, StreamObserver<Response> responseObserver){
        logger.error("cid is " + request.getContext().getCid());
        logger.error("contract is " + cm.getContract(request.getContext().getCid()));
        Task task = new InvokeTask(cm.getContract(request.getContext().getCid()), request, constructContext(request.getContext()));
        Future<Response> future = executor.execute(task);
        Response response = null;
        try{
            response = future.get();
        }catch (Exception e) {
            logger.error(e);
            e.printStackTrace();
            response = Response.newBuilder().setOk(false)
                    //.setId(request.getTxid())
                    .setResult(ByteString.copyFromUtf8(e.getMessage()))
                    .build();
        }finally {
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
    }

    public void deploy(Request request, StreamObserver<Response> responseObserver){
        //TODO: deploy request
        Response r = Response.newBuilder().setOk(true).build();
        responseObserver.onNext(r);
        responseObserver.onCompleted();

    }

    public ContractManager getContractMgr() {
        return cm;
    }

    public Context constructContext(RequestContext context) {
        Context ct = new Context(context.getTxid());
        ct.setRequestContext(context);
        return ct;
    }
}
