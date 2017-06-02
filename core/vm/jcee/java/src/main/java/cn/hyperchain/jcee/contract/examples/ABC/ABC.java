package cn.hyperchain.jcee.contract.examples.ABC;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.jcee.ledger.Batch;
import cn.hyperchain.jcee.ledger.BatchKey;
import cn.hyperchain.jcee.util.Bytes;
import org.apache.log4j.Logger;

import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;

/**
 * Created by huhu on 2017/5/31.
 */
public class ABC extends ContractTemplate {

    private static final Logger logger = Logger.getLogger(ABC.class.getSimpleName());
    private static final SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");

    private static final String accountPrefix = "account_";
    private static final String orderPrefix = "order_";
    private static final String draftPrefix = "draft_";


    @Override
    public ExecuteResult invoke(String funcName, List<String> args) {
        switch (funcName){
            case "newAccount":
                return newAccount(args);
            case "pubOrderInfo":
                return pubOrderInfo(args);
            case "issueDraftApply":
                return issueDraftApply(args);
            case "acceptByAccount":
                return acceptByAccount(args);
            case "getDraft":
                return getDraft(args);
            default:
                return result(false,funcName+"does not exist");

        }
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

        long time = System.currentTimeMillis();
        String modifyTime = sdf.format(new Date(time));

//        if(!IDTypes.contains(IDType)){
//            msg = "IDType is wrong";
//            logger.error(msg);
//            return result(false, msg);
//        }
//        if(!accountType.contains(accountType)){
//            msg = "accountType is wrong";
//            logger.error(msg);
//            return result(false, msg);
//        }
        if(ledger.get(accountPrefix +accountNumber) != null ){
            msg = "the accountNumber exists";
            logger.error(msg);
            return result(false,msg);
        }
        Account account = new Account(accountNumber,name,ID,IDType,accountType,
                businessBankNum,businessBankName,addr,phoneNum,modifyTime);
        logger.info("the account key is "+ accountPrefix +accountNumber);
        logger.info("the account value is "+ Bytes.toByteArray(account));
        if(ledger.put(accountPrefix +accountNumber,account) == false){
            msg = "put account data error";
            logger.error(msg);
            return result(false, msg);
        }

        return result(true);
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

        Account buy = ledger.getObject(accountPrefix+buyerAccountNum,Account.class);
        logger.info("the buyer num get from ledger is "+buy.getAccountNumber());

        Account sell = ledger.getObject(accountPrefix+sellerAccountNum,Account.class);
        logger.info("the seller num get from ledger is "+sell.getAccountNumber());

        logger.info("the order key "+orderPrefix+orderNum);
        Order data = ledger.getObject(orderPrefix+orderNum,Order.class);
        logger.info("get data from ledger");
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

        byte[] buyerAccount = batch.get(buyerKey);
        byte[] sellerAccount = batch.get(sellerKey);
        if(buyerAccount == null || sellerAccount == null){
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

        logger.info("the order key "+orderPrefix+orderNum);
        Order order = ledger.getObject(orderPrefix+orderNum,Order.class);
        if(order == null){
            msg = "the order has not been published";
            logger.error(msg);
            return result(false, msg);
        }
        if(order.getAmount() < amount + order.getDraftAmount()){
            msg = "order amount error";
            logger.error(msg);
            return result(false,msg);
        }
        Account account = ledger.getObject(accountPrefix+drawerId,Account.class);
        if(account == null){
            msg = "drawer account error";
            logger.error(msg);
            return result(false, msg);
        }
        if(judgeAcceptor(acceptorId, draftType)){
            if(judgePayee(payeeId)){
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
                        params.add(account.getName());
                        params.add(account.getBusinessBankNum());
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
        Draft draft = ledger.getObject(draftPrefix+draftNum,Draft.class);
        if(draft == null){
            String msg = "draft is null";
            logger.error(msg);
            return result(false,msg);
        }
        logger.info(draft.getDraftNum());
        logger.info(draft.getFirstOwner());
        logger.info(draft.getSecondOwner());
        logger.info(draft.getDraftStatus());
        logger.info(draft.getLastStatus());
        logger.info(draft.getAmount());
        return result(true);
    }

    public ExecuteResult acceptByAccount(List<String> args){
        String replyerNum = args.get(0);
        String replyerName = args.get(1);
        String replyerBankNum = args.get(2);

        String draftNum = args.get(3);
        String responseType = args.get(4);

        String msg;
        logger.info("accept by account");

        if(!judgeAccountCorrect(replyerNum,replyerName,replyerBankNum)){
            msg = "the replyer info is wrong";
            logger.equals(msg);
            return result(false,msg);
        }
        Draft draft = ledger.getObject(draftPrefix+draftNum,Draft.class);
        if(draft == null){
            msg = "draft is not exist";
            logger.error(msg);
            return result(false,msg);
        }
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
        Order order = ledger.getObject(orderPrefix+draft.getOrderNum(),Order.class);
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
                if(ledger.put(orderPrefix+order.getOrderNum(),order) == false){
                    msg ="put order data error";
                    logger.error(msg);
                    return result(false,msg);
                }
            }

            if(draft.isAutoReceiveDraft()){
                Account account = ledger.getObject(accountPrefix +draft.getDrawerId(),Account.class);
                ExecuteResult result1 = receiveDraftApply(draft,draft.getDrawerId(),
                        account.getName(),account.getBusinessBankNum());
                if(!result1.isSuccess()){
                    return result(false,result1.getResult());
                }
            }
        }
        if(ledger.put(draftPrefix+draftNum,draft) == false){
            msg ="put draft data error";
            logger.error(msg);
            return result(false,msg);
        }

        return result(true,"accept success");
    }

    public ExecuteResult receiveDraftApply(Draft draft,String applicantAccountNum,
                                               String applicantName,String applicantBankNum){
        ArrayList<String> result = new ArrayList<>();
        String msg;
        if(!judgeAccountCorrect(applicantAccountNum,applicantName,applicantBankNum)){
            msg = "name or bankNum error";
            return result(false,msg);
        }

        if(!draft.getDrawerId().equals(applicantAccountNum)){
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

    public boolean judgeAccountCorrect(String accountNum,String name,String businessBankNum){
        Account account = ledger.getObject(accountPrefix +accountNum,Account.class);
        if(account == null){
            return false;
        }
        if(account.getName().equals(name) && account.getBusinessBankNum().equals(businessBankNum)){
            return true;
        }
        return false;
    }

    public boolean judgeAcceptor(String acceptorId, String draftType){
        Account acceptor =  ledger.getObject(accountPrefix +acceptorId,Account.class);
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

    public boolean judgePayee(String payeeId){
        Account payee = ledger.getObject(accountPrefix +payeeId,Account.class);
        if(payee == null || payee.getAccountType().equals("RC00")){
            return false;
        }
        return true;
    }


}
