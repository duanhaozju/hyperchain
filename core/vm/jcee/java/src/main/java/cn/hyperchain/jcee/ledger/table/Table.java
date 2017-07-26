/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger.table;

import java.util.ArrayList;
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

    ArrayList<Row> getRows(String start, String end);

    boolean deleteRow(String rowId);
}
