/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.client.mock;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.server.contract.ContractInfo;
import cn.hyperchain.jcee.client.contract.ContractTemplate;
import cn.hyperchain.jcee.client.ledger.AbstractLedger;
import cn.hyperchain.jcee.server.common.HashFunction;
import lombok.Getter;
import lombok.Setter;

import java.util.HashMap;
import java.util.List;

/**
 * Created by huhu on 2017/6/21.
 */
public class MockServer {
    private HashMap<String,ContractTemplate> contractHolder = new HashMap<>();
    private AbstractLedger ledger;
    @Getter @Setter
    private String cid;

    public MockServer(){
        ledger = new MockLedger();
    }
    public String deploy(ContractTemplate ct){
        AbstractLedger ledger = new MockLedger();
        String cid = HashFunction.computeCodeHash(ct.toString().getBytes());
        ct.setLedger(ledger);

        ContractInfo cif = new ContractInfo();
        cif.setNamespace("global");
        ct.setInfo(cif);

        contractHolder.put(cid, ct);
        return cid;
    }
    public ExecuteResult invoke(String funcName, List<String> args){
        ContractTemplate ct = contractHolder.get(cid);
        return ct.invoke(funcName,args);
    }
}
