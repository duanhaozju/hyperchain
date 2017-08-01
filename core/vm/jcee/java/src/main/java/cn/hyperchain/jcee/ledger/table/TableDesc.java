/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger.table;

import com.google.gson.Gson;

import java.util.HashSet;
import java.util.Set;
import java.util.TreeSet;

/**
 * Created by wangxiaoyi on 2017/6/29.
 * TableDesc is the description of table schema
 */
public class TableDesc{

    private TableName tableName;
    private Set<ColumnDesc> columnDescSet;

    public TableDesc() {
        tableName = new TableName("", "", "");
        columnDescSet = new TreeSet<>();
    }

    public TableDesc(TableName name) {
        this();
        this.tableName = name;
    }

    public void AddColumn(ColumnDesc columnDesc) {
        columnDescSet.add(columnDesc);
    }

    public String toJSON() {
        Gson gson = new Gson();
        return gson.toJson(this);
    }

    public TableName getTableName() {
        return tableName;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof TableDesc)) return false;

        TableDesc tableDesc = (TableDesc) o;

        if (tableName != null ? !tableName.equals(tableDesc.tableName) : tableDesc.tableName != null) return false;
        return columnDescSet != null ? columnDescSet.equals(tableDesc.columnDescSet) : tableDesc.columnDescSet == null;
    }

    @Override
    public int hashCode() {
        int result = tableName != null ? tableName.hashCode() : 0;
        result = 31 * result + (columnDescSet != null ? columnDescSet.hashCode() : 0);
        return result;
    }

    @Override
    public String toString() {
        return "TableDesc{" +
                "tableName=" + tableName +
                ", columnDescSet=" + columnDescSet +
                '}';
    }
}
