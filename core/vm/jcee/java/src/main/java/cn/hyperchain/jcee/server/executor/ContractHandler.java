/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.executor;

import cn.hyperchain.jcee.client.executor.AbstractContractHandler;
import cn.hyperchain.jcee.client.executor.IHandler;
import org.apache.log4j.Logger;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class ContractHandler extends AbstractContractHandler {
    private static final Logger logger = Logger.getLogger(ContractHandler.class.getSimpleName());

    private Map<String, IHandler> handlers; // <namespace, handler>
    private int ledgerPort;

    private ContractHandler(int ledgerPort){
        handlers = new ConcurrentHashMap<>();
        this.ledgerPort = ledgerPort;
    }

    //@warn: this init method must be invoked after bootstrap.
    public synchronized static void init(int ledgerPort) {
        if (ch == null) {
            ch = new ContractHandler(ledgerPort);
        }
    }

    public void addHandler(String namespace) {
        logger.info("Add handler for namespace " + namespace);
        Handler handler = new Handler(ledgerPort);
        handlers.put(namespace, handler);
    }

    public boolean hasHandlerForNamespace(String namespace) {
        return handlers.containsKey(namespace);
    }

    public IHandler get(String namespace) {
        return handlers.get(namespace);
    }

}
