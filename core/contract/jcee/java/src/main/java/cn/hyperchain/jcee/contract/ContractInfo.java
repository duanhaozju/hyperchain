/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract;

import java.util.Date;

//Contract represent a contract instance
public class ContractInfo {
    private String name;
    private String id;
    private String owner;
    private String contractPath;
    private long createTime;
    private long modifyTime;

    public ContractInfo(String name, String id, String owner) {
        this.name = name;
        this.id = id;
        this.owner = owner;
        this.createTime = new Date().getTime();
        this.modifyTime = this.createTime;
    }

    public String getName() {
        return name;
    }

    public String getId() {
        return id;
    }

    public String getOwner() {
        return owner;
    }

    public long getCreateTime() {
        return createTime;
    }

    public long getModifyTime() {
        return modifyTime;
    }

    public String getContractPath() {
        return contractPath;
    }

    public void setContractPath(String contractPath) {
        this.contractPath = contractPath;
    }

    @Override
    public String toString() {
        return "Contract{" +
                "name='" + name + '\'' +
                ", id='" + id + '\'' +
                ", owner='" + owner + '\'' +
                ", createTime=" + createTime +
                ", modifyTime=" + modifyTime +
                '}';
    }
}
