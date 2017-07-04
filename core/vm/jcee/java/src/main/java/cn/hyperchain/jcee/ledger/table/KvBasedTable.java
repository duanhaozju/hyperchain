/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
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
        return desc.getTableName().getName();
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
    public boolean Insert(Row row) {
        return false;
    }

    @Override
    public boolean Update(Row row) {
        return false;
    }

    @Override
    public Row getRow(String rowId) {
        return null;
    }

    @Override
    public Iterator<Row> getRows(String start, String end) {
        return null;
    }

    @Override
    public boolean deleteRow(String rowId) {
        return false;
    }
}
