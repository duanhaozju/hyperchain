package cn.hyperchain.jcee.db;

import cn.hyperchain.jcee.contract.ContractInfo;

import java.util.HashMap;
import java.util.Map;

/**
 * Created by wangxiaoyi on 2017/5/4.
 */
public class ContractsMeta {
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

    public Map<String, Map<String, ContractInfo>> getContractInfo() {
        return contractInfo;
    }

    public void setContractInfo(Map<String, Map<String, ContractInfo>> contractInfo) {
        this.contractInfo = contractInfo;
    }

    @Override
    public String toString() {
        return "ContractsMeta{" +
                "contractInfo=" + contractInfo +
                '}';
    }
}
