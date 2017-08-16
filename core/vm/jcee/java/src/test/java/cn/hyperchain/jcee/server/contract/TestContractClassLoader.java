/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.contract;

import cn.hyperchain.jcee.client.contract.ContractTemplate;
import cn.hyperchain.jcee.client.contract.examples.c2.ReloadableContract;
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
        ContractClassLoader loader = new ContractClassLoader(contractDir, "cn.hyperchain.jcee.server.contract.examples");
        try{
            ContractTemplate c1 = (ContractTemplate) loader.loadClass("cn.hyperchain.jcee.server.contract.examples.c2.ReloadableContract").newInstance();
            logger.info(c1.getClass().getCanonicalName());
            logger.info(c1.toString());

            ContractClassLoader loader2 = new ContractClassLoader(contractDir, "cn.hyperchain.jcee.server.contract.examples");
            ContractTemplate c2 = (ContractTemplate) loader2.loadClass("cn.hyperchain.jcee.server.contract.examples.c2.ReloadableContract").newInstance();
            logger.info(c2.getClass().getCanonicalName());
            assert c1.getClass() != c2.getClass();
            assert c1.getClass() != ReloadableContract.class;
            logger.info(c2.toString());
        }catch (Exception e) {
            e.printStackTrace();
        }
    }
}
