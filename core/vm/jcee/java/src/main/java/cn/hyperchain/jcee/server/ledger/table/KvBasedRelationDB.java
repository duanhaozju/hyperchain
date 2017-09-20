/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.ledger.table;

import cn.hyperchain.jcee.client.ledger.table.RelationDB;
import cn.hyperchain.jcee.client.ledger.table.Table;
import cn.hyperchain.jcee.client.ledger.AbstractLedger;
import cn.hyperchain.jcee.client.ledger.table.TableDesc;
import cn.hyperchain.jcee.client.ledger.table.TableName;
import cn.hyperchain.jcee.client.ledger.Result;
import com.google.gson.Gson;
import org.apache.log4j.Logger;

import java.util.HashMap;

/**
 * Created by wangxiaoyi on 2017/6/29.
 */
public class KvBasedRelationDB implements RelationDB {

    private static final Logger LOG = Logger.getLogger(KvBasedRelationDB.class);
    private AbstractLedger ledger;

    public KvBasedRelationDB(AbstractLedger ledger) {
        this.ledger = ledger;
    }

    @Override
    public boolean CreateTable(TableDesc tableDesc) {
        String compositeName = tableDesc.getTableName().getCompositeName();
        LOG.info("tableDesc: " + tableDesc.toJSON());
        LOG.info("compositeName: " + compositeName);
        HashMap schemasMap = getSchemasMap();
        schemasMap.put(compositeName, tableDesc.toJSON());
        Gson gson = new Gson();
        return ledger.put("database_schemas", gson.toJson(schemasMap));
    }

    @Override
    public Table getTable(TableName name) {
        String compositeName = name.getCompositeName();
        HashMap<String, String> schemasMap = getSchemasMap();
        Gson gson = new Gson();
        if (!schemasMap.containsKey(compositeName)) {
            return null;
        }
        TableDesc tableDesc = gson.fromJson(schemasMap.get(compositeName), TableDesc.class);
        Table table = new KvBasedTable(tableDesc, ledger);
        return table;
    }

    @Override
    public boolean deleteTable(TableName name) {
        String compositeName = name.getCompositeName();
        HashMap<String, String> schemasMap = getSchemasMap();
        schemasMap.remove(compositeName);
        Gson gson = new Gson();
        return ledger.put("database_schemas", gson.toJson(schemasMap));
    }

    @Override
    public TableDesc getTableDesc(TableName name) {
        String compositeName = name.getCompositeName();
        HashMap<String, String> schemasMap = getSchemasMap();
        if (!schemasMap.containsKey(compositeName)) {
            return null;
        } else {
            Gson gson = new Gson();
            TableDesc tableDesc = gson.fromJson(schemasMap.get(compositeName), TableDesc.class);
            return tableDesc;
        }
    }

    public HashMap<String, String> getSchemasMap() {
        Result rs = ledger.get("database_schemas");
        HashMap<String, String> schemasMap;
        if (rs.isEmpty()) {
            schemasMap = new HashMap<>();
        } else {
            Gson gson = new Gson();
            schemasMap = gson.fromJson(rs.toString(), HashMap.class);
        }
        return schemasMap;
    }
}
