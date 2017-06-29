package cn.hyperchain.jcee.ledger.table;

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

    public KvBasedRelationDB() {
        tableMap = new ConcurrentHashMap<>();
    }

    @Override
    public boolean CreateTable(TableDesc tableDesc) {
        String name = tableDesc.getName();
        if (tableMap.containsKey(name)) {
            LOG.error("table " + name + "is existed");
            return false;
        }
        tableMap.put(name, new KvBasedTable(tableDesc));
        return true;
    }

    @Override
    public Table getTable(String name) {
        return tableMap.get(name);
    }

    @Override
    public boolean deleteTable(String name) {
        //TODO: impl it
        //1.delete form local cache
        //2.delete from remote leveldb
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
}
