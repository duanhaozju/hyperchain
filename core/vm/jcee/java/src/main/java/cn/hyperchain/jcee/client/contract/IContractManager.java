package cn.hyperchain.jcee.client.contract;

import cn.hyperchain.jcee.server.contract.ContractHolder;
import cn.hyperchain.jcee.server.contract.ContractInfo;

public interface IContractManager {

    ContractTemplate getContract(String cid);

    ContractInfo getContractInfoByCid(String cid);

    void removeContract(String cid);

    void destroyContract(String cid);

    void addContract(ContractHolder holder);

    boolean deployContract(ContractInfo info) throws ClassNotFoundException;
}
