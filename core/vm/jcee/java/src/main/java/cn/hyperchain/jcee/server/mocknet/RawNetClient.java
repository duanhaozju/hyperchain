package cn.hyperchain.jcee.server.mocknet;

import cn.hyperchain.protos.ContractGrpc;
import cn.hyperchain.protos.ContractProto;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.stub.StreamObserver;

import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;

public class RawNetClient {

    public static void main(String []args) throws Exception {

      long start = System.currentTimeMillis();
      //syncCallTest();
      asyncCallTest();
      long end = System.currentTimeMillis();
      System.out.println(String.format("call time used: %f s", (end - start) / 1000 * 1.0));
    }

    public static void syncCallTest() {
        ManagedChannel channel;

        ContractGrpc.ContractBlockingStub blockingStub;

        ContractGrpc.ContractStub asyncStub;

        channel = ManagedChannelBuilder.forAddress("localhost", 50051)
                .usePlaintext(true)
                .build();
        blockingStub = ContractGrpc.newBlockingStub(channel);
        int totalCount = 100000;

        for (int i = 0; i < totalCount; i ++) {

            ContractProto.Response rsp = blockingStub.heartBeat(ContractProto.Request.newBuilder().setMethod("syncCall").build());
            System.out.println(String.format("receive response %d %s", i, rsp.getOk()));
        }

    }

    public static void asyncCallTest() throws Exception{
        ManagedChannel channel;


        ContractGrpc.ContractStub asyncStub;

        channel = ManagedChannelBuilder.forAddress("localhost", 50051)
                .usePlaintext(true)
                .build();
        asyncStub = ContractGrpc.newStub(channel);

        int totalCount = 100000;

        BlockingQueue<ContractProto.Message> messages = new LinkedBlockingQueue<>();

        StreamObserver<ContractProto.Message> observer = asyncStub.register(new StreamObserver<ContractProto.Message>() {
            @Override
            public void onNext(ContractProto.Message message) {
                messages.add(message);
            }

            @Override
            public void onError(Throwable throwable) {

            }

            @Override
            public void onCompleted() {

            }
        });

        for (int i = 0; i < totalCount; i ++) {

            ContractProto.Message message = ContractProto.Message.newBuilder().setType(ContractProto.Message.Type.TRANSACTION).build();
            observer.onNext(message);
            
            ContractProto.Message msg = messages.take();
            System.out.println(String.format("receive response %d %s", i, msg.toString()));
        }

    }


}
