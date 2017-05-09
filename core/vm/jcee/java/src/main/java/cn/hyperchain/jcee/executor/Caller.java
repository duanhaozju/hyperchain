package cn.hyperchain.jcee.executor;

import cn.hyperchain.jcee.util.Errors;
import cn.hyperchain.protos.ContractProto;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.Logger;

/**
 * Created by wangxiaoyi on 2017/5/3.
 */
public class Caller {

    private static final Logger logger = Logger.getLogger(Caller.class.getSimpleName());

    private Handler handler;
    private ContractProto.Request request;
    private StreamObserver<ContractProto.Response> responseObserver;

    public Caller(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver) {
        this.request = request;
        this.responseObserver = responseObserver;
    }

    public void setHandler(Handler handler) {
        this.handler = handler;
    }

    public String getNamespace () {
        return this.request.getContext().getNamespace();
    }

    public void Call() {
        logger.info("request method " + request.getMethod());
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
