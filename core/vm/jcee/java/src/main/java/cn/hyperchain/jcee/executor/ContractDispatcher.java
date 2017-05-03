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
public class ContractDispatcher {

    private static final Logger logger = Logger.getLogger(ContractDispatcher.class.getSimpleName());

    private volatile boolean close;
    private ExecutorService executor;
    private Map<String, Dispatcher> dispatchers;

    public ContractDispatcher() {
        close = false;
        dispatchers = new ConcurrentHashMap<>();
        executor = Executors.newCachedThreadPool();
    }

    public void dispatch(Caller caller) throws InterruptedException {
        dispatchers.get(caller.getNamespace()).Call(caller);
    }

    class Dispatcher implements Runnable {

        private BlockingQueue <Caller> callers;
        private String namespace;

        public Dispatcher(String namespace) {
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

    public void addDispatcher(String namespace) {
        Dispatcher dispatcher = new Dispatcher(namespace);
        this.dispatchers.put(namespace, dispatcher);
        this.executor.submit(dispatcher);
    }

    public void close() {
        this.close = true;
        this.executor.shutdown();
    }
}
