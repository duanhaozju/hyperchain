package cn.hyperchain.jcee.server.ledger.table;

import org.apache.log4j.Logger;
import org.junit.Assert;
import org.junit.Test;

/**
 * Created by wangxiaoyi on 2017/7/3.
 */
public class TestTableName {
    private static final Logger LOG = Logger.getLogger(TestTableName.class.getName());

    @Test
    public void testGetCompositeName() {
        TableName tableName = new TableName("global", "cid", "person");
        Assert.assertEquals("global_cid_person", tableName.getCompositeName());
    }
}
