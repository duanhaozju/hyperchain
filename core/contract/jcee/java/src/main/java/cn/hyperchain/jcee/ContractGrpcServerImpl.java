/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee;

import cn.hyperchain.jcee.contract.ContractManager;
import cn.hyperchain.jcee.executor.ContractExecutor;
import cn.hyperchain.protos.Command;
import cn.hyperchain.protos.ContractGrpc;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.Response;
import io.grpc.stub.StreamObserver;

import java.util.Map;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.LinkedBlockingQueue;


public class ContractGrpcServerImpl extends ContractGrpc.ContractImplBase {

    private StreamObserver<Command> commandObserver;
    private BlockingQueue<Response> queue;
    private BlockingQueue<Command> commands;
    private Map<String, Response> cmdrs;
    private ContractManager cm;
    private ContractExecutor executor;

    public ContractGrpcServerImpl() {
        this.queue = new LinkedBlockingQueue<Response>();
        this.commands = new LinkedBlockingQueue<Command>();
        this.cmdrs = new ConcurrentHashMap<String, Response>();
    }


    //TODO: try to design a async cmd invoke system

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

//    /**
//     * <pre>
//     * used to transfer data between hyperchain and contract
//     * TODO:this args design isn't perfect
//     * </pre>
//     *
//     * @param responseObserver
//     */
//    @Override
//    public StreamObserver<Response> dataPipeline(final StreamObserver<Command> responseObserver) {
//        this.commandObserver = responseObserver;
//        return new StreamObserver<Response>() {
//            public void onNext(Response response) {
//                queue.add(response);
//            }
//
//            public void onError(Throwable throwable) {
//                System.out.println(throwable.getMessage());
//            }
//
//            public void onCompleted() {
//               responseObserver.onCompleted();
//            }
//        };
//    }

    //executeCmd execute command remotely and wait for the result.
    public byte[] executeCmd(Command cmd) {
        this.commandObserver.onNext(cmd);
        //get result from cmd rs

        //TODO:
        return null;
    }
}
