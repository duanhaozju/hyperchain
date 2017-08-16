/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.client.ledger;


/**
 * Created by wangxiaoyi on 2017/5/5.
 */
public interface Batch {
    //Put keyvalue into the batch
    void put(byte[] key, byte[] value);
    void put(byte[] key, Object value);
    void put(String key, byte[] value);
    void put(String key, Object value);

    Result get(byte[] key);
    Result get(String key);

    //Clear data in batch
    void reset();

    //Commit batch data into ledger
    boolean commit();
}
