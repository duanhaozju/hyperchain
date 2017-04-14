/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract;

import cn.hyperchain.jcee.ledger.AbstractLedger;
import com.google.protobuf.ByteString;
import java.util.List;

//ContractBase which is used as a skeleton of smart contract
public abstract class ContractBase {

    private String owner;

    protected AbstractLedger ledger;

    public void setLedger(AbstractLedger ledger) {
        this.ledger = ledger;
    }

    public AbstractLedger getLedger() {
        return ledger;
    }

    public void setOwner(String owner) {
        this.owner = owner;
    }

    public String getOwner() {
        return owner;
    }

    /**
     * Invoke smart contract method
     * @param funcName function name user defined in contract
     * @param args arguments of funcName
     */
    public abstract boolean Invoke(String funcName, List<String> args);

    /**
     * Query data stored in the smart contract
     * @param funcName function name
     * @param args function related arguments
     * @return the query result
     */
    public abstract ByteString Query(String funcName, List<String> args);

}
