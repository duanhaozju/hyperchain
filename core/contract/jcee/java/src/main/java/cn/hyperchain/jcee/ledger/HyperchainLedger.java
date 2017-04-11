/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger;

/**
 * HyperchainLedger is an implementation of ILedger
 * which store manipulate the data using remote hyperchain server.
 */
public class HyperchainLedger implements ILedger{

    public byte[] get(byte[] key) {
        return new byte[0];
    }

    public boolean put(byte[] key, byte[] value) {
        return false;
    }
}
