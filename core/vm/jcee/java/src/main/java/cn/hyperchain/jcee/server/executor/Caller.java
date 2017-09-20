/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.executor;

import cn.hyperchain.jcee.client.executor.IHandler;
import cn.hyperchain.jcee.server.common.Errors;
import cn.hyperchain.protos.ContractProto;
import io.grpc.stub.StreamObserver;
import lombok.Getter;
import org.apache.log4j.Logger;

/**
 * Created by wangxiaoyi on 2017/5/3.
 */
public class Caller {

    private static final Logger logger = Logger.getLogger(Caller.class.getSimpleName());

    private IHandler handler;
    @Getter
    private ContractProto.Request request;
    @Getter
    private StreamObserver<ContractProto.Message> responseObserver;

    public Caller(ContractProto.Request request, StreamObserver<ContractProto.Message> responseObserver) {
        this.request = request;
        this.responseObserver = responseObserver;
    }

    public void setHandler(IHandler handler) {
        this.handler = handler;
    }

    public String getNamespace () {
        return this.request.getContext().getNamespace();
    }

    public void Call() {
        logger.debug("request method " + request.getMethod());
        try {
            switch (CallType.valueOf(request.getMethod())) {
                case invoke: {
                    handler.invoke(request, responseObserver);
                    break;
                }
                case deploy: {
                    handler.deploy(request, responseObserver);
                    break;
                }
                case freeze: {
                    handler.freeze(request, responseObserver);
                    break;
                }

                case unfreeze: {
                    handler.unfreeze(request, responseObserver);
                    break;
                }

                case destroy: {
                    handler.destroy(request, responseObserver);
                    break;
                }

                case update: {
                    handler.update(request, responseObserver);
                    break;
                }
            }
        }catch (IllegalArgumentException iae){
            String msg = "method " + request.getMethod() + " is not implemented yet!";
            logger.error(iae.getMessage());
            Errors.ReturnErrMsg(msg, responseObserver);
        }catch (Exception e) {
            String msg = "method " + request.getMethod() + " execute failed " + e.getMessage();
            logger.error(msg);
            Errors.ReturnErrMsg(msg, responseObserver);
        }
    }

    enum CallType {
        deploy,
        invoke,
        update,
        freeze,
        unfreeze,
        destroy
    }
}
