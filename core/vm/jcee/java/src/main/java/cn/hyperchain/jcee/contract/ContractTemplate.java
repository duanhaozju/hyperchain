/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.filter.Filter;
import cn.hyperchain.jcee.contract.filter.FilterChain;
import cn.hyperchain.jcee.contract.filter.FilterManager;
import cn.hyperchain.jcee.executor.Context;
import cn.hyperchain.jcee.executor.ContractHandler;
import cn.hyperchain.jcee.executor.Handler;
import cn.hyperchain.jcee.ledger.AbstractLedger;
import lombok.*;
import org.apache.log4j.Logger;

import java.util.*;

//ContractBase which is used as a skeleton of smart contract
@AllArgsConstructor
@NoArgsConstructor
public abstract class ContractTemplate {
    @Setter
    @Getter
    private ContractInfo info;
    @Setter
    @Getter
    private String owner;
    @Setter
    @Getter
    private String cid;
    @Setter
    @Getter
    protected AbstractLedger ledger;
    protected Logger logger = Logger.getLogger(this.getClass().getSimpleName());
    protected final FilterManager filterManager = new FilterManager();

    /**
     * invoke smart contract method
     * @param funcName function name user defined in contract
     * @param args arguments of funcName
     * @return {@link ExecuteResult}
     */
    public abstract ExecuteResult invoke(String funcName, List<String> args);

    /**
     * openInvoke provide a interface to be invoked by other contracts
     * @param funcName contract name
     * @param args arguments
     * @return {@link ExecuteResult}
     */
    protected  ExecuteResult openInvoke(String funcName, List<String> args){
        //this is a empty method by default
        return result(false, "no method to invoke");
    }

    /**
     * invoke other contract method, this method is not safe enough now
     * @param ns contract namespace
     * @param contractAddr contract address
     * @param func function name
     * @param args function arguments
     * @return {@link ExecuteResult}
     */
    protected final ExecuteResult invokeContract(String ns, String contractAddr, String func, List<String> args) {
        //TODO: how to do the authority control

        // 1.check invoke identifier namespace and the contract Address
        if (ns == null || ns.isEmpty() || contractAddr == null || contractAddr.isEmpty())
            return result(false, "invalid namespace or contract address");

        ContractHandler ch = ContractHandler.getContractHandler();
        Handler handler = ch.get(ns);

        if(! ch.hasHandlerForNamespace(ns)) {
            return result(false, String.format("no namespace named: %s found", ns));
        }
        ContractTemplate ct = handler.getContractMgr().getContract(contractAddr);
        if (ct == null) {
            return result(false, String.format("no contract with address: %s found", contractAddr));
        }

        Context context = new Context(ledger.getContext().getId());
        context.setRequestContext(ledger.getContext().getRequestContext());
        return ct.securityCheck(func, args, context);
    }

    private final ExecuteResult securityCheck(String funcName, List<String> args, Context context){

        FilterChain fc = filterManager.getFilterChain(funcName);
        if (fc == null) {
            if (!fc.doFilter(context)) {
                return result(false,
                        String.format("Invoker %s at contract %s try to invoke function %s failed", context.getRequestContext().getInvoker()
                        , context.getRequestContext().getCid(), funcName));
            }
        }
        return openInvoke(funcName, args);
    }

    /**
     * execution result encapsulation method.
     * @param exeSuccess indicate the execute status, true for success, false for failed
     * @param result contract execution result if success or failed message
     * @return
     */
    protected final ExecuteResult result(boolean exeSuccess, Object result) {
        ExecuteResult rs = new ExecuteResult<>(exeSuccess, result);
        return rs;
    }

    //just indicate the execution status, no result specified
    protected final ExecuteResult result(boolean exeSuccess) {
        ExecuteResult rs = new ExecuteResult<>(exeSuccess, "");
        return rs;
    }

    protected final void addFilter(String funcName, Filter filter) {
        filterManager.AddFilter(funcName, filter);
    }
}
