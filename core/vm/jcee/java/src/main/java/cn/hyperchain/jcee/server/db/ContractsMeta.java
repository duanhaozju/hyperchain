/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.db;

import cn.hyperchain.jcee.client.contract.ContractInfo;
import lombok.Getter;
import lombok.Setter;

import java.util.HashMap;
import java.util.Map;

/**
 * Created by wangxiaoyi on 2017/5/4.
 */
public class ContractsMeta {

    @Getter
    @Setter
    private Map<String, Map<String, ContractInfo>> contractInfo;

    public ContractsMeta() {
        contractInfo = new HashMap<>();
    }
    /**
     * get contract info
     * @param namespace namespace
     * @param cid contract id
     * @return ContractInfo instance
     */
    public synchronized ContractInfo getContractInfo(String namespace, String cid) {
        Map<String, ContractInfo> infos = contractInfo.get(namespace);
        if (infos == null) return null;
        else return infos.get(cid);
    }

    public synchronized void addContractInfo(ContractInfo info) {
        String namespace = info.getNamespace();
        String cid = info.getCid();
        Map<String, ContractInfo> infoMap = contractInfo.get(namespace);
        if (infoMap == null) {
            infoMap = new HashMap<>();
            infoMap.put(cid, info);
            contractInfo.put(namespace, infoMap);
        }else {
            infoMap.put(cid, info);
        }
    }

    public synchronized void removeContractInfo(ContractInfo info) {
        String namespace = info.getNamespace();
        String cid = info.getCid();
        Map<String, ContractInfo> infoMap = contractInfo.get(namespace);
        infoMap.remove(cid);
        if (infoMap.size() == 0) {
            contractInfo.remove(namespace);
        }
    }

    @Override
    public String toString() {
        return "ContractsMeta{" +
                "contractInfo=" + contractInfo +
                '}';
    }
}
