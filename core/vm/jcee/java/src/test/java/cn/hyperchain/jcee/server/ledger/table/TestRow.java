/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.ledger.table;

import cn.hyperchain.jcee.client.ledger.table.Row;
import com.google.gson.Gson;
import org.apache.log4j.Logger;
import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;

/**
 * Created by wangxiaoyi on 2017/7/4.
 */
public class TestRow {

    private static final Logger LOG = Logger.getLogger(TestRow.class);

    static Row[] rows = new Row[3];

    @Before
    public void setUp() {
        LOG.info("before invoke");
        Row row1 = new Row("00001");
        row1.put("k1", "v1".getBytes());
        row1.put("k2", "v2".getBytes());
        rows[0] = row1;

        Row row2 = new Row("00001");
        row2.put("k3", "v3".getBytes());
        row2.put("k2", "v2".getBytes());
        rows[1] = row2;

        Row row3 = new Row("00001");
        row3.put("k1", "v1".getBytes());
        row3.put("k2", "v2".getBytes());
        row3.put("k3", "v3".getBytes());
        rows[2] = row3;
    }

    @Test
    public void testToGSON() {
        Gson gson = new Gson();
        Assert.assertEquals(rows[0], gson.fromJson(rows[0].toJSON(), Row.class));
        LOG.info(rows[0].toJSON());
    }

    @Test
    public void testMerge() {
        Row merged = rows[0].merge(rows[1]);
        Assert.assertEquals(rows[2], merged);
    }
}
