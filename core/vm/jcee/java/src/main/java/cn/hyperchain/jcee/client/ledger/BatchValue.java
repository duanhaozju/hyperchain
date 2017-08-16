/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.client.ledger;

import java.util.Iterator;

/**
 * Created by wangxiaoyi on 2017/5/8.
 */
public interface BatchValue extends Iterator<Result>{
    Result next();
    boolean hasNext();
}
