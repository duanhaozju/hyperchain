/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger;

import cn.hyperchain.protos.Key;
import com.google.protobuf.ByteString;
import org.apache.log4j.Logger;

/**
 * HyperchainLedger is an implementation of AbstractLedger
 * which store manipulate the data using remote hyperchain server.
 */
public class HyperchainLedger extends AbstractLedger{

    private static final Logger logger = Logger.getLogger(HyperchainLedger.class.getSimpleName());
    private LedgerClient client;
    public HyperchainLedger(){
        client = new LedgerClient("localhost", 50052);
    }

    public byte[] get(byte[] key) {
        Key sendkey = Key.newBuilder()
                .setId(getContext().getId())
                .setIdBytes(ByteString.copyFrom(key))
                .build();
        logger.info("Transaction id: " + getContext().getId());
        return client.get(sendkey).toByteArray();
    }

    public boolean put(byte[] key, byte[] value) {
        return false;
    }
}
