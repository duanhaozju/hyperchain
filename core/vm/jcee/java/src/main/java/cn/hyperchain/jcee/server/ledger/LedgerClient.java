/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.ledger;

import cn.hyperchain.protos.ContractProto;
import cn.hyperchain.protos.LedgerGrpc;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import lombok.Getter;
import lombok.Setter;

import java.util.Iterator;

public class LedgerClient {

    @Setter
    @Getter
    private String host;

    @Getter
    @Setter
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

    public void shutdown(){
        channel.shutdownNow();
    }

    public boolean put(ContractProto.KeyValue kv) {
        ContractProto.Response response = blockingStub.put(kv);
        return response.getOk();
    }

    public ContractProto.Value get(ContractProto.Key key) {
        return blockingStub.get(key);
    }

    public boolean delete(ContractProto.Key key) {
        return blockingStub.delete(key).getOk();
    }

    public boolean batchWrite(ContractProto.BatchKV bkv) {
        return blockingStub.batchWrite(bkv).getOk();
    }

    public ContractProto.BathValue bathRead(ContractProto.BatchKey bk) {
        return blockingStub.batchRead(bk);
    }

    public Iterator<ContractProto.BathValue> rangeQuery(ContractProto.Range range) {
        return blockingStub.rangeQuery(range);
    }

    public boolean post(ContractProto.Event event) {
        return blockingStub.post(event).getOk();
    }
}
