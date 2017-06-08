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

import cn.hyperchain.protos.ContractProto;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.Logger;

import java.util.Map;

public class ContractGrpcServerImpl extends ContractGrpc.ContractImplBase {
    private ContractExecutor contractExecutor;
    private MetaDB metaDB;
    private volatile State state;

    private static final Logger logger = Logger.getLogger(ContractGrpcServerImpl.class.getSimpleName());

    public ContractGrpcServerImpl() {
        contractExecutor = new ContractExecutor();
        state = State.initialized;
    }

    public void init() {
        state = State.starting;
        metaDB = MetaDB.getDb();
        if (metaDB != null) {
           boolean success = recovery();
           if (!success) {
                logger.error("HyperJVM recovery failed!");
           }
           state = State.running;
        }
    }

    public boolean recovery() {
        // reload pre-deployed contracts
        // TODO: try multi_thread recovery
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
                    boolean rs = false;
                    try {
                        rs = contractExecutor.getContractHandler().get(namespace).deploy(info);
                    }catch (ClassNotFoundException e) {
                        logger.error(e);
                    }
                    if (rs == false) {
                        logger.error("reload contract for " + infoEntry.getValue().getCid() + " not success");
                    }
                }
            }
        }
        return true;
    }
    /**
     * @param request
     * @param responseObserver
     */
    @Override
    public void execute(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver) {
        if (state == State.running) {
            if (!pass(request, responseObserver)) return;
            Caller caller = new Caller(request, responseObserver);
            try {
                contractExecutor.dispatch(caller);
            } catch (InterruptedException ie) {
                logger.error(ie.getMessage());
                Errors.ReturnErrMsg(ie.getMessage(), responseObserver);
            }
        }else {
            Errors.ReturnErrMsg("HyperVM can not execute request now, current state is " + state, responseObserver);
        }
    }

    //pass validate the request header.
    public boolean pass(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver) {
        ContractProto.RequestContext rc = request.getContext();
        String err;
        if (rc == null) {
            err = "No request context specified!";
            logger.error(err);
            Errors.ReturnErrMsg(err, responseObserver);
            return false;
        }

        String namespace = rc.getNamespace();
        if (namespace == null || namespace.length() == 0) {
            //TODO: validate the existence of namespace
            err = "No valid namespace specified!";
            logger.error(err);
            Errors.ReturnErrMsg(err, responseObserver);
            return false;
        }

        if (rc.getCid() == null || rc.getCid().isEmpty()) {
            err = "No cid specified";
            logger.error(err);
            Errors.ReturnErrMsg(err, responseObserver);
            return false;
        }

        if (rc.getTxid() == null || rc.getTxid().isEmpty()) {
            err = "No cid specified";
            logger.error(err);
            Errors.ReturnErrMsg(err, responseObserver);
            return false;
        }
        return true;
    }

    /**
     * @param request
     * @param responseObserver
     * heartBeat method is globally shared
     */
    @Override
    public void heartBeat(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver) {
        ContractProto.Response r = ContractProto.Response.newBuilder().setOk(true).build();
        responseObserver.onNext(r);
        responseObserver.onCompleted();
    }

    //ContractGrpcServerImpl state
    enum State {
        initialized, starting, running, stopped
    }
}
