/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger.table;

import cn.hyperchain.jcee.ledger.Result;

import java.util.Iterator;
import java.util.List;

/**
 * Created by wangxiaoyi on 2017/6/23.
 */
public interface Table {

    String getName();

    TableDesc getDesc();

    boolean insert(Row row);

    boolean insertRows(List<Row> rows);

    boolean update(Row row);

    Row getRow(String rowId);

    Iterator<Result> getRows(String start, String end);

    boolean deleteRow(String rowId);
}
