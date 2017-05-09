/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.executor;

import org.apache.log4j.Logger;

import java.util.Map;
import java.util.concurrent.*;

/**
 * Created by wangxiaoyi on 2017/5/3.
 * dispatch call by namespace
 */
public class ContractExecutor {

    private static final Logger logger = Logger.getLogger(ContractExecutor.class.getSimpleName());

    private volatile boolean close;
    private ExecutorService threadPool;
    private Map<String, Executor> executors;
    private ContractHandler contractHandler;

    public ContractExecutor(int ledgerPort) {
        close = false;
        executors = new ConcurrentHashMap<>();
        threadPool = Executors.newCachedThreadPool();
        contractHandler = new ContractHandler(ledgerPort);
    }

    public void dispatch(Caller caller) throws InterruptedException {
        String namespace = caller.getNamespace();
        if (! contractHandler.hasHandlerForNamespace(namespace)) {
            contractHandler.addHandler(namespace);
            this.addExecutor(namespace);
        }
        caller.setHandler(contractHandler.get(namespace));
        executors.get(namespace).Call(caller);
    }

    class Executor implements Runnable {

        private BlockingQueue <Caller> callers;
        private String namespace;

        public Executor(String namespace) {
            this.namespace = namespace;
            this.callers = new LinkedBlockingQueue<>();
        }

        public void Call(Caller caller) throws InterruptedException{
            callers.put(caller);
        }

        @Override
        public void run() {
            while (!close) {
                try {
                    Caller caller = callers.take();
                    caller.Call();
                }catch (InterruptedException ie) {
                    logger.error(ie);
                }
            }
        }
    }

    public ContractHandler getContractHandler() {
        return contractHandler;
    }

    public void addExecutor(String namespace) {
        Executor executor = new Executor(namespace);
        this.executors.put(namespace, executor);
        this.threadPool.submit(executor);
    }

    public void close() {
        this.close = true;
        this.threadPool.shutdown();
    }
}
