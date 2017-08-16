package cn.hyperchain.jcee.client.contract;

public interface IContractHolder {

    ContractInfo getInfo();

    void setInfo(ContractInfo info);

    ContractTemplate getContract();

    void setContract(ContractTemplate contract);
}
