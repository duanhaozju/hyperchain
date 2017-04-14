/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract;

import org.apache.log4j.Logger;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * ContractManager manage the contract load and fetch
 */
public class ContractManager {

    private static Logger logger = Logger.getLogger(ContractManager.class.getSimpleName());

    private Map<String, ContractHolder> contracts;

    public ContractManager(){
        contracts = new ConcurrentHashMap<String, ContractHolder>();
    }

    public ContractHolder getContractHolder(String cid) {
        return contracts.get(cid);
    }

    public ContractBase getContract(String cid) {
        ContractHolder holder = contracts.get(cid);
        if(holder == null) return null;
        return holder.getContract();
    }

    public void removeContract(){
        //TODO: 1. remove instance of this contract
    }

    public void destroyContract(){
        //TODO: remove and unload the class from jvm
    }

    //TODO: invoke after contract deploy, load related class into jvm
    public void addContract(ContractHolder holder) {
        String key = holder.getInfo().getId();
        if(contracts.containsKey(key)) {
            logger.error(key + "existed!");
        }else {
            logger.info("register contract with id: " + key);
            contracts.put(key, holder);
        }
    }

    public void deployContract(String contractPath, String contractName){

    }
}
