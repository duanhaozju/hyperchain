package cn.hyperchain.jcee.contract.examples.ABC;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Created by huhu on 2017/5/31.
 */

public class Order {
    String orderNum;
    int amount;
    String buyerAccountNum;
    String buyerAccountName;
    String buyerBankNum;
    String sellerAccountNum;
    String sellerAccountName;
    String sellerBankNum;
    String orderConfirmTime;
    String orderDueTime;
    String orderState;
    int draftAmount;//订单所有票据的面额总和
    String updatable;

    public Order(){}

    public Order(String orderNum,int amount,String buyerAccountNum,String buyerAccountName,
            String buyerBankNum,String sellerAccountNum,String sellerAccountName,String sellerBankNum,
            String orderConfirmTime,String orderDueTime,String orderState){
        this.orderNum = orderNum;
        this.amount = amount;
        this.buyerAccountNum = buyerAccountNum;
        this.buyerAccountName = buyerAccountName;
        this.buyerBankNum = buyerBankNum;
        this.sellerAccountNum = sellerAccountNum;
        this.sellerAccountName = sellerAccountName;
        this.sellerBankNum = sellerBankNum;
        this.orderConfirmTime = orderConfirmTime;
        this.orderDueTime = orderDueTime;
        this.orderState = orderState;
    }

    public String getOrderNum() {
        return orderNum;
    }

    public void setOrderNum(String orderNum) {
        this.orderNum = orderNum;
    }

    public int getAmount() {
        return amount;
    }

    public void setAmount(int amount) {
        this.amount = amount;
    }

    public String getBuyerAccountNum() {
        return buyerAccountNum;
    }

    public void setBuyerAccountNum(String buyerAccountNum) {
        this.buyerAccountNum = buyerAccountNum;
    }

    public String getBuyerAccountName() {
        return buyerAccountName;
    }

    public void setBuyerAccountName(String buyerAccountName) {
        this.buyerAccountName = buyerAccountName;
    }

    public String getBuyerBankNum() {
        return buyerBankNum;
    }

    public void setBuyerBankNum(String buyerBankNum) {
        this.buyerBankNum = buyerBankNum;
    }

    public String getSellerAccountNum() {
        return sellerAccountNum;
    }

    public void setSellerAccountNum(String sellerAccountNum) {
        this.sellerAccountNum = sellerAccountNum;
    }

    public String getSellerAccountName() {
        return sellerAccountName;
    }

    public void setSellerAccountName(String sellerAccountName) {
        this.sellerAccountName = sellerAccountName;
    }

    public String getSellerBankNum() {
        return sellerBankNum;
    }

    public void setSellerBankNum(String sellerBankNum) {
        this.sellerBankNum = sellerBankNum;
    }

    public String getOrderConfirmTime() {
        return orderConfirmTime;
    }

    public void setOrderConfirmTime(String orderConfirmTime) {
        this.orderConfirmTime = orderConfirmTime;
    }

    public String getOrderDueTime() {
        return orderDueTime;
    }

    public void setOrderDueTime(String orderDueTime) {
        this.orderDueTime = orderDueTime;
    }

    public String getOrderState() {
        return orderState;
    }

    public void setOrderState(String orderState) {
        this.orderState = orderState;
    }

    public int getDraftAmount() {
        return draftAmount;
    }

    public void setDraftAmount(int draftAmount) {
        this.draftAmount = draftAmount;
    }

    public String getUpdatable() {
        return updatable;
    }

    public void setUpdatable(String updatable) {
        this.updatable = updatable;
    }
}
