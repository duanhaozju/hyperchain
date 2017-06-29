/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger.table;

import com.google.gson.Gson;

import java.util.HashSet;
import java.util.Set;

/**
 * Created by wangxiaoyi on 2017/6/29.
 * TableDesc is the description of table schema
 */
public class TableDesc {

    private String name;

    private Set<ColumnDesc> columnDescSet;

    public TableDesc() {
        name = "defaultName";
        columnDescSet = new HashSet<>();
    }

    public TableDesc(String name) {
        this();
        this.name = name;
    }

    public void AddColumn(ColumnDesc columnDesc) {
        columnDescSet.add(columnDesc);
    }

    @Override
    public String toString() {
        return "TableDesc{" +
                "name='" + name + '\'' +
                ", columnDescSet=" + columnDescSet +
                '}';
    }

    public String toJSON() {
        Gson gson = new Gson();
        return gson.toJson(this);
    }

    public String getName() {
        return name;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof TableDesc)) return false;

        TableDesc tableDesc = (TableDesc) o;

        if (name != null ? !name.equals(tableDesc.name) : tableDesc.name != null) return false;
        return columnDescSet != null ? columnDescSet.equals(tableDesc.columnDescSet) : tableDesc.columnDescSet == null;
    }

    @Override
    public int hashCode() {
        int result = name != null ? name.hashCode() : 0;
        result = 31 * result + (columnDescSet != null ? columnDescSet.hashCode() : 0);
        return result;
    }
}
