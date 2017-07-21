/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract.examples.c1;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;

import java.util.List;

/**
 * Created by wangxiaoyi on 2017/4/7.
 * Test open invoke authority.
 */
public class MySmartContract extends ContractTemplate {

    public MySmartContract() {
        ContractFilter cf = new ContractFilter();
        InvokerFilter invf = new InvokerFilter();
        cf.addPermittedContractAddr("bbe2b6412ccf633222374de8958f2acc76cda9c9");
        invf.addPermittedInvokerAddr("0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd");
        addFilter("openTest", cf);
        addFilter("openTest", invf);
    }

    /**
     * invoke smart contract method
     *
     * @param funcName function name user defined in contract
     * @param args     arguments of funcName
     * @return {@link ExecuteResult}
     */
    @Override
    public ExecuteResult invoke(String funcName, List<String> args) {
        //empty impl
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
            case "openTest": {
                return openTest(args);
            }
            default:
                logger.error("no such method found or the method cannot be invoked by other contract");
        }
        return result(false);
    }

    public ExecuteResult openTest(List<String> args) {
        logger.info(args.get(0));
        return result(true, args.get(0));
    }
}
