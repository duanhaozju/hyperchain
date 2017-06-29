package cn.hyperchain.jcee.ledger.table;

import cn.hyperchain.jcee.util.DataType;
import com.google.gson.Gson;
import org.junit.Assert;
import org.junit.Test;

/**
 * Created by wangxiaoyi on 2017/6/29.
 */
public class TestTableDesc {

    @Test
    public void testToGSON() {
        TableDesc tableDesc = new TableDesc("table001");
        for (int i = 0; i < 10; i ++) {
            ColumnDesc columnDesc = new ColumnDesc();
            columnDesc.setName("cl" + i);
            columnDesc.setType(DataType.BOOL);
            tableDesc.AddColumn(columnDesc);
        }
        Gson gson = new Gson();
        TableDesc reflectTable = gson.fromJson(tableDesc.toJSON(), TableDesc.class);

        Assert.assertEquals(true, tableDesc.equals(reflectTable));
    }
}
