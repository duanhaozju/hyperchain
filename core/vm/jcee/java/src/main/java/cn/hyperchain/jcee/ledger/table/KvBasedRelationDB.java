/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger.table;

import cn.hyperchain.jcee.ledger.AbstractLedger;
import org.apache.log4j.Logger;

import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Created by wangxiaoyi on 2017/6/29.
 */
public class KvBasedRelationDB implements RelationDB {

    private static final Logger LOG = Logger.getLogger(KvBasedRelationDB.class);
    private Map<String, Table> tableMap;
    private AbstractLedger ledger;

    public KvBasedRelationDB(AbstractLedger ledger) {
        this.ledger = ledger;
        tableMap = new ConcurrentHashMap<>();
    }

    @Override
    public boolean CreateTable(TableDesc tableDesc) {
        String compositeName = tableDesc.getTableName().getCompositeName();
        if (tableMap.containsKey(compositeName)) {
            LOG.error("table " + compositeName + "is existed");
            return false;
        }
        tableMap.put(compositeName, new KvBasedTable(tableDesc, ledger));
        LOG.info("compositeName: " + compositeName);
        LOG.info("tableDesc: " + tableDesc.toJSON());
        ledger.put(compositeName, tableDesc);
        LOG.info(tableMap);
        return true;
    }

    @Override
    public Table getTable(TableName name) {
        return tableMap.get(name.getCompositeName());
    }

    @Override
    public boolean deleteTable(TableName name) {
        String compositeName = name.getCompositeName();
        if (tableMap.containsKey(compositeName)) {
            tableMap.remove(compositeName);
            ledger.delete(compositeName);
        }
        return false;
    }

    @Override
    public List<String> listTables() {
        List<String> tables = new LinkedList<>();
        for (String name: tableMap.keySet()) {
            tables.add(name);
        }
        return tables;
    }

    @Override
    public TableDesc getTableDesc(TableName name) {
        return tableMap.get(name.getCompositeName()).getDesc();
    }
}
