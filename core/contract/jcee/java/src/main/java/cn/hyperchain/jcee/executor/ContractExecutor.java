/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.executor;


import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public class ContractExecutor {

    private ExecutorService exec;

    public ContractExecutor(){
        exec = Executors.newFixedThreadPool(2 * Runtime.getRuntime().availableProcessors());
    }

    public void execute(Task task) {
//        exec.sub
    }

}
