/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger;

/**
 * Created by huhu on 2017/5/13.
 */
public interface Cache {
    public void putInCache(byte[]key, byte[]value);
    public byte[] retrieveFromCache(byte[] key);
    public void removeFromCache(byte[] key);
    public int size();
}
