/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.ledger;

import cn.hyperchain.protos.ContractProto;
import cn.hyperchain.protos.LedgerGrpc;
import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;
import lombok.Getter;
import lombok.Setter;
import org.apache.log4j.Logger;

import java.util.Iterator;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;

public class LedgerClient {

    @Setter
    @Getter
    private String host;

    @Getter
    @Setter
    private int port;

    private static final Logger logger = Logger.getLogger(LedgerClient.class);
    private final ManagedChannel channel;
    private final LedgerGrpc.LedgerBlockingStub blockingStub;
    private final LedgerGrpc.LedgerStub asyncStub;
    private final BlockingQueue<ContractProto.Message> messages;
    private StreamObserver<ContractProto.Message> streamObserver;

    public LedgerClient(String host, int port) {
        this.host = host;
        this.port = port;
        this.channel = ManagedChannelBuilder.forAddress(host, port)
                .usePlaintext(true)
                .build();
        this.blockingStub = LedgerGrpc.newBlockingStub(channel);
        this.asyncStub = LedgerGrpc.newStub(channel);
        this.streamObserver = register();
        this.messages = new LinkedBlockingQueue<>();
        this.streamObserver = register();
    }

    public void shutdown(){
        channel.shutdownNow();
    }

    public boolean put(ContractProto.KeyValue kv) {
        if (streamObserver == null) {
            reconnect();
        }

        ContractProto.Message msg = ContractProto.Message.newBuilder()
                .setType(ContractProto.Message.Type.PUT)
                .setPayload(kv.toByteString())
                .build();
        streamObserver.onNext(msg);

        try {
            ContractProto.Message rsp = messages.take();
            ContractProto.Response response = ContractProto.Response.parseFrom(rsp.getPayload());
            return response.getOk();
        } catch (InterruptedException e) {
            logger.error(e);
            return false;
        }catch (InvalidProtocolBufferException ipbe) {
            logger.error(ipbe);
            return false;
        }
    }

    public ContractProto.Value get(ContractProto.Key key) {

        if (streamObserver == null) {
            reconnect();
        }

        ContractProto.Message msg = ContractProto.Message.newBuilder()
                .setType(ContractProto.Message.Type.GET)
                .setPayload(key.toByteString())
                .build();
        streamObserver.onNext(msg);

        try {
            ContractProto.Message rsp = messages.take();

            if (rsp.getPayload().isEmpty()) {
                return ContractProto.Value.newBuilder().build();
            }else {
                return ContractProto.Value.parseFrom(rsp.getPayload());
            }
        } catch (InterruptedException e) {
            logger.error(e);
            return ContractProto.Value.newBuilder().build();
        }catch (InvalidProtocolBufferException ipbe) {
            logger.error(ipbe);
            return ContractProto.Value.newBuilder().build();
        }
    }

    public boolean delete(ContractProto.Key key) {

        if(streamObserver == null) {
            reconnect();
        }

        ContractProto.Message msg = ContractProto.Message.newBuilder()
                .setType(ContractProto.Message.Type.DELETE)
                .setPayload(key.toByteString())
                .build();
        streamObserver.onNext(msg);

        try {
            ContractProto.Message rsp = messages.take();
            ContractProto.Response response = ContractProto.Response.parseFrom(rsp.getPayload());
            return response.getOk();
        } catch (InterruptedException e) {
            logger.error(e);
            return false;
        }catch (InvalidProtocolBufferException ipbe) {
            logger.error(ipbe);
            return false;
        }
    }

    public boolean batchWrite(ContractProto.BatchKV bkv) {
        if (streamObserver == null) {
            reconnect();
        }

        ContractProto.Message msg = ContractProto.Message.newBuilder()
                .setType(ContractProto.Message.Type.BATCH_WRITE)
                .setPayload(bkv.toByteString())
                .build();
        streamObserver.onNext(msg);

        try {
            ContractProto.Message rsp = messages.take();
            ContractProto.Response response = ContractProto.Response.parseFrom(rsp.getPayload());
            return response.getOk();
        } catch (InterruptedException e) {
            logger.error(e);
            return false;
        }catch (InvalidProtocolBufferException ipbe) {
            logger.error(ipbe);
            return false;
        }
    }

    public ContractProto.BathValue bathRead(ContractProto.BatchKey bk) {
        if (streamObserver == null) {
            reconnect();
        }

        ContractProto.Message msg = ContractProto.Message.newBuilder()
                .setType(ContractProto.Message.Type.BATCH_READ)
                .setPayload(bk.toByteString())
                .build();
        streamObserver.onNext(msg);

        try {
            ContractProto.Message rsp = messages.take();

            if (rsp.getPayload().isEmpty()) {
                return ContractProto.BathValue.newBuilder().build();
            }else {
                return ContractProto.BathValue.parseFrom(rsp.getPayload());
            }
        } catch (InterruptedException e) {
            logger.error(e);
            return ContractProto.BathValue.newBuilder().build();
        }catch (InvalidProtocolBufferException ipbe) {
            logger.error(ipbe);
            return ContractProto.BathValue.newBuilder().build();
        }
    }

    public Iterator<ContractProto.BathValue> rangeQuery(ContractProto.Range range) {
        return blockingStub.rangeQuery(range);
    }

    public boolean post(ContractProto.Event event) {

        if(streamObserver == null) {
            reconnect();
        }

        ContractProto.Message msg = ContractProto.Message.newBuilder()
                .setType(ContractProto.Message.Type.POST_EVENT)
                .setPayload(event.toByteString())
                .build();
        streamObserver.onNext(msg);

        try {
            ContractProto.Message rsp = messages.take();
            ContractProto.Response response = ContractProto.Response.parseFrom(rsp.getPayload());
            return response.getOk();
        } catch (InterruptedException e) {
            logger.error(e);
            return false;
        }catch (InvalidProtocolBufferException ipbe) {
            logger.error(ipbe);
            return false;
        }
    }


    //register create a new stream with ledger server.
    public StreamObserver<ContractProto.Message> register() {

        return asyncStub.register(new StreamObserver<ContractProto.Message>() {
            @Override
            public void onNext(ContractProto.Message message) {
                try {
                    messages.put(message);
                } catch (InterruptedException e) {
                    logger.error(e);
                }
            }

            @Override
            public void onError(Throwable throwable) {

            }

            @Override
            public void onCompleted() {

            }
        });
    }

    public void reconnect() {
        this.streamObserver = register();
    }
}
