package cn.hyperchain.jcee.ledger.table;

import java.util.Iterator;

/**
 * Created by wangxiaoyi on 2017/6/29.
 */
public class KvBasedTable implements Table {

    private TableDesc desc;

    public KvBasedTable(TableDesc desc) {
        this.desc = desc;
    }

    @Override
    public String getName() {
        return desc.getSimpleName();
    }

    @Override
    public TableDesc getDesc() {
        return desc;
    }

    @Override
    public Row newRow() {
        return new Row();
    }

    @Override
    public boolean Insert(String tableName, Row row) {
        return false;
    }

    @Override
    public boolean Update(String tableName, Row row) {
        return false;
    }

    @Override
    public Row getRow(String tableName, ColumnDesc[] key) {
        return null;
    }

    @Override
    public Iterator<Row> getRows(String tableName, ColumnDesc[] key) {
        return null;
    }

    @Override
    public boolean deleteRow(String tableName, ColumnDesc[] key) {
        return false;
    }
}
