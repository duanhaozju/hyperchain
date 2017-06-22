/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract.examples.ABC.src;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.jcee.ledger.Batch;
import cn.hyperchain.jcee.ledger.BatchKey;
import cn.hyperchain.jcee.ledger.Result;
import cn.hyperchain.jcee.util.Bytes;

import java.util.ArrayList;
import java.util.List;

/**
 * Created by huhu on 2017/5/31.
 */
public class ABC extends ContractTemplate {

    private static final String accountPrefix = "account_";
    private static final String orderPrefix = "order_";
    private static final String draftPrefix = "draft_";


    @Override
    public ExecuteResult invoke(String funcName, List<String> args) {
        switch (funcName){
            case "newAccount":
                return newAccount(args);
            case "getAccount":
                return getAccount(args);
            case "pubOrderInfo":
                return pubOrderInfo(args);
            case "issueDraftApply":
                return issueDraftApply(args);
            case "acceptByAccount":
                return acceptByAccount(args);
            case "getDraft":
                return getDraft(args);
            case "testBatch":
                return testBatch(args);
            default:
                return result(false,funcName+" does not exist");

        }
    }

    public ExecuteResult testBatch(List<String> args){
        Account account1 = new Account("A1","1","1","1","1","1","1","1","1");
        Account account2 = new Account("A2","1","1","1","1","1","1","1","1");

        byte[] key1 = account1.getAccountNumber().getBytes();
        byte[] key2 = account2.getAccountNumber().getBytes();
        ledger.put("A1",account1);
        ledger.put("A2",account2);

        logger.info(ledger.get("A1").toObeject(Account.class).getAccountNumber());

        BatchKey bk = ledger.newBatchKey();
        bk.put(key1);
        bk.put(key2);
        Batch batch = ledger.batchRead(bk);
        logger.info(batch.get(key1).toObeject(Account.class).getAccountNumber());

        //abnormal
        Result result = batch.get("A11");
        if(result.isEmpty()){
            logger.info("the value is empty");
        }
        return result(true);


    }

    public ExecuteResult newAccount(List<String> args){

        String accountNumber = args.get(0);
        String name = args.get(1);
        String ID = args.get(2);
        String IDType = args.get(3);
        String accountType = args.get(4);
        String businessBankNum = args.get(5);
        String businessBankName = args.get(6);
        String addr = args.get(7);
        String phoneNum = args.get(8);
        String msg;

        Result result = ledger.get(accountPrefix +accountNumber);
        if(!result.isEmpty() ){
            msg = "the accountNumber exists";
            logger.error(msg);
            return result(false,msg);
        }
        Account account = new Account(accountNumber,name,ID,IDType,accountType,
                businessBankNum,businessBankName,addr,phoneNum);
        logger.info("the account key is "+ accountPrefix +accountNumber);
        if(ledger.put(accountPrefix +accountNumber,account) == false){
            msg = "put account data error";
            logger.error(msg);
            return result(false, msg);
        }

        return result(true);
    }

    public ExecuteResult getAccount(List<String> args){
        String accountNumber = args.get(0);
        Result result = ledger.get(accountPrefix +accountNumber);
        String msg;
        if(result.isEmpty() ){
            msg = "the accountNumber does not exist";
            logger.error(msg);
            return result(false,msg);
        }
        Account account = result.toObeject(Account.class);

        return result(true,account);

    }

    public ExecuteResult pubOrderInfo(List<String> args){

        String orderNum = args.get(0);
        int amount = Integer.parseInt(args.get(1));
        String buyerAccountNum = args.get(2);
        String buyerAccountName = args.get(3);
        String buyerBankNum = args.get(4);
        String sellerAccountNum = args.get(5);
        String sellerAccountName = args.get(6);
        String sellerBankNum = args.get(7);
        String orderConfirmTime = args.get(8);
        String orderDueTime = args.get(9);
        String msg;

        Result result = ledger.get(accountPrefix+buyerAccountNum);
        Account buy = null;
        if(!result.isEmpty()){
            buy = result.toObeject(Account.class);
        }

        result = ledger.get(accountPrefix+sellerAccountNum);
        Account sell= null;
        if(!result.isEmpty()){
            sell = result.toObeject(Account.class);
        }

        result = ledger.get(orderPrefix+orderNum);
        Order data= null;
        if(!result.isEmpty()){
            data = result.toObeject(Order.class);
        }
        if(data != null){
            msg = "the order has been published";
            logger.error(msg);
            return result(false,msg);
        }

        byte[] buyerKey = (accountPrefix +buyerAccountNum).getBytes();
        byte[] sellerKey = (accountPrefix +sellerAccountNum).getBytes();
        BatchKey bk = ledger.newBatchKey();
        bk.put(buyerKey);
        bk.put(sellerKey);
        Batch batch = ledger.batchRead(bk);

        Result buyer = batch.get(buyerKey);
        Result seller = batch.get(sellerKey);
        if(buyer.isEmpty() || seller.isEmpty()){
            msg = "the buyer account or seller account does not exist";
            logger.error(msg);
            return result(false,msg);
        }
        Order order = new Order(orderNum,amount,buyerAccountNum,
                buyerAccountName,buyerBankNum, sellerAccountNum,sellerAccountName,
                sellerBankNum,orderConfirmTime,orderDueTime,"VALID");
        if(ledger.put(orderPrefix+orderNum,order) == false){
            msg = "put order data error";
            logger.error(msg);
            return result(false,msg);
        }
        //todo 写accountOrders,维护每个accout对应的orderNum数组
        return result(true);

    }

