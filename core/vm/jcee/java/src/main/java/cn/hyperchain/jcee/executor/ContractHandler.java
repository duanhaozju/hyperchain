/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.executor;

import cn.hyperchain.jcee.Handler;
import org.apache.log4j.Logger;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class ContractHandler {
    private static final Logger logger = Logger.getLogger(ContractHandler.class.getSimpleName());

    private Map<String, Handler> handlers;
    private int ledgerPort;

    public ContractHandler(int ledgerPort){
        handlers = new ConcurrentHashMap<>();
        this.ledgerPort = ledgerPort;
    }

//    public Future<Response> execute(Task task) {
//        return exec.submit(task);
//    }

    public void addHandler(String namespace) {
        Handler handler = new Handler(ledgerPort);
        handlers.put(namespace, handler);
    }

    public boolean hasHandlerForNamespace(String namespace) {
        return handlers.containsKey(namespace);
    }

    public Handler get(String namespace) {
        return handlers.get(namespace);
    }

}
