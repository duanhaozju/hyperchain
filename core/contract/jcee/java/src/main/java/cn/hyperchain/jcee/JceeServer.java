/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee;

import cn.hyperchain.jcee.contract.ContractBase;
import cn.hyperchain.jcee.contract.ContractHolder;
import cn.hyperchain.jcee.contract.ContractInfo;
import cn.hyperchain.jcee.contract.examples.c1.MySmartContract;
import cn.hyperchain.jcee.ledger.HyperchainLedger;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import org.apache.log4j.Logger;

/**
 * JceeServer
 * used to communicate with hyperchain jcee client
 */
public class JceeServer implements IServer {
    private static final Logger LOG = Logger.getLogger(JceeServer.class);
    private final int port;
    private Server server;
    private ContractGrpcServerImpl cgsi;

    public JceeServer(){
        port = 50051;
        cgsi = new ContractGrpcServerImpl();
    }
    public void Start() {
        try {
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
        LOG.info("Stop JCEE server");
    }

    public static void main(String []args){
        LOG.info("Start JCEE server ...");
        JceeServer cs = new JceeServer();
        //TODO: fix this kind of contract add
        ContractInfo info = new ContractInfo("msc", "msc001", "Wang Xiaoyi");
        ContractBase contract = new MySmartContract();
        contract.setLedger(new HyperchainLedger());
        ContractHolder holder = new ContractHolder(info, contract);
        cs.cgsi.getHandler().getContractMgr().addContract(holder);
        cs.Start();
    }
}
