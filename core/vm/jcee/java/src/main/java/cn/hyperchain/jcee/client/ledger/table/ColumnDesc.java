/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.client.ledger.table;

import cn.hyperchain.jcee.common.DataType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Created by wangxiaoyi on 2017/6/23.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class ColumnDesc implements Comparable<ColumnDesc>{

    private String name;
    private DataType type;

    @Override
    public int compareTo(ColumnDesc o) {
        return this.name.compareTo(o.name);
    }
}
