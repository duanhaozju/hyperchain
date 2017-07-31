/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger.table;

/**
 * Created by wangxiaoyi on 2017/7/3.
 */
public class TableName {
    private String namespace = "";
    private String cid = "";
    private String name = "";

    private static final String TABLE_PREFIX = "kv_table_";

    public TableName(String namespace, String cid, String name) {
        this.namespace = namespace;
        this.cid = cid;
        this.name = name;
    }

    public String getNamespace() {
        return namespace;
    }

    public String getCid() {
        return cid;
    }

    //getName get the simple name of table.
    public String getName() {
        return name;
    }

    /**
     * construct a global unique name with combined namespace cid name and prefix: kv_table_
     * @return composite table name
     */
    public String getCompositeName() {
        return TABLE_PREFIX + namespace + "_" + cid + "_" + name;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof TableName)) return false;

        TableName tableName = (TableName) o;

        if (getNamespace() != null ? !getNamespace().equals(tableName.getNamespace()) : tableName.getNamespace() != null)
            return false;
        if (getCid() != null ? !getCid().equals(tableName.getCid()) : tableName.getCid() != null) return false;
        return getName() != null ? getName().equals(tableName.getName()) : tableName.getName() == null;
    }

    @Override
    public int hashCode() {
        int result = getNamespace() != null ? getNamespace().hashCode() : 0;
        result = 31 * result + (getCid() != null ? getCid().hashCode() : 0);
        result = 31 * result + (getName() != null ? getName().hashCode() : 0);
        return result;
    }

    @Override
    public String toString() {
        return "TableName{" +
                "namespace='" + namespace + '\'' +
                ", cid='" + cid + '\'' +
                ", name='" + name + '\'' +
                '}';
    }
}
