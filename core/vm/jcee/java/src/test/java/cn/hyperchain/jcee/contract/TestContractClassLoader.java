/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract;

import cn.hyperchain.jcee.contract.examples.c2.ReloadableContract;
import org.apache.log4j.Logger;
import org.junit.Test;

/**
 * Created by wangxiaoyi on 2017/4/13.
 */
public class TestContractClassLoader {
    private static final Logger logger = Logger.getLogger(TestContractClassLoader.class.getSimpleName());

    @Test
    public void testReLoadClass() {

        String contractDir = TestContractClassLoader.class.getResource("/classes").getPath();
        ContractClassLoader loader = new ContractClassLoader(contractDir, "cn.hyperchain.jcee.contract.examples");
        try{
            ContractBase c1 = (ContractBase) loader.loadClass("cn.hyperchain.jcee.contract.examples.c2.ReloadableContract").newInstance();
            logger.info(c1.getClass().getCanonicalName());
            logger.info(c1.toString());

            ContractClassLoader loader2 = new ContractClassLoader(contractDir, "cn.hyperchain.jcee.contract.examples");
            ContractBase c2 = (ContractBase) loader2.loadClass("cn.hyperchain.jcee.contract.examples.c2.ReloadableContract").newInstance();
            logger.info(c2.getClass().getCanonicalName());
            assert c1.getClass() != c2.getClass();
            assert c1.getClass() != ReloadableContract.class;
            logger.info(c2.toString());
        }catch (Exception e) {
            e.printStackTrace();
        }
    }
}
