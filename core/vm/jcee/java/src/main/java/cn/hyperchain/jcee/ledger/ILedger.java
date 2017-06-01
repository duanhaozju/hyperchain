/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger;

import cn.hyperchain.jcee.common.exception.NotExistException;

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

    byte[] get(byte[] key) throws NotExistException;
    boolean getBoolean(byte[] key) throws NotExistException;
    short getShort(byte[] key) throws NotExistException;
    char getChar(byte[] key) throws NotExistException;
    int getInt(byte[] key) throws NotExistException;
    float getFloat(byte[] key) throws NotExistException;
    double getDouble(byte[] key) throws NotExistException;
    String getString(byte[] key) throws NotExistException;
    <T> T getObject(byte[] key, Class<T> clazz);

    byte[] get(String key) throws NotExistException;
    boolean getBoolean(String key) throws NotExistException;
    short getShort(String key) throws NotExistException;
    char getChar(String key) throws NotExistException;
    int getInt(String key) throws NotExistException;
    float getFloat(String key) throws NotExistException;
    double getDouble(String key) throws NotExistException;
    String getString(String key) throws NotExistException;
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