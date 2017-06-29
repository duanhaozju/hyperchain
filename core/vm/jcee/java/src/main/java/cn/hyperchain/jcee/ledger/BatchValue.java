/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger;

/**
 * Created by wangxiaoyi on 2017/5/8.
 */
public interface BatchValue {
    Result next();
    boolean hasNext();
}
