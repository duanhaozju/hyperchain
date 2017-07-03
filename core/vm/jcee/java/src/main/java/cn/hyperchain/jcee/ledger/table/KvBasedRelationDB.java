package cn.hyperchain.jcee.ledger.table;

import cn.hyperchain.jcee.ledger.AbstractLedger;
import org.apache.log4j.Logger;

import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Created by wangxiaoyi on 2017/6/29.
 * A relationDB instance is attached to a specific contract
 */
public class KvBasedRelationDB implements RelationDB {

    private static final Logger LOG = Logger.getLogger(KvBasedRelationDB.class);
    private Map<String, Table> tableMap;
    private AbstractLedger ledger;
    private String namespace;
    private String cid;

    public KvBasedRelationDB(AbstractLedger ledger) {
        this.ledger = ledger;
        tableMap = new ConcurrentHashMap<>();
    }

    @Override
    public boolean CreateTable(TableDesc tableDesc) {
        String name = tableDesc.getSimpleName();
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

    /**
     * construct a composite table name by namespace contractid and simple table name
     * @param tableName
     * @return compositeTableName := namespace_contractid_tableName
     */
    public String getCompositeTableName(String tableName) {
        return namespace + "_" + cid + "_" + tableName;
    }
}
