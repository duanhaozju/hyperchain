/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract;

//ContractHolder contains a contract info and the contract instance
public class ContractHolder {
    private ContractInfo info;
    private ContractBase contract;
    private ContractClassLoader loader;

    public ContractHolder(final ContractInfo info, final ContractBase contract) {
        this.info = info;
        this.contract = contract;
    }

    public ContractHolder(final ContractInfo info, final ContractBase contract, ContractClassLoader loader) {
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

    public ContractBase getContract() {
        return contract;
    }

    public void setContract(ContractBase contract) {
        this.contract = contract;
    }
}
