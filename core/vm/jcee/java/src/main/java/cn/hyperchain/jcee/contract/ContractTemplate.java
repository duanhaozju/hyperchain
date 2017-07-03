/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.common.exception.NetworkUnavalibleException;
import cn.hyperchain.jcee.executor.ContractHandler;
import cn.hyperchain.jcee.executor.Handler;
import cn.hyperchain.jcee.ledger.AbstractLedger;
import lombok.*;
import org.apache.log4j.Logger;

import java.util.*;

//ContractBase which is used as a skeleton of smart contract
@Data
@AllArgsConstructor
@NoArgsConstructor
public abstract class ContractTemplate {
    private ContractInfo info;
    private String owner;
    private String cid;
    protected AbstractLedger ledger;
    protected Logger logger = Logger.getLogger(this.getClass().getSimpleName());

    private FilterChain filterChain = new FilterChain();

    public void init(){}

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
    protected ExecuteResult openInvoke(String funcName, List<String> args) {
        return result(true);
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

        ContractInfo info = this.getInfo();
        if (ns == null || ns.isEmpty() || contractAddr == null || contractAddr.isEmpty())
            return result(false, "invalid namespace or contract address");
        ContractHandler ch = ContractHandler.getContractHandler();
        Handler handler = ch.get(ns);
        if(! ch.hasHandlerForNamespace(ns)) {
            return result(false, "no namespace named: " + ns + "found");
        }
        ContractTemplate ct = handler.getContractMgr().getContract(contractAddr);
        if (ct == null) {
            return result(false, "no contract with address: " + contractAddr + "found");
        }
        return ct.checkRuler(func, args,info);
    }

    protected ExecuteResult checkRuler(String funcName, List<String> args,ContractInfo info){

        String contractAddr = info.getCid();

        if(!filterChain.doFilter(info)){
            return result(false, "there is no authority to invoke the method "+funcName+ " for contract "+contractAddr);
        }
        return openInvoke(funcName,args);
    }

    /**
     * post event out of hyperchain
     * @param event user defined event
     * @return the post result
     * @exception
     */
    protected boolean postEvent(Event event) throws NetworkUnavalibleException {
        //TODO: add event post interface

        return true;
    }




    public ExecuteResult result(boolean exeSuccess, Object result) {
        ExecuteResult rs = new ExecuteResult<>(exeSuccess, result);
        return rs;
    }

    public ExecuteResult result(boolean exeSuccess) {
        ExecuteResult rs = new ExecuteResult<>(exeSuccess, "");
        return rs;
    }
}
