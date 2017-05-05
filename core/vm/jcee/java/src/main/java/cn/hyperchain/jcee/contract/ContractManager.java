/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract;

import cn.hyperchain.jcee.db.MetaDB;
import cn.hyperchain.jcee.ledger.AbstractLedger;
import cn.hyperchain.jcee.ledger.HyperchainLedger;
import org.apache.log4j.Logger;

import java.lang.reflect.Constructor;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * ContractManager manage the contract load and fetch
 * A namespace has a independent contract manager,
 */
public class ContractManager {
    private static Logger logger = Logger.getLogger(ContractManager.class.getSimpleName());
    private Map<String, ContractHolder> contracts;
    //ledger is shared by contract with same namespace
    private AbstractLedger ledger;
    public ContractManager(int ledgerPort){
        contracts = new ConcurrentHashMap<String, ContractHolder>();
        ledger = new HyperchainLedger(ledgerPort);
    }

    public ContractHolder getContractHolder(String cid) {
        return contracts.get(cid);
    }

    public ContractBase getContract(String cid) {
        logger.debug(contracts.toString());
        logger.debug("cid is " + cid);
        ContractHolder holder = contracts.get(cid);
        if(holder == null) return null;
        return holder.getContract();
    }

    public void removeContract(String cid){
        contracts.remove(cid);
    }

    public void destroyContract(){
        //TODO: remove and unload the class from jvm
    }

    public void addContract(ContractHolder holder) {
        String key = holder.getInfo().getCid();
        if(contracts.containsKey(key)) {
            logger.error(key + "existed!");
        }else {
            logger.info("register contract with id: " + key);
            contracts.put(key, holder);
        }
    }

    /**
     * deployContract deploy contract by the contract info
     * @param info contract info
     * @return status of deploy
     */
    public boolean deployContract(ContractInfo info){
        logger.debug("contract info, " + info.toString());

        ContractClassLoader classLoader = new ContractClassLoader(info.getContractPath(), info.getClassPrefix());
        ContractBase contract = null;
        try {
            Class contractClass = classLoader.load(info.getContractMainName());
            Object ins = newInstance(contractClass, info.getArgClasses(), info.getArgs());
            if (ins == null) {
                logger.error("init contract for " + info.getName() + " faield");
                return false;
            }
            contract = (ContractBase) ins;
            contract.setCid(info.getCid());
            contract.setOwner(info.getOwner());
            contract.setLedger(ledger);
        } catch (ClassNotFoundException e) {
            e.printStackTrace();
        }
        if (contract != null) {
            ContractHolder holder = new ContractHolder(info, contract);
            addContract(holder);
            MetaDB db = MetaDB.getDb();
            if (db != null) {
                db.store(holder.getInfo());
            }
            return true;
        }else {
            return false;
        }
    }

    public AbstractLedger getLedger() {
        return ledger;
    }

    public void setLedger(AbstractLedger ledger) {
        this.ledger = ledger;
    }

    public Object newInstance(Class clazz, Class[] argClasses, Object[] args) {
        try {
            if (argClasses == null || args == null || argClasses.length == 0 || args.length == 0) {
                return clazz.newInstance();
            }
            Constructor constructor = clazz.getDeclaredConstructor(argClasses);
            if(constructor != null) {
               return constructor.newInstance(args);
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }
}