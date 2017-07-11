/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger.table;

import cn.hyperchain.jcee.ledger.AbstractLedger;
import cn.hyperchain.jcee.ledger.BatchValue;
import cn.hyperchain.jcee.ledger.Result;
import com.google.gson.Gson;

import java.util.Iterator;

/**
 * Created by wangxiaoyi on 2017/6/29.
 */
public class KvBasedTable implements Table {

    private TableDesc desc;
    private AbstractLedger ledger;

    public KvBasedTable(TableDesc desc, AbstractLedger ledger) {
        this.desc = desc;
        this.ledger = ledger;
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
    public boolean insert(Row row) {
        return ledger.put(row.getRowId(), row.toJSON());
    }

    @Override
    public boolean update(Row row) {
        Row oldRow = getRow(row.getRowId());
        if (oldRow == null) {
            return ledger.put(row.getRowId(), row.toJSON());
        }else {
            return ledger.put(row.getRowId(), row.merge(oldRow).toJSON());
        }
    }

    @Override
    public Row getRow(String rowId) {
        Result data = ledger.get(rowId);
        if (!data.isEmpty()) {
            Gson gson = new Gson();
            Row row = gson.fromJson(data.getValue().toString(), Row.class);
            return row;
        }else {
            return null;
        }
    }

    @Override
    public Iterator<Result> getRows(String start, String end) {
        //TODO: maybe bugs existed here
        BatchValue bv = ledger.rangeQuery(start.getBytes(), end.getBytes());
        return bv;
    }

    @Override
    public boolean deleteRow(String rowId) {
        return ledger.delete(rowId);
    }
}
