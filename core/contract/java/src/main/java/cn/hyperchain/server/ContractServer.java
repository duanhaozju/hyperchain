/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.server;


import io.grpc.Server;
import io.grpc.ServerBuilder;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;
import org.apache.log4j.PropertyConfigurator;

import java.io.IOException;

/**
 * ContractServer
 * used to communicate with hyperchain server
 */
public class ContractServer implements IServer {

    static {
        //PropertyConfigurator.configure("/Users/wangxiaoyi/codes/go/src/hyperchain/core/contract/java/src/main/java/cn/hyperchain/server/logging.properties");
        BasicConfigurator.configure();
    }

    private static final Logger LOG = Logger.getLogger(ContractServer.class);

    private final int port = 50051;
    private Server server;

    public void Start() {
        try {
            server = ServerBuilder.forPort(port)
                    .addService(new ContractGrpcServerImpl())
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
        LOG.info("start contract server ...");
        new ContractServer().Start();
    }
}
