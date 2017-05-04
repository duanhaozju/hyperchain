/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee;

import cn.hyperchain.jcee.contract.ContractInfo;
import cn.hyperchain.jcee.db.ContractsMeta;
import cn.hyperchain.jcee.db.MetaDB;
import cn.hyperchain.jcee.executor.Caller;
import cn.hyperchain.jcee.executor.ContractExecutor;
import cn.hyperchain.jcee.util.Errors;
import cn.hyperchain.protos.ContractGrpc;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.RequestContext;
import cn.hyperchain.protos.Response;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.Logger;

import java.util.Map;

public class ContractGrpcServerImpl extends ContractGrpc.ContractImplBase {
    private final String global = "global";
    private ContractExecutor contractExecutor;
    private MetaDB metaDB;

    private static final Logger logger = Logger.getLogger(ContractGrpcServerImpl.class.getSimpleName());

    public ContractGrpcServerImpl(int ledgerPort) {
        contractExecutor = new ContractExecutor(ledgerPort);
    }

    public void init() {
        metaDB = MetaDB.getDb();
        if (metaDB != null) {
            recovery();
        }
    }

    public void recovery() {
        // reload pre-deployed contracts
        ContractsMeta meta = metaDB.load();
        if (meta != null) {
            Map<String, Map<String, ContractInfo>> infoMap = meta.getContractInfo();
            for (Map.Entry<String, Map<String, ContractInfo>> entry : infoMap.entrySet()) {
                String namespace = entry.getKey();
                Map<String, ContractInfo> contractInfoMap = entry.getValue();
                contractExecutor.addExecutor(namespace);
                contractExecutor.getContractHandler().addHandler(namespace);
                for (Map.Entry<String, ContractInfo> infoEntry : contractInfoMap.entrySet()) {
                    ContractInfo info = infoEntry.getValue();
                    boolean rs = contractExecutor.getContractHandler().get(namespace).deploy(info);
                    if (rs == false) {
                        logger.error("reload contract for " + infoEntry.getValue().getCid() + " not success");
                    }
                }
            }
        }
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
