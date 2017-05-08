package cn.hyperchain.jcee.contract.examples.sb;

import cn.hyperchain.jcee.contract.ContractBase;
import cn.hyperchain.jcee.ledger.Batch;
import cn.hyperchain.jcee.ledger.BatchKey;
import cn.hyperchain.jcee.ledger.BatchValue;
import cn.hyperchain.jcee.util.Bytes;
import com.google.protobuf.ByteString;
import org.apache.log4j.Logger;

import java.util.List;

/**
 * Created by wangxiaoyi on 2017/4/14.
 */
public class SimulateBank extends ContractBase{
    private static final Logger logger = Logger.getLogger(SimulateBank.class.getSimpleName());

    private String bankName;
    private long bankNum;
    private boolean isValid;

    public SimulateBank() {}

    public SimulateBank(String bankName, long bankNum, boolean isValid){
        this.bankName = bankName;
        this.bankNum = bankNum;
        this.isValid = isValid;
    }

    /**
     * Invoke smart contract method
     *
     * @param funcName function name user defined in contract
     * @param args     arguments of funcName
     */
    @Override
    public boolean Invoke(String funcName, List<String> args) {
        switch (funcName) {
            case "issue":
                return issue(args);
            case "transfer":
                return transfer(args);
            case "transferByBatch":
                return transferByBatch(args);
            default:
                logger.error("method " + funcName  + " not found!");

        }
        return false;
    }

    /**
     * Query data stored in the smart contract
     *
     * @param funcName function name
     * @param args     function related arguments
     * @return the query result
     */
    @Override
    public ByteString Query(String funcName, List<String> args) {
        switch (funcName) {
            case "getAccountBalance":
                return getAccountBalance(args);
            default:
                String errMsg = "method " + funcName  + " not found!";
                logger.error(errMsg);
        }
        return null;
    }

    //String account, double num
    private boolean issue(List<String> args) {
        if(args.size() != 2) {
            logger.error("args num is invalid");
        }
        logger.info("account: " + args.get(0));
        logger.info("num: " + args.get(1));

        boolean rs = ledger.put(args.get(0).getBytes(), args.get(1).getBytes());
        if(rs == false) {
            logger.error("issue func error");
        }
        return true;
    }

    //String accountA, String accountB, double num
    private boolean transfer(List<String> args) {
        try {
            String accountA = args.get(0);
            String accountB = args.get(1);
            double num = Double.valueOf(args.get(2));
            byte[] balance = ledger.get(accountA.getBytes());

            if(balance != null) {
                double balanceA = Double.parseDouble(new String(balance));
                double balanceB = ledger.getDouble(accountB.getBytes());
                if (balanceA >= num) {
                    ledger.put(accountA, balanceA - num);
                    ledger.put(accountB, balanceB + num);
                }
            }else {
                logger.error("get account " + accountA  + " balance error");
                return false;
            }

        }catch (Exception e) {
            e.printStackTrace();
        }

        return true;
    }

    private ByteString getAccountBalance(List<String> args) {
        if(args.size() != 1) {
            logger.error("args num is invalid");
        }
        try {
            byte[] data = ledger.get(args.get(0).getBytes());
            logger.info(new String(data));
            if (data != null) {
                return ByteString.copyFrom(data);
            }else {
                logger.error("getAccountBalance error");
            }
        }catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }

    //1.test read batch
    //2.test write batch
    private boolean transferByBatch(List<String> args) {
        if(args.size() != 3) {
            logger.error("args num is invalid");
        }
        byte[] A = args.get(0).getBytes();
        byte[] B = args.get(1).getBytes();

        BatchKey bk = ledger.newBatchKey();
        bk.put(A);
        bk.put(B);
        Batch batch = ledger.batchRead(bk);
        byte[] ba = batch.get(A);
        if (ba == null) {
            return false;
        }
        double abalance = Bytes.toDouble(ba);
        byte[] bb = batch.get(B);
        if (bb == null) {
            return false;
        }

        double bbalance = Bytes.toDouble(bb);
        double amount = Bytes.toDouble(args.get(2).getBytes());
        if (abalance < abalance) {
            return false;
        }

        Batch wb = ledger.newBatch();
        wb.put(A, Bytes.toByteArray(abalance - amount));
        wb.put(B, Bytes.toByteArray(bbalance + abalance));
        return wb.commit();
    }

    //testRangeQuery
    private boolean testRangeQuery(List<String> args) {
        Batch batch = ledger.newBatch();
        String keyPrefix = "bk-";
        int count = 10009;
        for (int i = 0; i < count; i ++) {
            batch.put((keyPrefix + i).getBytes(), (i + "").getBytes());
        }
        batch.commit();

        BatchValue bv = ledger.rangeQuery("bk-0".getBytes(), "bk-10009".getBytes());
        int bvCount = 0;
        while (bv.hasNext()) {
            bv.next();
            bvCount ++;
        }
        return bvCount == count;
    }
}