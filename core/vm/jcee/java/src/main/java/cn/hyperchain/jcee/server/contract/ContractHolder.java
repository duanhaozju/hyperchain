/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.contract;

import cn.hyperchain.jcee.client.contract.ContractTemplate;

//ContractHolder contains a contract info and the contract instance
public class ContractHolder {
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

    public ContractInfo getInfo() {
        return info;
    }

    public void setInfo(ContractInfo info) {
        this.info = info;
    }

    public ContractTemplate getContract() {
        return contract;
    }

    public void setContract(ContractTemplate contract) {
        this.contract = contract;
    }
}
