package cn.hyperchain.jcee.contract.examples.sb;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
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
public class SimulateBank extends ContractTemplate {
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
     * invoke smart contract method
     *
     * @param funcName function name user defined in contract
     * @param args     arguments of funcName
     */
    @Override
    public ExecuteResult invoke(String funcName, List<String> args) {
        switch (funcName) {
            case "issue":
                return issue(args);
            case "transfer":
                return transfer(args);
            case "transferByBatch":
                return transferByBatch(args);
            case "getAccountBalance":
                return getAccountBalance(args);
            case "testRangeQuery":
                return testRangeQuery(args);
            case "testDelete":
                return testDelete(args);
            default:
                String err = "method " + funcName  + " not found!";
                logger.error(err);
                return new ExecuteResult(false, err);

        }
    }

    //String account, double num
    private ExecuteResult issue(List<String> args) {
        if(args.size() != 2) {
            logger.error("args num is invalid");
            return result(false, "args num is invalid");
        }
        logger.info("account: " + args.get(0));
        logger.info("num: " + args.get(1));

        boolean rs = ledger.put(args.get(0).getBytes(), args.get(1).getBytes());
        if(rs == false) {
            logger.error("issue func error");
            return result(false, "put data error");
        }
        return result(true);
    }

    //String accountA, String accountB, double num
    private ExecuteResult transfer(List<String> args) {
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
                String msg = "get account " + accountA  + " balance error";
                logger.error(msg);
                return result(false, msg);
            }

        }catch (Exception e) {
            e.printStackTrace();
        }

        return result(true);
    }

    private ExecuteResult getAccountBalance(List<String> args) {
        if(args.size() != 1) {
            logger.error("args num is invalid");
        }
        try {
            byte[] data = ledger.get(args.get(0).getBytes());
            logger.info(new String(data));
            if (data != null) {
                return result(true, data);
            }else {
                String msg = "getAccountBalance error no data found for" + args.get(0);
                logger.error(msg);
                return result(false, msg);
            }
        }catch (Exception e) {
            e.printStackTrace();
            return result(false, e);
        }
    }

    //1.test read batch
    //2.test write batch
    private ExecuteResult transferByBatch(List<String> args) {
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
            return result(false, args.get(0) + " no account");
        }
        double abalance = Bytes.toDouble(ba);
        byte[] bb = batch.get(B);
        if (bb == null) {
            return result(false, args.get(1) + " no account");
        }

        double bbalance = Bytes.toDouble(bb);
        double amount = Bytes.toDouble(args.get(2).getBytes());
        if (abalance < abalance) {
            return result(false, args.get(0) + " balance is not enough");
        }

        Batch wb = ledger.newBatch();
        wb.put(A, Bytes.toByteArray(abalance - amount));
        wb.put(B, Bytes.toByteArray(bbalance + abalance));
        return result(wb.commit());
    }

    //testRangeQuery
    private ExecuteResult testRangeQuery(List<String> args) {
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
        return result(bvCount == count);
    }

    public ExecuteResult testDelete(List<String> args) {
        String key = "key-001";
        String value = "vvv";
        if (ledger.put(key, value) == false) return result(false);
        logger.info("put success");
        if (ledger.delete(key) == false) return result(false);
        logger.info("delete success");
        String getV = ledger.getString(key);
        logger.info("get deleted value is " + getV);
        return result(getV.isEmpty());
    }
}