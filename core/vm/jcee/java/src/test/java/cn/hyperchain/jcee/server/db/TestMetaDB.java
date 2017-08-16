package cn.hyperchain.jcee.server.db;

import cn.hyperchain.jcee.server.contract.ContractInfo;
import cn.hyperchain.jcee.server.contract.ContractState;
import org.junit.Assert;
import org.junit.Test;

/**
 * Created by wangxiaoyi on 2017/5/4.
 */
public class TestMetaDB {

    @Test
    public void testStoreAndLoadMeta() {
        ContractInfo info = new ContractInfo("contractA", "cid0000001", "hyperchain");
        info.setContractPath("path/to/contract");
        info.setClassPrefix("cn.hyperchain.jcee.server.contract.examples.sb");
        info.setNamespace("namespace_global");
        info.setState(ContractState.NORMAL);

//        Class[] cls = new Class[] {String.class, Integer.class};
//        info.setArgClasses(cls);

        Object [] objects = new Object[] {Double.valueOf(123.445), "ssss", new A(123)};
        info.setArgs(objects);

//        MetaDB metaDB = new MetaDB(TestMetaDB.class.getResource("/meta/meta.yaml").getPath());

        MetaDB metaDB = MetaDB.getDb();
        metaDB.setPath(TestMetaDB.class.getResource("/meta/meta.yaml").getPath());
        ContractsMeta contractsMeta = new ContractsMeta();
        contractsMeta.addContractInfo(info);
        metaDB.store(contractsMeta);

        ContractsMeta loadMeta = metaDB.load();
        Assert.assertEquals(loadMeta.toString(), contractsMeta.toString());
    }
}
