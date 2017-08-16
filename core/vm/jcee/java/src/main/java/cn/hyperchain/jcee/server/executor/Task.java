/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.executor;

import cn.hyperchain.jcee.common.Context;
import cn.hyperchain.jcee.client.contract.ContractTemplate;
import cn.hyperchain.jcee.client.ledger.AbstractLedger;
import cn.hyperchain.protos.ContractProto;
import org.apache.log4j.Logger;

import java.util.concurrent.Callable;

public abstract class Task implements Callable<ContractProto.Response> {
    private static final Logger logger = Logger.getLogger(Task.class.getSimpleName());
    protected ContractTemplate contract;
    protected ContractProto.Request request;
    protected Context context;

    public Task(ContractTemplate contract, ContractProto.Request request, Context context) {
        this.contract = contract;
        this.request = request;
        this.context = context;
    }

    public void beforeExecute() throws Exception{
        //called before execute;
        if(contract == null) {
            logger.error("contract is null");
            return;
        }
        AbstractLedger ledger = contract.getLedger();
        if(ledger != null) {
            ledger.setContext(context);
        }else {
            logger.error("no ledger found");
            throw new Exception("no ledger found");
        }
    }

    /**
     * Computes a result, or throws an exception if unable to do so.
     * @return computed result
     * @throws Exception if unable to compute a result
     */
    @Override
    public ContractProto.Response call() throws Exception {
        beforeExecute();
        ContractProto.Response response = execute();
        afterExecute();
        return response;
    }

    /**
     * execute the contract method, which should
     * be implemented by the concrete method
     * @return executed result
     */
    public abstract ContractProto.Response execute();

    public void afterExecute() throws Exception{
        //called after execute
        AbstractLedger ledger = contract.getLedger();
        if(ledger != null) {
            ledger.removeContext();
        }else {
            logger.error("no ledger found");
            throw new Exception("no ledger found");
        }
    }
}
