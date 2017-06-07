/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract.examples.c1;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import com.google.protobuf.ByteString;
import org.apache.log4j.Logger;

import java.util.List;

/**
 * Created by wangxiaoyi on 2017/4/7.
 */
public class MySmartContract extends ContractTemplate {
    private final Logger logger = Logger.getLogger(MySmartContract.class.getSimpleName());

    public ExecuteResult invoke(String funcName, List<String> args) {
        //logger.info("invoke function name: " + funcName);
        switch (funcName) {
            case "test": {
                this.test("invoke method:" + args.get(0));
                //String name = new String(ledger.get("name".getBytes()));
                //System.out.println("get name from ledger: " + name);
                return result(true);
            }
            default:
                logger.error("no such method found");
        }
        return result(false);
    }

    /**
     * openInvoke provide a interface to be invoked by other contracts
     *
     * @param funcName contract name
     * @param args     arguments
     * @return {@link ExecuteResult}
     */
    @Override
    protected ExecuteResult openInvoke(String funcName, List<String> args) {
        switch (funcName) {
            case "test": {
                this.test("invoke method:" + args.get(0));
                //String name = new String(ledger.get("name".getBytes()));
                //System.out.println("get name from ledger: " + name);
                return result(true);
            }
            default:
                logger.error("no such method found");
        }
        return result(false);
    }

    public void test(String name) {
        logger.info(getLedger().getContext().getId());
        //logger.info("invoke in test");
    }
}