    public ExecuteResult issueDraftApply(List<String> args){
        String draftNum = args.get(0);
        String draftType = args.get(1);
        int amount = Integer.parseInt(args.get(2));
        int issueDraftApplyDate = Integer.parseInt(args.get(3));
        int dueDate = Integer.parseInt(args.get(4));
        String drawerId = args.get(5);
        String acceptorId = args.get(6);
        String payeeId = args.get(7);
        String note = args.get(8);
        String orderNum = args.get(9);
        boolean autoReceiveDraft = Boolean.parseBoolean(args.get(10));
        String msg;

        byte[] drawerKey = (accountPrefix+drawerId).getBytes();
        byte[] acceptorKey = (accountPrefix+acceptorId).getBytes();
        byte[] payeeKey = (accountPrefix+payeeId).getBytes();
        byte[] orderKey = (orderPrefix+orderNum).getBytes();
        BatchKey bk = ledger.newBatchKey();
        bk.put(drawerKey);
        bk.put(acceptorKey);
        bk.put(payeeKey);
        bk.put(orderKey);
        Batch batch = ledger.batchRead(bk);

        Result result = batch.get(orderKey);
        if(result.isEmpty()){
            msg = "the order has not been published";
            logger.error(msg);
            return result(false, msg);
        }
        Order order = null;
        if(!result.isEmpty()){
            order = result.toObeject(Order.class);
        }

        if(order.getAmount() < amount + order.getDraftAmount()){
            msg = "order amount error";
            logger.error(msg);
            return result(false,msg);
        }

        result = batch.get(drawerKey);

        if(result.isEmpty()){
            msg = "drawer account error";
            logger.error(msg);
            return result(false, msg);
        }

        result = batch.get(acceptorKey);

        if(result.isEmpty()){
            msg = "acceptor account error";
            logger.error(msg);
            return result(false, msg);
        }

        Account acceptor = null;
        if(!result.isEmpty()){
            acceptor = result.toObeject(Account.class);
        }

        result = batch.get(payeeKey);
        if(result.isEmpty()){
            msg = "payee account error";
            logger.error(msg);
            return result(false, msg);
        }
        Account payee = null;
        if(!result.isEmpty()){
            payee = result.toObeject(Account.class);
        }

        if(judgeAcceptor(acceptor, draftType)){
            if(judgePayee(payee)){
                Draft draft = new Draft(draftNum,draftType,amount,issueDraftApplyDate,
                        dueDate,drawerId,acceptorId,payeeId,note,orderNum,autoReceiveDraft);
                if(ledger.put(draftPrefix+draftNum, draft) == false){
                    msg = "put draft data error";
                    logger.error(msg);
                    return result(false,msg);
                }
                if(draftType.equals("AC02")){
                    if(drawerId.equals(acceptorId)){
                        logger.info("draft type is AC02");
                        draft.setDraftStatus("020006");
                        //todo return accept by account
                        List<String> params = new ArrayList<>();
                        params.add(acceptorId);
                        params.add(acceptor.getName());
                        params.add(acceptor.getBusinessBankNum());
                        params.add(draftNum);
                        params.add("SU00");

                        return acceptByAccount(params);
                    }
                }
                return result(true,"draft apply success");
            }else {
                return result(false,"payee account error");
            }
        }else {
            return result(false,"acceptor account type error");
        }
    }

    public ExecuteResult getDraft(List<String> args){
        String draftNum = args.get(0);
        Result result = ledger.get(draftPrefix+draftNum);

        if(result.isEmpty()){
            String msg = "draft is null";
            logger.error(msg);
            return result(false,msg);
        }

        Draft draft = result.toObeject(Draft.class);

        logger.info(draft.getDraftNum());
        logger.info(draft.getFirstOwner());
        logger.info(draft.getSecondOwner());
        logger.info(draft.getDraftStatus());
        logger.info(draft.getLastStatus());
        logger.info(draft.getAmount());
//        return result(true);
        logger.info("testBatch");
        Account account1 = new Account("A1","1","1","1","1","1","1","1","1");
        Account account2 = new Account("A2","1","1","1","1","1","1","1","1");

        ledger.put(account1.getAccountNumber(),account1);
        ledger.put(account2.getAccountNumber(),account2);

        result = ledger.get("A1");
        if(!result.isEmpty()){
            Account account = result.toObeject(Account.class);
            logger.info(account.getAccountNumber());
        }

        return result(true);
    }

