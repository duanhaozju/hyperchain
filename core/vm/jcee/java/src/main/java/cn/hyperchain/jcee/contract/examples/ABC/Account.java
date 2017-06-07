package cn.hyperchain.jcee.contract.examples.ABC;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Created by huhu on 2017/5/31.
 */

public class Account {
    private String accountNumber;
    private String name;
    private String ID;
    private String IDType;
    private String accountType;
    private String businessBankNum;
    private String businessBankName;
    private String addr;
    private String phoneNum;
    private String modifyTime;

    public Account(){}

    public Account(String accountNumber,String name,String ID,String IDType,String accountType,
                   String businessBankNum,String businessBankName,String addr,String phoneNum,String modifyTime){
        this.accountNumber = accountNumber;
        this.name = name;
        this.ID = ID;
        this.IDType = IDType;
        this.accountType = accountType;
        this.businessBankNum = businessBankNum;
        this.businessBankName = businessBankName;
        this.addr = addr;
        this.phoneNum = phoneNum;
        this.modifyTime = modifyTime;
    }

    public String getAccountNumber() {
        return accountNumber;
    }

    public void setAccountNumber(String accountNumber) {
        this.accountNumber = accountNumber;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getID() {
        return ID;
    }

    public void setID(String ID) {
        this.ID = ID;
    }

    public String getIDType() {
        return IDType;
    }

    public void setIDType(String IDType) {
        this.IDType = IDType;
    }

    public String getAccountType() {
        return accountType;
    }

    public void setAccountType(String accountType) {
        this.accountType = accountType;
    }

    public String getBusinessBankNum() {
        return businessBankNum;
    }

    public void setBusinessBankNum(String businessBankNum) {
        this.businessBankNum = businessBankNum;
    }

    public String getBusinessBankName() {
        return businessBankName;
    }

    public void setBusinessBankName(String businessBankName) {
        this.businessBankName = businessBankName;
    }

    public String getAddr() {
        return addr;
    }

    public void setAddr(String addr) {
        this.addr = addr;
    }

    public String getPhoneNum() {
        return phoneNum;
    }

    public void setPhoneNum(String phoneNum) {
        this.phoneNum = phoneNum;
    }

    public String getModifyTime() {
        return modifyTime;
    }

    public void setModifyTime(String modifyTime) {
        this.modifyTime = modifyTime;
    }
}

