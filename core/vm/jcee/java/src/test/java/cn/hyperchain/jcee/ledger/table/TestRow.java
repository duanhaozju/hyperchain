package cn.hyperchain.jcee.ledger.table;

import com.google.gson.Gson;
import org.apache.log4j.Logger;
import org.junit.Assert;
import org.junit.Test;

/**
 * Created by wangxiaoyi on 2017/7/4.
 *
 */
public class TestRow {

    private static final Logger LOG = Logger.getLogger(TestRow.class);

    @Test
    public void testToGSON() {
        Row row = new Row("00001");
        row.put("k1", "v1".getBytes());
        row.put("k2", "v2".getBytes());
        Gson gson = new Gson();
        Assert.assertEquals(row, gson.fromJson(row.toJSON(), Row.class));
        LOG.info(row.toJSON());
    }
}
