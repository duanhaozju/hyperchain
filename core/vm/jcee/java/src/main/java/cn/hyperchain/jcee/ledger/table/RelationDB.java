/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger.table;

import java.util.List;

/**
 * Created by wangxiaoyi on 2017/6/23.
 * used by {@link cn.hyperchain.jcee.ledger.ILedger}
 * provide relational db operations
 */
public interface RelationDB {

    boolean CreateTable(TableDesc tableDesc);

    Table getTable(String name);

    boolean deleteTable(String name);

    //list all stored table names
    List<String> listTables();
}
