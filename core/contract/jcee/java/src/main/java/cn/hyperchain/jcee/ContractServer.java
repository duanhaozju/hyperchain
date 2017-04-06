/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee;

import cn.hyperchain.protos.Command;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;

/**
 * ContractServer
 * used to communicate with hyperchain jcee
 */
public class ContractServer implements IServer {

    static {
        //PropertyConfigurator.configure("/Users/wangxiaoyi/codes/go/src/hyperchain/core/contract/java/src/main/java/cn/hyperchain/jcee/logging.properties");
        BasicConfigurator.configure();
    }

    private static final Logger LOG = Logger.getLogger(ContractServer.class);

    private final int port = 50051;
    private Server server;

    public void Start() {
        try {
            final ContractGrpcServerImpl cgsi = new ContractGrpcServerImpl();
            server = ServerBuilder.forPort(port)
                    .addService(cgsi)
                    .build().start();
            server.awaitTermination();

            LOG.info("ContractServer start listening on port " + port);
        }catch (Exception e) {
            LOG.error(e);
        }
    }

    public void Stop() {
        if(server != null) {
            server.shutdownNow();
        }
    }

    public static void main(String []args){
        LOG.info("start contract jcee ...");
        ContractServer cs = new ContractServer();
        cs.Start();

    }
}
