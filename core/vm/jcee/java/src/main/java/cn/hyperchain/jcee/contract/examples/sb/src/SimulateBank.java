/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract.examples.sb.src;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.jcee.contract.Event;
import cn.hyperchain.jcee.ledger.Batch;
import cn.hyperchain.jcee.ledger.BatchKey;
import cn.hyperchain.jcee.ledger.BatchValue;
import cn.hyperchain.jcee.ledger.Result;
import cn.hyperchain.jcee.ledger.table.ColumnDesc;
import cn.hyperchain.jcee.ledger.table.RelationDB;
import cn.hyperchain.jcee.ledger.table.TableDesc;
import cn.hyperchain.jcee.ledger.table.TableName;
import cn.hyperchain.jcee.util.Bytes;
import cn.hyperchain.jcee.util.DataType;

import java.util.LinkedList;
import java.util.List;

/**
 * Created by wangxiaoyi on 2017/4/14.
 */
public class SimulateBank extends ContractTemplate {

    public SimulateBank() {}

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
            case "testInvokeContract":
                logger.info("testInvokeContract");
                return testInvokeContract(args);
            case "testPostEvent":
                return testPostEvent(args);
            case "newAccountTable":
                return newAccountTable(args);
            case "getTableDesc":
                return getTableDesc(args);
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

            Result result = ledger.get(accountA.getBytes());

            if(!result.isEmpty()) {
                double balanceA = result.toDouble();
                result = ledger.get(accountB.getBytes());
                double balanceB ;
                if(!result.isEmpty()){
                    balanceB = result.toDouble();
                    if (balanceA >= num) {
                        ledger.put(accountA, balanceA - num);
                        ledger.put(accountB, balanceB + num);
                    }
                }

            }else {
                String msg = "get account " + accountA  + " balance error";
                logger.error(msg);
                return result(false, msg);
            }

        }catch (Exception e) {
            logger.error(e.getMessage());
            return result(false, e.getMessage());
        }

        return result(true);
    }

    private ExecuteResult getAccountBalance(List<String> args) {
        if(args.size() != 1) {
            logger.error("args num is invalid");
        }
        try {
            Result result = ledger.get(args.get(0).getBytes());
            if (!result.isEmpty()) {
                return result(true, result.toDouble());
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
        Result ba = batch.get(A);
        if (ba.isEmpty()) {
            return result(false, args.get(0) + " no account");
        }
        double abalance = ba.toDouble();
        Result bb = batch.get(B);
        if (bb.isEmpty()) {
            return result(false, args.get(1) + " no account");
        }

        double bbalance = bb.toDouble();
        double amount = Bytes.toDouble(args.get(2).getBytes());
        if (abalance < amount) {
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
        int count = 9;
        for (int i = 0; i < count; i ++) {
            batch.put((keyPrefix + i).getBytes(), (i + "").getBytes());
        }
        batch.commit();

        BatchValue bv = ledger.rangeQuery("bk-0".getBytes(), "bk-9".getBytes());
        int bvCount = 0;
        while (bv.hasNext()) {
            bv.next();
            bvCount ++;
        }
        logger.info(bvCount == count);
        return result(bvCount == count);
    }

    public ExecuteResult testDelete(List<String> args) {
        String key = "key-001";
        String value = "vvv";
        if (ledger.put(key, value) == false) return result(false);
        logger.info("put success");
        if (ledger.delete(key) == false) return result(false);
        logger.info("delete success");
        String getV = "";

        Result result = ledger.get(key);
        if(!result.isEmpty()){
            getV = result.toString();
            logger.error("get deleted value is " + getV);
        }
        else {
            logger.info("the value has been deleted and is empty now");
        }
        return result(getV.isEmpty());
    }
     public ExecuteResult testInvokeContract(List<String> args) {
            List<String> subArgs = new LinkedList<>();
            subArgs.add(args.get(0));
            String contractAddr = args.get(0);
            return invokeContract("global", contractAddr, "openTest", subArgs);
        }

    /**
     * test the post event mechanism
     * @param args
     * @return
     */
    public ExecuteResult testPostEvent(List<String> args) {
        logger.info(args);
        for (int i = 0; i < 10; i ++) {
            Event event = new Event("event" + i);
            event.addTopic("simulate_bank");
            event.addTopic("test");
            event.put("attr1", "value1");
            event.put("attr2", "value2");
            event.put("attr3", "value3");
            ledger.post(event);
        }
        return result(true);
    }

    public ExecuteResult newAccountTable(List<String> args) {
        logger.info(args);
        RelationDB db = ledger.getDataBase(); //do not new database instance every time
        TableName tn = new TableName(getNamespace(), getCid(), "Account");
        logger.info(tn.toString());
        TableDesc desc = new TableDesc(tn);
        desc.AddColumn(new ColumnDesc("name", DataType.STRING));
        desc.AddColumn(new ColumnDesc("id", DataType.LONG));
        desc.AddColumn(new ColumnDesc("age", DataType.INT));
        desc.AddColumn(new ColumnDesc("balance", DataType.DOUBLE));
        db.CreateTable(desc);
        return result(true);
    }

    public ExecuteResult getTableDesc(List<String> args) {
        logger.info(args);
        TableName tn = new TableName(getNamespace(), getCid(), args.get(0));
        RelationDB db = ledger.getDataBase();
        TableDesc desc = db.getTableDesc(tn);
        return result(true, desc);
    }
}