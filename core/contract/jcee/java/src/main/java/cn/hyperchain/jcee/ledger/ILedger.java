package cn.hyperchain.jcee.ledger;

/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */

public interface ILedger {
    byte[] get(byte[] key);
    boolean put(byte[] key, byte[]value);
}

