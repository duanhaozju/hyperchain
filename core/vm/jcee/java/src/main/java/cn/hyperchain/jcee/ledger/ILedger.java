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

    byte[] get(byte[] key);
    boolean getBoolean(byte[] key);
    short getShort(byte[] key);
    char getChar(byte[] key);
    int getInt(byte[] key);
    float getFloat(byte[] key);
    double getDouble(byte[] key);
    String getString(byte[] key);
    <T> T getObject(byte[] key, Class<T> clazz);

    byte[] get(String key);
    boolean getBoolean(String key);
    short getShort(String key);
    char getChar(String key);
    int getInt(String key);
    float getFloat(String key);
    double getDouble(String key);
    String getString(String key);
    <T> T getObject(String key, Class<T> clazz);

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