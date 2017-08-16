/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.client.ledger.table;

import cn.hyperchain.jcee.server.ledger.table.TableDesc;
import cn.hyperchain.jcee.server.ledger.table.TableName;

/**
 * Created by wangxiaoyi on 2017/6/23.
 * used by {@link cn.hyperchain.jcee.client.ledger.ILedger}
 * provide relational db operations
 */
public interface RelationDB {

    boolean CreateTable(TableDesc tableDesc);

    Table getTable(TableName name);

    TableDesc getTableDesc(TableName name);

    boolean deleteTable(TableName name);
}
