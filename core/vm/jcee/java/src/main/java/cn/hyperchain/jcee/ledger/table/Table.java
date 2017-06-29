/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger.table;

import java.util.Iterator;

/**
 * Created by wangxiaoyi on 2017/6/23.
 */
public interface Table {

    String getName();

    TableDesc getDesc();

    Row newRow();

    boolean Insert(String tableName, Row row);

    boolean Update(String tableName, Row row);

    Row getRow(String tableName, ColumnDesc[] key);

    Iterator<Row> getRows(String tableName, ColumnDesc[] key);

    boolean deleteRow(String tableName, ColumnDesc[] key);
}
