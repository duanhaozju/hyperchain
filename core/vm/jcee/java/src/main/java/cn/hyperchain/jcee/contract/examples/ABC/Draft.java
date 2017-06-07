package cn.hyperchain.jcee.contract.examples.ABC;


/**
 * Created by huhu on 2017/5/31.
 */
public class Draft {
    private  String draftNum;
    private  String draftTypes;
    private  int amount;
    private  int issueDraftApplyDate;
    private  int dueDate;
    private String drawerId;
    private String acceptorId;
    private String payeeId;
    private String bearerId;
    private String note;
    private String orderNum;
    private boolean autoReceiveDraft;
    private String draftStatus;
    private String lastStatus;
    private String firstOwner;
    private String secondOwner;
    private String discountRate;

    public Draft(){}

    public Draft(String draftNum, String draftTypes, int amount, int issueDraftApplyDate,
                 int dueDate, String drawerId,String acceptorId,String payeeId,String notes,
                 String orderNum, boolean isAuto){

        this.drawerId = drawerId;
        this.acceptorId = acceptorId;
        this.payeeId = payeeId;
        this.bearerId = drawerId;
        this.note = notes;
        this.orderNum = orderNum;
        this.firstOwner = drawerId;
        this.autoReceiveDraft = isAuto;

        this.draftNum = draftNum;
        this.draftTypes = draftTypes;
        this.draftStatus = "02001";

        this.dueDate = dueDate;
        this.amount = amount;
        this.lastStatus = "02001";
        this.issueDraftApplyDate = issueDraftApplyDate;
    }

    public String getDraftNum() {
        return draftNum;
    }

    public void setDraftNum(String draftNum) {
        this.draftNum = draftNum;
    }

    public String getDraftTypes() {
        return draftTypes;
    }

    public void setDraftTypes(String draftTypes) {
        this.draftTypes = draftTypes;
    }

    public int getAmount() {
        return amount;
    }

    public void setAmount(int amount) {
        this.amount = amount;
    }

    public int getIssueDraftApplyDate() {
        return issueDraftApplyDate;
    }

    public void setIssueDraftApplyDate(int issueDraftApplyDate) {
        this.issueDraftApplyDate = issueDraftApplyDate;
    }

    public int getDueDate() {
        return dueDate;
    }

    public void setDueDate(int dueDate) {
        this.dueDate = dueDate;
    }

    public String getDrawerId() {
        return drawerId;
    }

    public void setDrawerId(String drawerId) {
        this.drawerId = drawerId;
    }

    public String getAcceptorId() {
        return acceptorId;
    }

    public void setAcceptorId(String acceptorId) {
        this.acceptorId = acceptorId;
    }

    public String getPayeeId() {
        return payeeId;
    }

    public void setPayeeId(String payeeId) {
        this.payeeId = payeeId;
    }

    public String getBearerId() {
        return bearerId;
    }

    public void setBearerId(String bearerId) {
        this.bearerId = bearerId;
    }

    public String getNote() {
        return note;
    }

    public void setNote(String note) {
        this.note = note;
    }

    public String getOrderNum() {
        return orderNum;
    }

    public void setOrderNum(String orderNum) {
        this.orderNum = orderNum;
    }

    public boolean isAutoReceiveDraft() {
        return autoReceiveDraft;
    }

    public void setAutoReceiveDraft(boolean autoReceiveDraft) {
        this.autoReceiveDraft = autoReceiveDraft;
    }

    public String getDraftStatus() {
        return draftStatus;
    }

    public void setDraftStatus(String draftStatus) {
        this.draftStatus = draftStatus;
    }

    public String getLastStatus() {
        return lastStatus;
    }

    public void setLastStatus(String lastStatus) {
        this.lastStatus = lastStatus;
    }

    public String getFirstOwner() {
        return firstOwner;
    }

    public void setFirstOwner(String firstOwner) {
        this.firstOwner = firstOwner;
    }

    public String getSecondOwner() {
        return secondOwner;
    }

    public void setSecondOwner(String secondOwner) {
        this.secondOwner = secondOwner;
    }

    public String getDiscountRate() {
        return discountRate;
    }

    public void setDiscountRate(String discountRate) {
        this.discountRate = discountRate;
    }
}
