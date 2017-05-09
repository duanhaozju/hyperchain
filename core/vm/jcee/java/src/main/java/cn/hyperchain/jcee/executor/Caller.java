package cn.hyperchain.jcee.executor;

import cn.hyperchain.protos.ContractProto;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.Logger;

/**
 * Created by wangxiaoyi on 2017/5/3.
 */
public class Caller {

    private final String invoke = "invoke";
    private final String deploy = "deploy";

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
        switch (request.getMethod()) {
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
}
