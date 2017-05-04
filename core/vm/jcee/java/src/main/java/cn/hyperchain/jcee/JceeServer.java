/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee;

import cn.hyperchain.jcee.ledger.AbstractLedger;
import cn.hyperchain.jcee.ledger.HyperchainLedger;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import org.apache.log4j.Logger;
import org.apache.log4j.PropertyConfigurator;

import java.io.FileInputStream;
import java.io.IOException;
import java.util.Properties;

/**
 * JceeServer
 * used to communicate with hyperchain jcee client
 */
public class JceeServer implements IServer {

    static {
        Properties props = new Properties();
        try {
            props.load(new FileInputStream("./hyperjvm/config/log4j.properties"));
        } catch (IOException e) {
            e.printStackTrace();
        }
        PropertyConfigurator.configure(props);
    }

    private static final Logger logger = Logger.getLogger(JceeServer.class);

    private int localPort;
    private int ledgerPort;
    private Server server;
    private ContractGrpcServerImpl cgsi;

    public JceeServer(int localPort, int ledgerPort){
        this.localPort = localPort;
        this.ledgerPort = ledgerPort;
        cgsi = new ContractGrpcServerImpl();
        AbstractLedger ledger = new HyperchainLedger(ledgerPort);
        this.cgsi.getHandler().getContractMgr().setLedger(ledger);
    }
    public void Start() {
        try {
            logger.info("ContractServer start listening on port " + localPort);
            server = ServerBuilder.forPort(localPort)
                    .addService(cgsi)
                    .build().start();
            server.awaitTermination();
        }catch (Exception e) {
            logger.error(e);
        }
    }

    public void Stop() {
        if(server != null) {
            server.shutdownNow();
        }
        logger.info("Stop JCEE server");
    }

    public static void main(String []args){
        if (args.length != 2) {
            logger.error("Invalid start args, need localPort and ledgerPort");
            System.exit(1);
        }
        int localPort = Integer.parseInt(args[0]);
        int ledgerPort = Integer.parseInt(args[1]);
        JceeServer server = new JceeServer(localPort, ledgerPort);
        server.Start();
    }
}
