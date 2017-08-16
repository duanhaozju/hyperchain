package cn.hyperchain.jcee.client.contract;

public interface IContractManager {

    ContractTemplate getContract(String cid);

    ContractInfo getContractInfoByCid(String cid);

    void removeContract(String cid);

    void destroyContract(String cid);

    void addContract(IContractHolder holder);

    boolean deployContract(ContractInfo info) throws ClassNotFoundException;
}
