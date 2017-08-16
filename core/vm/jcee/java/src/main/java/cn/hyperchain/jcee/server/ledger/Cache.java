/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.ledger;

/**
 * Created by huhu on 2017/5/13.
 */
public interface Cache {
     void put(byte[]key, byte[]value);
     byte[] get(byte[] key);
     void delete(byte[] key);
     int size();
}
