/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger.table;

import cn.hyperchain.jcee.util.DataType;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Created by wangxiaoyi on 2017/6/23.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
public class ColumnDesc {

    private String name;
    private DataType type;
}
