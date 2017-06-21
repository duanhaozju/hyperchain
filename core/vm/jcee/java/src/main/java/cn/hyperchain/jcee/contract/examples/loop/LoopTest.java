/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract.examples.loop;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;

import java.util.List;

import cn.hyperchain.jcee.ledger.Result;
import org.apache.log4j.Logger;

/**
 * Created by huhu on 2017/5/31.
 */
public class LoopTest extends ContractTemplate {
    private final Logger logger = Logger.getLogger(LoopTest.class.getCanonicalName());

    @Override
    public ExecuteResult invoke(String funcName, List<String> args) {
        switch (funcName){
            case "deadLoop":
                boolean param = Boolean.parseBoolean(args.get(0));
                return deadLoop(param);
            case "getTest":
                return getTest(args);
            default:
                logger.error("no such method found");

        }
        return result(false);
    }
    private ExecuteResult deadLoop(boolean flag){
        int i = 0;
        while (flag){
            i++;
        }
        return result(true);
    }

    private ExecuteResult getTest(List<String> args){
        int a = 2;
        double d = 2.0;
        ledger.put("int",a);
        ledger.put("double",d);

        Result r = ledger.get("int");
        if(!r.isEmpty()){
            logger.info("int value from ledger "+r.toInt());
        }
        r = ledger.get("double");
        if(!r.isEmpty()){
            logger.info("double value from ledger "+r.toDouble());
        }
        return result(true);
    }
}
