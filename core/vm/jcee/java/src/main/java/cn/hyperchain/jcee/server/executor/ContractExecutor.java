/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.executor;

import cn.hyperchain.jcee.client.executor.AbstractContractHandler;
import cn.hyperchain.jcee.server.common.Errors;
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
    private AbstractContractHandler contractHandler;

    public ContractExecutor() {
        close = false;
        executors = new ConcurrentHashMap<>();
        threadPool = Executors.newCachedThreadPool();
        contractHandler = AbstractContractHandler.getContractHandler();
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
            ExecutorService executor = Executors.newSingleThreadExecutor();

            while (!close) {

                Caller caller = null;
                FutureTask<String> futureTask = null;
                try {
                    caller = callers.take();
                    if(executor.isShutdown()){
                        executor = Executors.newSingleThreadExecutor();
                        logger.info("new executor after shutdown");

                    }
                    Caller finalCaller = caller;
                    futureTask =
                            new FutureTask<String>(new Callable<String>() {//使用Callable接口作为构造参数
                                public String call() {
                                    finalCaller.Call();
                                    return "finish call";
                                }});
                    executor.execute(futureTask);

                    String result = futureTask.get(4000, TimeUnit.MILLISECONDS);
//                    logger.info("Current call result:"+result);
                }catch (TimeoutException e) {
                    logger.error("Current call result :Time out");
                    futureTask.cancel(true);
                    if(caller == null){
                        logger.error("caller is null");
                        continue;
                    }
                    Errors.ReturnErrMsg(e.toString(),caller.getResponseObserver());

                }catch (Exception e){
                    futureTask.cancel(true);
                    if(caller == null){
                        logger.error("caller is null");
                        continue;
                    }
                    Errors.ReturnErrMsg(e.toString(),caller.getResponseObserver());
                }finally {
                    if(futureTask.isCancelled()){
                        executor.shutdown();
                        logger.info("executor shutdown");

                    }
                }
            }
        }
    }

    public AbstractContractHandler getContractHandler() {
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
