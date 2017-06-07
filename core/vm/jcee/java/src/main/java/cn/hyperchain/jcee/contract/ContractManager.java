/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract;

import cn.hyperchain.jcee.contract.security.ByteCodeChecker;
import cn.hyperchain.jcee.contract.security.Checker;
import cn.hyperchain.jcee.db.MetaDB;
import cn.hyperchain.jcee.ledger.AbstractLedger;
import cn.hyperchain.jcee.ledger.HyperchainLedger;
import lombok.Getter;
import lombok.Setter;
import org.apache.log4j.Logger;

import java.lang.reflect.Constructor;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * ContractManager manage the contract load and fetch
 * A namespace has a independent contract manager,
 */
public class ContractManager {
    private static Logger logger = Logger.getLogger(ContractManager.class.getSimpleName());
    private Map<String, ContractHolder> contracts;
    private List<Checker> checkers;

    //ledger is shared by contract with same namespace
    @Setter
    @Getter
    private AbstractLedger ledger;
    public ContractManager(int ledgerPort){
        contracts = new ConcurrentHashMap<>();
        ledger = new HyperchainLedger(ledgerPort);
        checkers = new LinkedList<>();
        checkers.add(new ByteCodeChecker());
    }

    public ContractTemplate getContract(String cid) {
        logger.debug(contracts.toString());
        logger.debug("cid is " + cid);
        ContractHolder holder = contracts.get(cid);
        if(holder == null) return null;
        return holder.getContract();
    }

    public ContractInfo getContractInfoByCid(String cid) {
        if (contracts.containsKey(cid)) {
            return contracts.get(cid).getInfo();
        }
        return null;
    }

    public void removeContract(String cid){
        contracts.remove(cid);
    }

    public void destroyContract(String cid){
        ContractInfo info = getContractInfoByCid(cid);
        removeContract(cid);
        MetaDB db = MetaDB.getDb();
        if (db != null) {
            db.remove(info);
        }
    }

    public boolean containsContract(String cid) {
        return contracts.containsKey(cid);
    }

    public void addContract(ContractHolder holder) {
        String key = holder.getInfo().getCid();
        if(contracts.containsKey(key)) {
            logger.error(key + " existed!");
        }else {
            logger.info("register contract with id: " + key);
            contracts.put(key, holder);
        }
    }

    public boolean isSourceSafe(String path) {
        for (Checker checker : checkers) {
            boolean isSafe = checker.passAll(path);
            if (!isSafe) {
                return false;
            }
        }
        return true;
    }

    /**
     * deployContract deploy contract by the contract info
     * @param info contract info
     * @return status of deploy
     */
    public boolean deployContract(ContractInfo info) throws ClassNotFoundException{

        if (! isSourceSafe(info.getContractPath())) {
            return false;
        }
        logger.debug("contract info, " + info.toString());
        ContractClassLoader classLoader = new ContractClassLoader(info.getContractPath(), info.getClassPrefix());
        ContractTemplate contract = null;

        Class contractClass = classLoader.load(info.getContractMainName());
        Object ins = newInstance(contractClass, info.getArgClasses(), info.getArgs());
        if (ins == null) {
            logger.error("init contract for " + info.getName() + " faield");
            return false;
        }
            contract = (ContractTemplate) ins;
            contract.setCid(info.getCid());
            contract.setOwner(info.getOwner());
            contract.setLedger(ledger);
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