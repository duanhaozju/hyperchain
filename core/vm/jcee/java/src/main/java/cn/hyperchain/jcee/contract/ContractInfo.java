/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract;

import java.util.Arrays;
import java.util.Date;

//Contract represent a contract instance
public class ContractInfo {
    private String name;
    private String contractMainName; //fully-qualified java name
    private String id;
    private String owner;
    private String contractPath;
    private String classPrefix;
    private String namespace;

    private Class []argClasses;
    private Object[]args;

    private String codeHash; // contract classes code hash
    private long createTime;
    private long modifyTime;

    public ContractInfo(String name, String id, String owner) {
        this.name = name;
        this.id = id;
        this.owner = owner;
        this.createTime = new Date().getTime();
        this.modifyTime = this.createTime;
    }

    public void setId(String id) {
        this.id = id;
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

    public String getClassPrefix() {
        return classPrefix;
    }

    public void setClassPrefix(String classPrefix) {
        this.classPrefix = classPrefix;
    }

    public String getNamespace() {
        return namespace;
    }

    public void setNamespace(String namespace) {
        this.namespace = namespace;
    }

    public void setCreateTime(long createTime) {
        this.createTime = createTime;
    }

    public void setModifyTime(long modifyTime) {
        this.modifyTime = modifyTime;
    }

    public String getContractMainName() {
        return contractMainName;
    }

    public void setContractMainName(String contractMainName) {
        this.contractMainName = contractMainName;
    }

    public Class[] getArgClasses() {
        return argClasses;
    }

    public void setArgClasses(Class[] argClasses) {
        this.argClasses = argClasses;
    }

    public Object[] getArgs() {
        return args;
    }

    public void setArgs(Object[] args) {
        this.args = args;
    }

    public void setOwner(String owner) {
        this.owner = owner;
    }

    public String getCodeHash() {
        return codeHash;
    }

    public void setCodeHash(String codeHash) {
        this.codeHash = codeHash;
    }

    @Override
    public String toString() {
        return "ContractInfo{" +
                "name='" + name + '\'' +
                ", contractMainName='" + contractMainName + '\'' +
                ", id='" + id + '\'' +
                ", owner='" + owner + '\'' +
                ", contractPath='" + contractPath + '\'' +
                ", classPrefix='" + classPrefix + '\'' +
                ", namespace='" + namespace + '\'' +
                ", argClasses=" + Arrays.toString(argClasses) +
                ", args=" + Arrays.toString(args) +
                ", codeHash='" + codeHash + '\'' +
                ", createTime=" + createTime +
                ", modifyTime=" + modifyTime +
                '}';
    }
}
