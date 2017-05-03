package cn.hyperchain.jcee.executor;

import cn.hyperchain.jcee.Handler;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.Response;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.Logger;

/**
 * Created by wangxiaoyi on 2017/5/3.
 */
public class Caller {

    private final String query = "query";
    private final String invoke = "invoke";
    private final String deploy = "deploy";

    private static final Logger logger = Logger.getLogger(Caller.class.getSimpleName());

    private Handler handler;
    private Request request;
    private StreamObserver<Response> responseObserver;

    public Caller(Handler handler, Request request, StreamObserver<Response> responseObserver) {
        this.handler = handler;
        this.request = request;
        this.responseObserver = responseObserver;
    }

    public String getNamespace () {
        return this.request.getContext().getNamespace();
    }

    public void Call() {
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
}
