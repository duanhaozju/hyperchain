/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger;


public interface ILedger {
    //newBatch create a new batch
    Batch newBatch();

    //newBatchKey construct a batch key
    BatchKey newBatchKey();

    //batchRead read data by batch
    Batch batchRead(BatchKey key);

    //rangeQuery query data by range
    BatchValue rangeQuery(byte[] start, byte[] end);

    //delete data
    boolean delete(byte[] key);
    boolean delete(String key);

    Result get(byte[] key);

    Result get(String key);

    boolean put(byte[] key, byte[]value);
    boolean put(byte[] key, boolean value);
    boolean put(byte[] key, short value);
    boolean put(byte[] key, char value);
    boolean put(byte[] key, int value);
    boolean put(byte[] key, float value);
    boolean put(byte[] key, double value);
    boolean put(byte[] key, String value);
    boolean put(byte[] key, Object object);

    boolean put(String key, byte[]value);
    boolean put(String key, boolean value);
    boolean put(String key, short value);
    boolean put(String key, char value);
    boolean put(String key, int value);
    boolean put(String key, float value);
    boolean put(String key, double value);
    boolean put(String key, String value);
    boolean put(String key, Object object);
}