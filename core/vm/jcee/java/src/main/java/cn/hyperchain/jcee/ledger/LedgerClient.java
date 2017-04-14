/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger;

import cn.hyperchain.protos.Key;
import cn.hyperchain.protos.KeyValue;
import cn.hyperchain.protos.LedgerGrpc;
import cn.hyperchain.protos.Value;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;

public class LedgerClient {
    private String host;
    private int port;

    private final ManagedChannel channel;
    private final LedgerGrpc.LedgerBlockingStub blockingStub;

    public LedgerClient(String host, int port) {
        this.host = host;
        this.port = port;
        this.channel = ManagedChannelBuilder.forAddress(host, port)
                .usePlaintext(true)
                .build();
        this.blockingStub = LedgerGrpc.newBlockingStub(channel);
    }

    public String getHost() {
        return host;
    }

    public void setHost(String host) {
        this.host = host;
    }

    public int getPort() {
        return port;
    }

    public void setPort(int port) {
        this.port = port;
    }

    public void shutdown(){
        channel.shutdownNow();
    }

    public boolean put(KeyValue kv) {
        blockingStub.put(kv);
        return true;
    }

    public Value get(Key key) {
        return blockingStub.get(key);
    }
}