    public ExecuteResult acceptByAccount(List<String> args){
        String replyerNum = args.get(0);
        String replyerName = args.get(1);
        String replyerBankNum = args.get(2);

        String draftNum = args.get(3);
        String responseType = args.get(4);

        String msg;

        BatchKey bk = ledger.newBatchKey();
        bk.put(accountPrefix+replyerNum);
        bk.put(draftPrefix+draftNum);

        Batch batch = ledger.batchRead(bk);

        Result result = batch.get(accountPrefix+replyerNum);
        if(result.isEmpty()){
            msg = "reply account error";
            logger.error(msg);
            return result(false, msg);
        }
        Account replyer = null;
        if(!result.isEmpty()){
            replyer = result.toObeject(Account.class);
        }

        if(!judgeAccountCorrect(replyer,replyerName,replyerBankNum)){
            msg = "the replyer info is wrong";
            logger.equals(msg);
            return result(false,msg);
        }

        result = batch.get(draftPrefix+draftNum);

        if(result.isEmpty()){
            msg = "draft is not exist";
            logger.error(msg);
            return result(false,msg);
        }

        Draft draft = result.toObeject(Draft.class);
        if(!draft.getDraftStatus().equals("020001")){
            msg = "draft status is not satisfied";
            logger.error(msg);
            return result(false,msg);
        }
        if(draft.getDraftTypes().equals("AC01")){
            msg = "no permission";
            logger.equals(msg);
            return result(false,msg);
        }

        bk = ledger.newBatchKey();
        bk.put(orderPrefix+draft.getOrderNum());
        bk.put(accountPrefix +draft.getDrawerId());
        batch = ledger.batchRead(bk);

        result = batch.get(orderPrefix+draft.getOrderNum());
        Order order = null;
        if(!result.isEmpty()){
            order = result.toObeject(Order.class);
        }
        if(responseType.equals("SU01")){
            draft.setDraftStatus("000002");
            order.setDraftAmount(order.getDraftAmount()+draft.getAmount());

            if(ledger.put(orderPrefix+order.getOrderNum(),order) == false){
                msg ="put order data error";
                logger.error(msg);
                return result(false,msg);
            }

            if(ledger.put(draftPrefix+draftNum,draft) == false){
                msg ="put draft data error";
                logger.error(msg);
                return result(false,msg);
            }
            return result(true,"resp reject");
        }else if(responseType.equals("SU00")){
            draft.setLastStatus(draft.getDraftStatus());
            draft.setDraftStatus("020006");
            if(order.getUpdatable().equals("YES")){
                order.setUpdatable("NO");
            }

            if(draft.isAutoReceiveDraft()){
                result = batch.get(accountPrefix +draft.getDrawerId());
                Account account = null;
                if(!result.isEmpty()){
                    account = result.toObeject(Account.class);
                }
                ExecuteResult result1 = receiveDraftApply(draft,account,
                        account.getName(),account.getBusinessBankNum());
                if(!result1.isSuccess()){
                    return result(false,result1.getResult());
                }
            }
        }

        Batch wb = ledger.newBatch();
        wb.put(orderPrefix+order.getOrderNum(),order);
        wb.put(draftPrefix+draftNum,draft);
        return result(wb.commit());

    }

    public ExecuteResult receiveDraftApply(Draft draft,Account account,
                                               String applicantName,String applicantBankNum){
        ArrayList<String> result = new ArrayList<>();
        String msg;
        if(!judgeAccountCorrect(account,applicantName,applicantBankNum)){
            msg = "name or bankNum error";
            return result(false,msg);
        }

        if(!draft.getDrawerId().equals(account.getAccountNumber())){
            msg = "applicant is not drawer";
            return result(false,msg);
        }
        if(draft.getDraftStatus().equals("020006")){
            draft.setFirstOwner(draft.getDrawerId());
            draft.setSecondOwner(draft.getPayeeId());
            draft.setDraftStatus("030001");
            draft.setLastStatus("020006");
            return result(true,"apply receive draft success");
        }else {
            msg = "draftStatus is not satisfied";
            return result(false,msg);
        }

    }

    public boolean judgeAccountCorrect(Account account,String name,String businessBankNum){

        if(account == null){
            return false;
        }
        if(account.getName().equals(name) && account.getBusinessBankNum().equals(businessBankNum)){
            return true;
        }
        return false;
    }

    public boolean judgeAcceptor(Account acceptor, String draftType){

        if(acceptor == null){
            return false;
        }
        if(draftType.equals("AC01") && acceptor.getAccountType().equals("RC00")){
            return true;
        }
        if(draftType.equals("AC02") && acceptor.getAccountType().equals("RC01")){
            return true;
        }
        return false;
    }

    public boolean judgePayee(Account payee){
        if(payee == null || payee.getAccountType().equals("RC00")){
            return false;
        }
        return true;
    }
}
