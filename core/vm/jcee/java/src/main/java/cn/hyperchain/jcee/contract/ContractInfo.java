/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract;

import org.apache.log4j.Logger;

import java.util.Arrays;
import java.util.Date;

//Contract represent a contract instance
public class ContractInfo {
    private static final Logger logger = Logger.getLogger(ContractInfo.class.getSimpleName());
    private String name;
    private String contractMainName; //fully-qualified java name
    private String cid;// contract cid
    private String owner;
    private String contractPath;
    private String classPrefix;
    private String namespace;

    private String []argTypes;
    private Object[]args;

    private String codeHash; // contract classes code hash
    private long createTime;
    private long modifyTime;

    public ContractInfo() {}

    public ContractInfo(String name, String id, String owner) {
        this.name = name;
        this.cid = id;
        this.owner = owner;
        this.createTime = new Date().getTime();
        this.modifyTime = this.createTime;
    }

    public void setCid(String cid) {
        this.cid = cid;
    }

    public String getName() {
        return name;
    }

    public String getCid() {
        return cid;
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

    public String[] getArgTypes() {
        return argTypes;
    }

    public void setArgTypes(String[] argTypes) {
        this.argTypes = argTypes;
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

    public void setName(String name) {
        this.name = name;
    }

    public Class[] getArgClasses() {
        Class[] argClasses = new Class[this.getArgTypes().length];
        for (int i = 0; i < argClasses.length; ++ i) {
            switch (argTypes[i]) {
                case "boolean":
                    argClasses[i] = boolean.class;
                    break;
                case "char":
                    argClasses[i] = char.class;
                    break;
                case "short":
                    argClasses[i] = short.class;
                    break;
                case "int":
                    argClasses[i] = int.class;
                    break;
                case "long":
                    argClasses[i] = long.class;
                    break;
                case "float":
                    argClasses[i] = float.class;
                    break;
                case "double":
                    argClasses[i] = double.class;
                    break;
                case "String":
                    argClasses[i] = String.class;
                    break;
                default:
                    logger.error("can not pass non string object to contract constructor");
                    return argClasses;
            }
        }
        return argClasses;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof ContractInfo)) return false;

        ContractInfo info = (ContractInfo) o;

        if (getCreateTime() != info.getCreateTime()) return false;
        if (getModifyTime() != info.getModifyTime()) return false;
        if (getName() != null ? !getName().equals(info.getName()) : info.getName() != null) return false;
        if (getContractMainName() != null ? !getContractMainName().equals(info.getContractMainName()) : info.getContractMainName() != null)
            return false;
        if (getCid() != null ? !getCid().equals(info.getCid()) : info.getCid() != null) return false;
        if (getOwner() != null ? !getOwner().equals(info.getOwner()) : info.getOwner() != null) return false;
        if (getContractPath() != null ? !getContractPath().equals(info.getContractPath()) : info.getContractPath() != null)
            return false;
        if (getClassPrefix() != null ? !getClassPrefix().equals(info.getClassPrefix()) : info.getClassPrefix() != null)
            return false;
        if (getNamespace() != null ? !getNamespace().equals(info.getNamespace()) : info.getNamespace() != null)
            return false;
        // Probably incorrect - comparing Object[] arrays with Arrays.equals
        if (!Arrays.equals(getArgTypes(), info.getArgTypes())) return false;
        // Probably incorrect - comparing Object[] arrays with Arrays.equals
        if (!Arrays.equals(getArgs(), info.getArgs())) return false;
        return getCodeHash() != null ? getCodeHash().equals(info.getCodeHash()) : info.getCodeHash() == null;
    }

    @Override
    public int hashCode() {
        int result = getName() != null ? getName().hashCode() : 0;
        result = 31 * result + (getContractMainName() != null ? getContractMainName().hashCode() : 0);
        result = 31 * result + (getCid() != null ? getCid().hashCode() : 0);
        result = 31 * result + (getOwner() != null ? getOwner().hashCode() : 0);
        result = 31 * result + (getContractPath() != null ? getContractPath().hashCode() : 0);
        result = 31 * result + (getClassPrefix() != null ? getClassPrefix().hashCode() : 0);
        result = 31 * result + (getNamespace() != null ? getNamespace().hashCode() : 0);
        result = 31 * result + Arrays.hashCode(getArgTypes());
        result = 31 * result + Arrays.hashCode(getArgs());
        result = 31 * result + (getCodeHash() != null ? getCodeHash().hashCode() : 0);
        result = 31 * result + (int) (getCreateTime() ^ (getCreateTime() >>> 32));
        result = 31 * result + (int) (getModifyTime() ^ (getModifyTime() >>> 32));
        return result;
    }

    @Override
    public String toString() {
        return "ContractInfo{" +
                "name='" + name + '\'' +
                ", contractMainName='" + contractMainName + '\'' +
                ", cid='" + cid + '\'' +
                ", owner='" + owner + '\'' +
                ", contractPath='" + contractPath + '\'' +
                ", classPrefix='" + classPrefix + '\'' +
                ", namespace='" + namespace + '\'' +
                ", argTypes=" + Arrays.toString(argTypes) +
                ", args=" + Arrays.toString(args) +
                ", codeHash='" + codeHash + '\'' +
                ", createTime=" + createTime +
                ", modifyTime=" + modifyTime +
                '}';
    }
}
