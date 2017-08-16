/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.ledger.table;

import cn.hyperchain.jcee.client.ledger.Batch;
import cn.hyperchain.jcee.client.ledger.BatchValue;
import cn.hyperchain.jcee.client.ledger.table.Table;
import cn.hyperchain.jcee.client.ledger.AbstractLedger;
import cn.hyperchain.jcee.server.ledger.Result;
import com.google.gson.Gson;
import org.apache.log4j.Logger;

import java.util.Iterator;
import java.util.List;

/**
 * Created by wangxiaoyi on 2017/6/29.
 */
public class KvBasedTable implements Table {

    private static final Logger logger = Logger.getLogger(KvBasedRelationDB.class);

    private TableDesc desc;
    private AbstractLedger ledger;

    public KvBasedTable(TableDesc desc, AbstractLedger ledger) {
        this.desc = desc;
        this.ledger = ledger;
    }

    public String getCompositeName(String rowId) {
        return desc.getTableName().getCompositeName() + rowId;
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
        logger.fatal("insert row is " + row.toJSON());
        return ledger.put(getCompositeName(row.getRowId()), row.toJSON());
    }

    @Override
    public boolean insertRows(List<Row> rows) {
        Batch batch = ledger.newBatch();
        for (Row row : rows) {
            logger.error(row.toJSON());
            batch.put(getCompositeName(row.getRowId()), row.toJSON().getBytes());
        }
        return batch.commit();
    }

    @Override
    public boolean update(Row row) {
        Row oldRow = getRow(row.getRowId());
        if (oldRow == null) {
            return ledger.put(getCompositeName(row.getRowId()), row.toJSON());
        }else {
            return ledger.put(getCompositeName(row.getRowId()), row.merge(oldRow).toJSON());
        }
    }

    @Override
    public Row getRow(String rowId) {
        logger.error("get rowId is " + rowId);
        Result data = ledger.get(getCompositeName(rowId));
        if (!data.isEmpty()) {
            Gson gson = new Gson();
            Row row = gson.fromJson(data.toString(), Row.class);
            return row;
        } else {
            return null;
        }
    }

    @Override
    public Iterator<Result> getRows(String start, String end) {
        //TODO: maybe bugs existed here
        BatchValue bv = ledger.rangeQuery(getCompositeName(start).getBytes(), getCompositeName(end).getBytes());
        return bv;
    }

    @Override
    public boolean deleteRow(String rowId) {
        return ledger.delete(getCompositeName(rowId));
    }
}
