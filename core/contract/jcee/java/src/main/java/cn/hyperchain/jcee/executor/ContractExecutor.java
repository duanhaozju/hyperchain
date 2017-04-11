/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.executor;

import cn.hyperchain.protos.Response;
import org.apache.log4j.Logger;

import java.util.concurrent.Callable;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

public class ContractExecutor {
    private static final Logger logger = Logger.getLogger(ContractExecutor.class.getSimpleName());

    private ExecutorService exec;

    public ContractExecutor(){
        exec = Executors.newFixedThreadPool(2 * Runtime.getRuntime().availableProcessors());
    }

    public Future<Response> execute(Task task) {
        logger.info(task.toString());
        logger.info(task instanceof Callable);
        return exec.submit(task);
    }
}
