/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee;

import cn.hyperchain.jcee.executor.Caller;
import cn.hyperchain.jcee.executor.ContractDispatcher;
import cn.hyperchain.jcee.executor.ContractExecutor;
import cn.hyperchain.jcee.ledger.AbstractLedger;
import cn.hyperchain.jcee.ledger.HyperchainLedger;
import cn.hyperchain.jcee.util.Errors;
import cn.hyperchain.protos.ContractGrpc;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.RequestContext;
import cn.hyperchain.protos.Response;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.Logger;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class ContractGrpcServerImpl extends ContractGrpc.ContractImplBase {
    private final String global = "global";

    private Map<String, Handler> handlers;
    private ContractExecutor ce;
    private ContractDispatcher dispatcher;
    private int ledgerPort;

    private static final Logger logger = Logger.getLogger(ContractGrpcServerImpl.class.getSimpleName());

    public ContractGrpcServerImpl() {
        handlers = new ConcurrentHashMap<String, Handler>();
        ce = new ContractExecutor();
        addHandler(global);
        dispatcher = new ContractDispatcher();
        dispatcher.addDispatcher(global);
        AbstractLedger ledger = new HyperchainLedger(ledgerPort);
        getHandler(global).getContractMgr().setLedger(ledger);

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

        Handler handler = null;
        if (! handlers.containsKey(namespace)) {
            addHandler(namespace);
            dispatcher.addDispatcher(namespace);
            AbstractLedger ledger = new HyperchainLedger(ledgerPort);
            getHandler(namespace).getContractMgr().setLedger(ledger);
        }
        handler = handlers.get(namespace);
        Caller caller = new Caller(handler, request, responseObserver);
        try {
            dispatcher.dispatch(caller);
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

    public Handler getHandler(String namespace) {
        return handlers.get(namespace);
    }

    public void setHandler(String namespace, Handler handler) {
        this.handlers.put(namespace, handler);
    }

    public void addHandler(String namespace) {
        Handler handler = new Handler();
        handler.setExecutor(ce);
        handlers.put(namespace, handler);
    }

    public void setLedgerPort(int ledgerPort) {
        this.ledgerPort = ledgerPort;
    }
}
