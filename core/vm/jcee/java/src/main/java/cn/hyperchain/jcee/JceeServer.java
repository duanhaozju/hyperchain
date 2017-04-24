/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee;

import cn.hyperchain.jcee.contract.ContractBase;
import cn.hyperchain.jcee.contract.ContractHolder;
import cn.hyperchain.jcee.contract.ContractInfo;
import cn.hyperchain.jcee.contract.examples.c1.MySmartContract;
import cn.hyperchain.jcee.contract.examples.sb.SimulateBank;
import cn.hyperchain.jcee.ledger.HyperchainLedger;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import org.apache.log4j.Logger;
import org.apache.log4j.spi.NOPLogger;

/**
 * JceeServer
 * used to communicate with hyperchain jcee client
 */
public class JceeServer implements IServer {
    private static final Logger LOG = Logger.getLogger(JceeServer.class);
    private int port;
    private Server server;
    private ContractGrpcServerImpl cgsi;

    public JceeServer(){
        // port = 50051;
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
        final int localPorts[] = new int[] {50081, 50082, 50083, 50084};
        final int ledgerPorts[] = new int[] {50051, 50052, 50053, 50054};

        for(int i = 0; i < localPorts.length; ++ i){
            final int k = i;
            new Thread(new Runnable() {
                @Override
                public void run() {
                    while (true) {
                        LOG.info("Start JCEE server ...");
                        JceeServer cs = new JceeServer();
                        cs.port = localPorts[k];
                        //TODO: fix this kind of contract add
                        ContractInfo info = new ContractInfo("msc", "e81e714395549ba939403c7634172de21367f8b5", "Wang Xiaoyi");
                        ContractBase contract = new SimulateBank("bank001", 001, true);
                        contract.setOwner(info.getOwner());
                        contract.setLedger(new HyperchainLedger(ledgerPorts[k]));
                        ContractHolder holder = new ContractHolder(info, contract);
                        cs.cgsi.getHandler().getContractMgr().addContract(holder);
                        cs.Start();
                    }
                }
            }).start();
        }
    }
}
