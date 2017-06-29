package cn.hyperchain.jcee.ledger.table;

import java.util.Iterator;

/**
 * Created by wangxiaoyi on 2017/6/23.
 * used by {@link cn.hyperchain.jcee.ledger.ILedger}
 * provide relational db operations
 */
public interface RelationDB {

    boolean CreateTable(String name, Column[] columns);

    Table getTable(String name);

    boolean deleteTable(String name);

    boolean Insert(String tableName, Row row);

    boolean Update(String tableName, Row row);

    Row getRow(String tableName, Column[] key);

    Iterator<Row> getRows(String tableName, Column[] key);

    boolean deleteRow(String tableName, Column[] key);
}
