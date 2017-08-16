/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.contract;

import cn.hyperchain.jcee.client.contract.ContractInfo;
import cn.hyperchain.jcee.client.contract.ContractTemplate;
import cn.hyperchain.jcee.client.contract.IContractHolder;

//ContractHolder contains a contract info and the contract instance
public class ContractHolder implements IContractHolder {
    private ContractInfo info;
    private ContractTemplate contract;
    private ContractClassLoader loader;

    public ContractHolder(final ContractInfo info, final ContractTemplate contract) {
        this.info = info;
        this.contract = contract;
        this.contract.setInfo(info);
    }

    public ContractHolder(final ContractInfo info, final ContractTemplate contract, ContractClassLoader loader) {
        this.info = info;
        this.contract = contract;
        this.loader = loader;
    }

    @Override
    public ContractInfo getInfo() {
        return info;
    }

    @Override
    public void setInfo(ContractInfo info) {
        this.info = info;
    }

    @Override
    public ContractTemplate getContract() {
        return contract;
    }

    @Override
    public void setContract(ContractTemplate contract) {
        this.contract = contract;
    }

}
