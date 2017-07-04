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

    boolean Insert(Row row);

    boolean Update(Row row);

    Row getRow(String rowId);

    Iterator<Row> getRows(String start, String end);

    boolean deleteRow(String rowId);
}
