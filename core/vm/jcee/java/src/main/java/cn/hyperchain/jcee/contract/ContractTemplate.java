/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.ledger.AbstractLedger;
import lombok.*;

import java.util.List;

//ContractBase which is used as a skeleton of smart contract
@Data
@AllArgsConstructor
@NoArgsConstructor
public abstract class ContractTemplate {
    private ContractInfo info;
    private String owner;
    private String cid;
    protected AbstractLedger ledger;

    /**
     * invoke smart contract method
     * @param funcName function name user defined in contract
     * @param args arguments of funcName
     * @return ExecuteResult
     */
    public abstract ExecuteResult invoke(String funcName, List<String> args);

    public ExecuteResult result(boolean exeSuccess, Object result) {
        ExecuteResult rs = new ExecuteResult<>(exeSuccess, result);
        return rs;
    }

    public ExecuteResult result(boolean exeSuccess) {
        ExecuteResult rs = new ExecuteResult<>(exeSuccess, "");
        return rs;
    }
}
